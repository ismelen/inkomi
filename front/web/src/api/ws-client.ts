import { wsUrl } from '../config';
import { getValidAccessToken, refreshAccessToken } from './http-client';
import { SourceEvent } from './types/events';

type EventListener = (event: SourceEvent) => void;
type OpenListener = () => void;
type StatusListener = (connected: boolean) => void;
type ChannelMessage = { type: 'status'; connected: boolean } | { type: 'ask' };

const minBackoffMs = 1_000;
const maxBackoffMs = 30_000;
// The server keeps one socket per user: a connection displaced by another device closes quickly.
// Only a connection that lived this long resets the backoff, so two clients don't ping-pong.
const stableAfterMs = 30_000;
// Browsers can't send headers on a websocket: the token travels as the subprotocol list.
const authProtocol = 'bearer';

const channel = typeof BroadcastChannel === 'undefined' ? null : new BroadcastChannel('inkomi-ws');

/**
 * The single websocket of the user (the server keeps one per user and sends no snapshot nor pings).
 * Only one tab of the browser (the one holding the "inkomi-ws" lock) opens it; the others learn
 * about the changes through the shared database. It reconnects with backoff.
 */
class WsClient {
  private socket: WebSocket | null = null;
  private running = false;
  private leader = false;
  private connected = false;
  private backoff = minBackoffMs;
  private retryTimer: ReturnType<typeof setTimeout> | null = null;
  private releaseLock: (() => void) | null = null;
  private eventListeners = new Set<EventListener>();
  private openListeners = new Set<OpenListener>();
  private statusListeners = new Set<StatusListener>();

  constructor() {
    channel?.addEventListener('message', (message: MessageEvent<ChannelMessage>) => {
      const data = message.data;
      if (data.type === 'ask' && this.leader) channel.postMessage(this.status());
      if (data.type === 'status' && !this.leader) this.setConnected(data.connected, false);
    });
  }

  start() {
    if (this.running) return;
    this.running = true;
    window.addEventListener('online', this.onOnline);
    channel?.postMessage({ type: 'ask' } satisfies ChannelMessage);

    if (!navigator.locks) {
      this.becomeLeader();
      return;
    }
    void navigator.locks.request(
      'inkomi-ws',
      () =>
        new Promise<void>((resolve) => {
          if (!this.running) return resolve();
          this.releaseLock = resolve;
          this.becomeLeader();
        })
    );
  }

  stop() {
    this.running = false;
    this.leader = false;
    window.removeEventListener('online', this.onOnline);
    this.clearRetry();
    this.close();
    this.releaseLock?.();
    this.releaseLock = null;
    this.setConnected(false, false);
  }

  /** Whether this tab owns the websocket (and so the sync and notifications). */
  isLeader(): boolean {
    return this.leader;
  }

  isConnected(): boolean {
    return this.connected;
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

  /** Connection state of the websocket of any tab. */
  onStatus(listener: StatusListener): () => void {
    this.statusListeners.add(listener);
    return () => this.statusListeners.delete(listener);
  }

  private becomeLeader() {
    this.leader = true;
    void this.connect();
  }

  private status(): ChannelMessage {
    return { type: 'status', connected: this.connected };
  }

  private setConnected(connected: boolean, broadcast: boolean) {
    if (broadcast) channel?.postMessage({ type: 'status', connected } satisfies ChannelMessage);
    if (this.connected === connected) return;
    this.connected = connected;
    this.statusListeners.forEach((listener) => listener(connected));
  }

  private onOnline = () => {
    if (!this.leader || this.socket) return;
    this.backoff = minBackoffMs;
    this.clearRetry();
    void this.connect();
  };

  private async connect(forceRefresh = false) {
    if (!this.leader || this.socket) return;

    let token: string | null = null;
    try {
      token = forceRefresh ? await refreshAccessToken() : await getValidAccessToken();
    } catch {
      token = null;
    }
    if (!this.leader || this.socket) return;
    if (!token) {
      this.scheduleRetry(false);
      return;
    }

    let openedAt = 0;
    const socket = new WebSocket(wsUrl, [authProtocol, token]);
    this.socket = socket;

    socket.onopen = () => {
      openedAt = Date.now();
      this.setConnected(true, true);
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
      this.setConnected(false, true);
      if (openedAt && Date.now() - openedAt > stableAfterMs) this.backoff = minBackoffMs;
      // A handshake rejected before opening is usually an expired token.
      this.scheduleRetry(!openedAt);
    };
    socket.onerror = () => {
      socket.close();
    };
  }

  private scheduleRetry(refresh: boolean) {
    if (!this.leader) return;
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
