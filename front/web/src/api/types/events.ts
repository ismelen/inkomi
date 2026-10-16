// Messages of GET /ws (back/transaction-manager/internal/usecases/forwardevents) and FCM data
// (back/push-service/internal/domain/models/push-message.go).
import { SourceRow, SourceStatus } from './uploads';

/** Step of a source, and also the close of a joined folder ({id, status: done|failed, error?}). */
export interface SourceStatusEvent {
  id: string;
  status: SourceStatus;
  filename?: string;
  size?: number;
  error?: string;
}

/** A join result created or updated; `result` is the full row (with userId and configKey). */
export interface SourceResultEvent {
  id: string;
  status: SourceStatus;
  result: SourceRow;
  error?: string;
}

export type SourceEvent = SourceStatusEvent | SourceResultEvent;

export function isResultEvent(event: SourceEvent): event is SourceResultEvent {
  return 'result' in event && event.result != null;
}

/** FCM data payload; every value is a string. */
export interface SentPushData {
  type: 'sent';
  id: string;
  filename: string;
  toCloud: string;
  parentId?: string;
}
