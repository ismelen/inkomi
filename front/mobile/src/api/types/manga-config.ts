// Mirror of back/transaction-manager/internal/infra/api/dtos/config-dto.go (camelCase).
// Ranges: back/manga-worker/internal/domain/models/manga-config.go.

export interface SharpenDTO {
  enabled: boolean;
  radius: number;
  x1: number;
  y2: number;
  y3: number;
  m1: number;
  m2: number;
}

export interface ConfigDTO {
  ereaderKey: string;
  rotateDoublePage: boolean;
  cropTolerance: number;
  enlarge: boolean;
  embed: boolean;
  gamma: number;
  brightness: number;
  contrast: number;
  grayscale: boolean;
  blackPointThreshold: number;
  whitePointThreshold: number;
  dithering: boolean;
  palette8bit: boolean;
  quality: number;
  sharpen: SharpenDTO;
}
