import { Injectable } from '@angular/core';
import { Subject, Observable } from 'rxjs';

export interface TerminalMessage {
  type: 'data' | 'resize';
  data?: ArrayBuffer;
  rows?: number;
  cols?: number;
}

@Injectable({
  providedIn: 'root'
})
export class TerminalService {
  private ws: WebSocket | null = null;
  private messageSubject = new Subject<ArrayBuffer>();
  private connectionSubject = new Subject<boolean>();

  messages$ = this.messageSubject.asObservable();
  connection$ = this.connectionSubject.asObservable();

  connect(url: string = 'ws://localhost:8080/ws'): void {
    if (this.ws) {
      this.disconnect();
    }

    this.ws = new WebSocket(url);
    this.ws.binaryType = 'arraybuffer';

    this.ws.onopen = () => {
      console.log('WebSocket connected');
      this.connectionSubject.next(true);
    };

    this.ws.onmessage = (event) => {
      if (event.data instanceof ArrayBuffer) {
        this.messageSubject.next(event.data);
      } else if (typeof event.data === 'string') {
        // Convert string message to ArrayBuffer for terminal display
        const encoder = new TextEncoder();
        this.messageSubject.next(encoder.encode(event.data).buffer);
      }
    };

    this.ws.onerror = (error) => {
      console.error('WebSocket error:', error);
    };

    this.ws.onclose = () => {
      console.log('WebSocket disconnected');
      this.connectionSubject.next(false);
    };
  }

  send(data: ArrayBuffer | string): void {
    if (this.ws && this.ws.readyState === WebSocket.OPEN) {
      this.ws.send(data);
    }
  }

  sendResize(rows: number, cols: number): void {
    if (this.ws && this.ws.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify({ rows, cols }));
    }
  }

  disconnect(): void {
    if (this.ws) {
      this.ws.close();
      this.ws = null;
    }
  }

  ngOnDestroy(): void {
    this.disconnect();
    this.messageSubject.complete();
    this.connectionSubject.complete();
  }
}