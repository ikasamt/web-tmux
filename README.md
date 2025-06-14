# Web Terminal with tmux Integration

A modern web-based terminal built with Angular and Go, featuring tmux session management for persistent remote sessions.

## Features

- 🌐 **Web-based Terminal**: Access your terminal through any modern browser
- 🔄 **tmux Integration**: Persistent sessions that survive browser refreshes
- 🎨 **Modern UI**: Built with Angular v19 and Tailwind CSS
- ⚡ **Real-time Communication**: WebSocket-based for low latency
- 🌍 **Japanese Support**: Full UTF-8 and Japanese locale support
- 📦 **Single Binary**: Self-contained executable with embedded frontend
- 🔒 **Secure**: Designed for use with Tailscale or similar VPN solutions

## Requirements

- **tmux**: Required for session management
  - macOS: `brew install tmux`
  - Ubuntu: `sudo apt install tmux` 
  - CentOS: `sudo yum install tmux`

## Quick Start

### Option 1: Use Pre-built Binary

```bash
git clone https://github.com/ikasamt/web-tmux.git
cd web-tmux
./web-terminal
```

Then open http://localhost:8080 in your browser.

### Option 2: Build from Source

```bash
git clone https://github.com/ikasamt/web-tmux.git
cd web-tmux
./build.sh
./web-terminal
```

## How It Works

1. **Session Management**: Creates or attaches to a tmux session named "web-terminal"
2. **Persistence**: Browser refreshes won't kill your session
3. **Multi-access**: Multiple browser tabs can connect to the same session
4. **tmux Features**: Full access to tmux's window/pane management

## Usage

### Basic Terminal Operations
- Type commands as you would in a regular terminal
- Use tmux key bindings (default prefix: `Ctrl+b`)
- Session persists across browser refreshes

### tmux Commands
```bash
# Create new window
Ctrl+b c

# Switch between windows  
Ctrl+b 0-9

# Split pane horizontally
Ctrl+b "

# Split pane vertically
Ctrl+b %

# Detach from session (keeps running)
Ctrl+b d
```

### Manual Session Management
```bash
# List active sessions
tmux list-sessions

# Attach to web-terminal session manually
tmux attach-session -t web-terminal

# Kill the session
tmux kill-session -t web-terminal
```

## Architecture

```
┌─────────────────┐    WebSocket    ┌──────────────┐    PTY    ┌─────────┐
│   Angular SPA   │ ◄──────────────► │ Go Backend   │ ◄────────► │  tmux   │
│   (xterm.js)    │                 │ (Gin + WS)   │           │ session │
└─────────────────┘                 └──────────────┘           └─────────┘
```

### Components

- **Frontend**: Angular v19 + xterm.js + Tailwind CSS
- **Backend**: Go + Gin + gorilla/websocket + creack/pty
- **Session**: tmux for persistence and advanced terminal features

## Development

### Prerequisites
- Node.js 18+
- Go 1.19+
- tmux

### Setup
```bash
# Install frontend dependencies
cd frontend
npm install

# Install Go dependencies  
cd ../backend
go mod download

# Development mode (run separately)
# Terminal 1: Frontend dev server
cd frontend && ng serve

# Terminal 2: Backend dev server
cd backend && go run main.go
```

### Build Process

The `build.sh` script:
1. Builds Angular for production
2. Embeds the dist files into Go binary using `embed`
3. Compiles a single self-contained executable

## Configuration

### Environment Variables
- `SHELL`: Shell to use (defaults to `/bin/bash`)
- Standard locale variables for Japanese support:
  - `LANG=ja_JP.UTF-8`
  - `LC_ALL=ja_JP.UTF-8`

### tmux Configuration
The application works with any tmux configuration. For better experience, consider:

```bash
# ~/.tmux.conf
set -g default-terminal "xterm-256color"
set -g mouse on
```

## Security Considerations

This application is designed for use in trusted environments:

- **No built-in authentication**: Intended for use with Tailscale or similar VPN
- **Local access by default**: Binds to localhost:8080
- **Shell access**: Provides full shell access to the host system

For production use:
- Deploy behind a reverse proxy with authentication
- Use with Tailscale for secure remote access
- Consider firewall rules to restrict access

## Troubleshooting

### tmux not found
```
╔══════════════════════════════════════════════════════════════╗
║                         tmux必須です                           ║
║  このWebターミナルを使用するにはtmuxが必要です。                ║
╚══════════════════════════════════════════════════════════════╝
```

Install tmux using your package manager.

### Japanese input not working
Ensure your system has Japanese locale support:
```bash
locale -a | grep ja_JP
```

### Session not persisting
- Check if tmux is running: `tmux list-sessions`
- Verify the web-terminal session exists
- Check server logs for tmux-related errors

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Test thoroughly
5. Submit a pull request

## License

MIT License - see LICENSE file for details.

## Acknowledgments

- [xterm.js](https://xtermjs.org/) - Terminal emulator for the web
- [tmux](https://github.com/tmux/tmux) - Terminal multiplexer
- [creack/pty](https://github.com/creack/pty) - Go PTY interface
- [Angular](https://angular.io/) - Frontend framework
- [Gin](https://gin-gonic.com/) - Go web framework