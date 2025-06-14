package main

import (
	"embed"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/creack/pty"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// Build information (set via ldflags)
var (
	version   = "dev"
	buildTime = "unknown"
	gitCommit = "unknown"
)

//go:embed static/frontend/*
var staticFiles embed.FS

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type TerminalSize struct {
	Rows uint16 `json:"rows"`
	Cols uint16 `json:"cols"`
}

func main() {
	var showVersion = flag.Bool("version", false, "Show version information")
	flag.Parse()

	if *showVersion {
		fmt.Printf("Web Terminal %s\n", version)
		fmt.Printf("Build Time: %s\n", buildTime)
		fmt.Printf("Git Commit: %s\n", gitCommit)
		os.Exit(0)
	}

	log.Printf("Starting Web Terminal %s (commit: %s, built: %s)", version, gitCommit, buildTime)

	r := gin.Default()

	// Serve static files
	staticFS, err := fs.Sub(staticFiles, "static/frontend/browser")
	if err != nil {
		log.Printf("Failed to create sub filesystem from 'static/frontend/browser': %v", err)
		// Try fallback path
		staticFS, err = fs.Sub(staticFiles, "static/frontend")
		if err != nil {
			log.Printf("Failed to create sub filesystem from 'static/frontend': %v", err)
			log.Fatal("Could not locate static files in embed filesystem")
		}
	}

	r.Use(func(c *gin.Context) {
		// CORS headers
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		
		c.Next()
	})

	// WebSocket endpoint
	r.GET("/ws", handleWebSocket)

	// Static file serving with SPA fallback
	r.NoRoute(func(c *gin.Context) {
		path := strings.TrimPrefix(c.Request.URL.Path, "/")
		if path == "" {
			path = "index.html"
		}

		// Check if file exists
		if _, err := staticFS.Open(path); err != nil {
			// File not found, serve index.html for SPA routing
			path = "index.html"
		}

		file, err := staticFS.Open(path)
		if err != nil {
			c.Status(404)
			return
		}
		defer file.Close()

		// Set appropriate content type
		if strings.HasSuffix(path, ".html") {
			c.Header("Content-Type", "text/html; charset=utf-8")
		} else if strings.HasSuffix(path, ".js") {
			c.Header("Content-Type", "application/javascript")
		} else if strings.HasSuffix(path, ".css") {
			c.Header("Content-Type", "text/css")
		} else if strings.HasSuffix(path, ".ico") {
			c.Header("Content-Type", "image/x-icon")
		}

		if seeker, ok := file.(io.ReadSeeker); ok {
			http.ServeContent(c.Writer, c.Request, path, time.Time{}, seeker)
		} else {
			// Fallback for files that don't implement ReadSeeker
			data, err := io.ReadAll(file)
			if err != nil {
				c.Status(500)
				return
			}
			c.Data(200, c.GetHeader("Content-Type"), data)
		}
	})

	// Serve root explicitly
	r.GET("/", func(c *gin.Context) {
		file, err := staticFS.Open("index.html")
		if err != nil {
			c.Status(404)
			return
		}
		defer file.Close()

		c.Header("Content-Type", "text/html; charset=utf-8")
		if seeker, ok := file.(io.ReadSeeker); ok {
			http.ServeContent(c.Writer, c.Request, "index.html", time.Time{}, seeker)
		} else {
			data, err := io.ReadAll(file)
			if err != nil {
				c.Status(500)
				return
			}
			c.Data(200, "text/html; charset=utf-8", data)
		}
	})

	log.Println("Server starting on :8080")
	log.Println("Access the terminal at http://localhost:8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}

func handleWebSocket(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println("Failed to upgrade connection:", err)
		return
	}
	defer conn.Close()

	// Check if tmux is available
	if _, err := exec.LookPath("tmux"); err != nil {
		log.Println("tmux not found")
		// Send error message to client
		errorMsg := "\r\n\r\n" +
			"╔══════════════════════════════════════════════════════════════╗\r\n" +
			"║                         tmux必須です                           ║\r\n" +
			"║                                                              ║\r\n" +
			"║  このWebターミナルを使用するにはtmuxが必要です。                ║\r\n" +
			"║                                                              ║\r\n" +
			"║  インストール方法:                                           ║\r\n" +
			"║    macOS: brew install tmux                                  ║\r\n" +
			"║    Ubuntu: sudo apt install tmux                            ║\r\n" +
			"║    CentOS: sudo yum install tmux                            ║\r\n" +
			"║                                                              ║\r\n" +
			"║  インストール後、サーバーを再起動してください。               ║\r\n" +
			"╚══════════════════════════════════════════════════════════════╝\r\n\r\n"
		
		conn.WriteMessage(websocket.TextMessage, []byte(errorMsg))
		return
	}

	// Start tmux session
	sessionName := "web-terminal"
	var cmd *exec.Cmd
	
	// Check if tmux session exists, create if not
	checkCmd := exec.Command("tmux", "has-session", "-t", sessionName)
	if checkCmd.Run() != nil {
		// Session doesn't exist, create it with proper UTF-8 support
		createCmd := exec.Command("tmux", "new-session", "-d", "-s", sessionName)
		createCmd.Env = append(os.Environ(), 
			"LANG=ja_JP.UTF-8",
			"LC_ALL=ja_JP.UTF-8",
			"LC_CTYPE=ja_JP.UTF-8",
		)
		if err := createCmd.Run(); err != nil {
			log.Println("Failed to create tmux session:", err)
			errorMsg := "\r\ntmuxセッションの作成に失敗しました: " + err.Error() + "\r\n"
			conn.WriteMessage(websocket.TextMessage, []byte(errorMsg))
			return
		}
		cmd = exec.Command("tmux", "attach-session", "-t", sessionName)
	} else {
		// Session exists, attach to it
		cmd = exec.Command("tmux", "attach-session", "-t", sessionName)
	}
	cmd.Env = append(os.Environ(), 
		"TERM=xterm-256color",
		"COLORTERM=truecolor",
		"LANG=ja_JP.UTF-8",
		"LC_ALL=ja_JP.UTF-8",
		"LC_CTYPE=ja_JP.UTF-8",
		"PS1=$ ",
		"COLUMNS=80",
		"LINES=24",
	)

	ptmx, err := pty.StartWithSize(cmd, &pty.Winsize{
		Rows: 24,
		Cols: 80,
	})
	if err != nil {
		log.Println("Failed to start pty:", err)
		return
	}
	defer func() {
		log.Println("Cleaning up process...")
		ptmx.Close()
		if cmd.Process != nil {
			cmd.Process.Kill()
		}
	}()

	// Channel to signal cleanup
	done := make(chan struct{})

	// Read from terminal and send to websocket
	go func() {
		defer close(done)
		buf := make([]byte, 1024)
		for {
			n, err := ptmx.Read(buf)
			if err != nil {
				log.Println("Failed to read from pty:", err)
				conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := conn.WriteMessage(websocket.BinaryMessage, buf[:n]); err != nil {
				log.Println("Failed to write to websocket:", err)
				return
			}
		}
	}()

	// Read from websocket and send to terminal
	go func() {
		for {
			messageType, data, err := conn.ReadMessage()
			if err != nil {
				log.Println("Failed to read from websocket:", err)
				return
			}

			switch messageType {
			case websocket.TextMessage:
				// Handle resize messages
				var size TerminalSize
				if err := json.Unmarshal(data, &size); err == nil {
					if err := pty.Setsize(ptmx, &pty.Winsize{
						Rows: size.Rows,
						Cols: size.Cols,
					}); err != nil {
						log.Println("Failed to set size:", err)
					}
				}
			case websocket.BinaryMessage:
				// Handle terminal input
				if _, err := ptmx.Write(data); err != nil {
					log.Println("Failed to write to pty:", err)
					return
				}
			}
		}
	}()

	// Wait for cleanup signal
	<-done
}