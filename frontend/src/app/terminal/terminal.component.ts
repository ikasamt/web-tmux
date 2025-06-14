import { Component, OnInit, OnDestroy, ViewChild, ElementRef, AfterViewInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { Terminal } from 'xterm';
import { FitAddon } from 'xterm-addon-fit';
import { TerminalService } from '../services/terminal.service';
import { Subject, takeUntil } from 'rxjs';

@Component({
  selector: 'app-terminal',
  standalone: true,
  imports: [CommonModule],
  templateUrl: './terminal.component.html',
  styleUrl: './terminal.component.scss'
})
export class TerminalComponent implements OnInit, AfterViewInit, OnDestroy {
  @ViewChild('terminal', { static: true }) terminalElement!: ElementRef<HTMLDivElement>;
  
  private terminal!: Terminal;
  private fitAddon!: FitAddon;
  private destroy$ = new Subject<void>();

  constructor(private terminalService: TerminalService) {}

  ngOnInit(): void {
    this.terminal = new Terminal({
      cursorBlink: true,
      fontSize: 14,
      fontFamily: 'Consolas, "Courier New", monospace',
      allowTransparency: false,
      convertEol: false,
      disableStdin: false,
      scrollback: 1000,
      windowsMode: false,
      macOptionIsMeta: true,
      theme: {
        background: '#1e1e1e',
        foreground: '#d4d4d4',
        cursor: '#d4d4d4',
        cursorAccent: '#1e1e1e',
        black: '#000000',
        red: '#cd3131',
        green: '#0dbc79',
        yellow: '#e5e510',
        blue: '#2472c8',
        magenta: '#bc3fbc',
        cyan: '#11a8cd',
        white: '#e5e5e5',
        brightBlack: '#666666',
        brightRed: '#f14c4c',
        brightGreen: '#23d18b',
        brightYellow: '#f5f543',
        brightBlue: '#3b8eea',
        brightMagenta: '#d670d6',
        brightCyan: '#29b8db',
        brightWhite: '#e5e5e5'
      }
    });

    this.fitAddon = new FitAddon();
    this.terminal.loadAddon(this.fitAddon);
  }

  ngAfterViewInit(): void {
    this.terminal.open(this.terminalElement.nativeElement);
    this.fitAddon.fit();

    // Connect to WebSocket
    this.terminalService.connect();

    // Handle incoming messages
    this.terminalService.messages$
      .pipe(takeUntil(this.destroy$))
      .subscribe(data => {
        const decoder = new TextDecoder();
        this.terminal.write(decoder.decode(data));
      });

    // Handle terminal input
    this.terminal.onData(data => {
      const encoder = new TextEncoder();
      this.terminalService.send(encoder.encode(data));
    });

    // Handle terminal resize
    this.terminal.onResize(({ cols, rows }) => {
      this.terminalService.sendResize(rows, cols);
    });

    // Handle window resize
    window.addEventListener('resize', this.handleResize.bind(this));
    
    // Initial resize
    setTimeout(() => {
      this.handleResize();
      const { rows, cols } = this.terminal;
      this.terminalService.sendResize(rows, cols);
    }, 100);
  }

  private handleResize(): void {
    if (this.fitAddon) {
      this.fitAddon.fit();
    }
  }

  ngOnDestroy(): void {
    window.removeEventListener('resize', this.handleResize.bind(this));
    this.destroy$.next();
    this.destroy$.complete();
    this.terminal.dispose();
    this.terminalService.disconnect();
  }
}