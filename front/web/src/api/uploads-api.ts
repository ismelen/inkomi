import { http } from './http-client';
import {
  DownloadResponse,
  SourceDTO,
  UploadListQuery,
  UploadListResponse,
  UploadRequestDTO,
} from './types/uploads';

export const uploadsApi = {
  request: (body: UploadRequestDTO) => http.post<SourceDTO[]>('/uploads/request', body),
  /** Confirms a file (after its PUT) or a library item. Never call it on a folder. */
  done: (sourceId: string) => http.patch<void>(`/uploads/done/${encodeURIComponent(sourceId)}`),
  list: (query: UploadListQuery = {}) =>
    http.get<UploadListResponse>('/uploads', { query: { ...query } }),
  download: (sourceId: string) =>
    http.get<DownloadResponse>(`/uploads/${encodeURIComponent(sourceId)}/download`),
};
