// Mirror of back/transaction-manager/internal/infra/api/dtos.
// The upload request is snake_case; the rest of the service is camelCase.
import { ConfigDTO } from './manga-config';

export type SourceStatus =
  | 'pending_upload'
  | 'queued'
  | 'processing'
  | 'joining'
  | 'joined'
  | 'kepubifying'
  | 'done'
  | 'sent'
  | 'failed';

export type SourceType = 'file' | 'folder' | 'library';
export type ReadingDirection = 'ltr' | 'rtl';

export interface NoItemsSourceDTO {
  title: string;
  type: SourceType;
  size?: number;
  /** md5 for library sources. */
  filename: string;
  /** Key of UploadRequestDTO.config. */
  config_hash?: string;
  reading_direction: ReadingDirection;
  to_cloud: boolean;
  /** Set by the server. */
  id?: string;
  /** Presigned PUT url, only for type "file". */
  url?: string;
}

export interface SourceDTO extends NoItemsSourceDTO {
  should_join: boolean;
  items: NoItemsSourceDTO[];
}

export interface UploadRequestDTO {
  ereader_key: string;
  sources: SourceDTO[];
  config: Record<string, ConfigDTO>;
}

/** A source row as GET /uploads and the websocket `result` send it. */
export interface SourceRow {
  id: string;
  title: string;
  filename: string;
  kepubify: boolean;
  toCloud: boolean;
  isRtl: boolean;
  join: boolean;
  status: SourceStatus;
  errorMsg: string | null;
  size: number | null;
  isFile: boolean;
  createdAt: string;
  updatedAt: string;
  completedAt: string | null;
  parentId: string | null;
  position: number;
  isResult: boolean;
}

export interface UploadDTO extends SourceRow {
  children: SourceRow[];
}

export interface UploadListResponse {
  uploads: UploadDTO[];
  nextCursor: string | null;
  syncedAt: string;
}

export interface UploadListQuery {
  limit?: number;
  cursor?: string;
  updatedSince?: string;
}

export interface DownloadResponse {
  url: string;
  filename: string;
  size: number | null;
  expiresAt: string;
}
