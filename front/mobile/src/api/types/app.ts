// Mirror of back/release-service/internal/infra/api/dtos/release-response.go.

export interface ReleaseResponse {
  /** MAJOR.MINOR.PATCH, compared with expo.version. */
  version: string;
  notes: string;
  publishedAt: string;
  fileName: string;
  size: number;
  sha256?: string;
  /** Relative to the backend url; redirects to the apk of GitHub Releases. */
  downloadPath: string;
}
