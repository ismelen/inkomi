import { uploadsApi } from '../api/uploads-api';

/**
 * Asks for a fresh presigned link (never cached: it expires after 15 min) and lets the browser
 * download it. Errors (409 not ready, 404 purged) are thrown to be shown, never retried.
 */
export async function downloadSource(sourceId: string): Promise<void> {
  const link = await uploadsApi.download(sourceId);
  const a = document.createElement('a');
  a.href = link.url;
  a.download = link.filename;
  a.rel = 'noopener';
  document.body.appendChild(a);
  a.click();
  a.remove();
}
