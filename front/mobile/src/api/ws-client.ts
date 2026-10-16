import { AppState, AppStateStatus, NativeEventSubscription } from 'react-native';
import { wsUrl } from '../config';
import { getValidAccessToken, refreshAccessToken } from './http-client';
import { SourceEvent } from './types/events';

// React Native's WebSocket accepts headers as a third argument, unlike the DOM one.
type NativeWebSocket = new (
  url: string,
  protocols?: string | string[],
  options?: { headers: Record<string, string> }
) => WebSocket;

type EventListener = (event: SourceEvent) => void;
type OpenListener = () => void;

const minBackoffMs = 1_000;
const maxBackoffMs = 30_000;

/**
 * The single websocket of the user (the server keeps one per user and sends no snapshot nor pings).
 * It reconnects with backoff, closes in background and reopens when the app is active again.
 */
class WsClient {
  private socket: WebSocket | null = null;
  private running = false;
  private backoff = minBackoffMs;
  private retryTimer: ReturnType<typeof setTimeout> | null = null;
  private appState: NativeEventSubscription | null = null;
  private eventListeners = new Set<EventListener>();
  private openListeners = new Set<OpenListener>();

  start() {
    if (this.running) return;
    this.running = true;
    this.appState = AppState.addEventListener('change', this.onAppState);
    void this.connect();
  }

  stop() {
    this.running = false;
    this.appState?.remove();
    this.appState = null;
    this.clearRetry();
    this.close();
  }

  onEvent(listener: EventListener): () => void {
    this.eventListeners.add(listener);
    return () => this.eventListeners.delete(listener);
  }

  /** Called on every (re)connection, the moment to reconcile what was missed. */
  onOpen(listener: OpenListener): () => void {
    this.openListeners.add(listener);
    return () => this.openListeners.delete(listener);
  }

  private onAppState = (state: AppStateStatus) => {
    if (!this.running) return;
    if (state === 'active') {
      this.backoff = minBackoffMs;
      if (!this.socket) {
        this.clearRetry();
        void this.connect();
      }
    } else if (state === 'background') {
      this.clearRetry();
      this.close();
    }
  };

  private async connect(forceRefresh = false) {
    if (!this.running || this.socket) return;

    let token: string | null = null;
    try {
      token = forceRefresh ? await refreshAccessToken() : await getValidAccessToken();
    } catch {
      token = null;
    }
    if (!this.running || this.socket) return;
    if (!token) {
      this.scheduleRetry(false);
      return;
    }

    let opened = false;
    const socket = new (WebSocket as unknown as NativeWebSocket)(wsUrl, undefined, {
      headers: { Authorization: `Bearer ${token}` },
    });
    this.socket = socket;

    socket.onopen = () => {
      opened = true;
      this.backoff = minBackoffMs;
      this.openListeners.forEach((listener) => listener());
    };
    socket.onmessage = (message) => {
      try {
        const event = JSON.parse(String(message.data)) as SourceEvent;
        if (event?.id && event.status) this.eventListeners.forEach((listener) => listener(event));
      } catch {
        // not a source event
      }
    };
    socket.onclose = () => {
      if (this.socket !== socket) return;
      this.socket = null;
      // A handshake rejected before opening is usually an expired token.
      this.scheduleRetry(!opened);
    };
    socket.onerror = () => {
      socket.close();
    };
  }

  private scheduleRetry(refresh: boolean) {
    if (!this.running || AppState.currentState === 'background') return;
    this.clearRetry();
    const delay = this.backoff;
    this.backoff = Math.min(this.backoff * 2, maxBackoffMs);
    this.retryTimer = setTimeout(() => {
      this.retryTimer = null;
      void this.connect(refresh);
    }, delay);
  }

  private clearRetry() {
    if (this.retryTimer) clearTimeout(this.retryTimer);
    this.retryTimer = null;
  }

  private close() {
    const socket = this.socket;
    this.socket = null;
    socket?.close();
  }
}

export const wsClient = new WsClient();
