import { ConfigDTO } from '../api/types/manga-config';

/** ConfigDTO without the e-reader, which comes from the settings when sending. */
export type MangaConfig = Omit<ConfigDTO, 'ereaderKey'>;

export const cropDisabled = -1;

// Same defaults as ConfigDTO.UnmarshalJSON in the transaction-manager.
export const defaultMangaConfig: MangaConfig = {
  rotateDoublePage: false,
  cropTolerance: cropDisabled,
  enlarge: false,
  embed: false,
  gamma: 1,
  brightness: 0,
  contrast: 1,
  grayscale: false,
  blackPointThreshold: 0.06,
  whitePointThreshold: 0.95,
  dithering: false,
  palette8bit: false,
  quality: 90,
  sharpen: { enabled: false, radius: 1, x1: 1.5, y2: 20, y3: 50, m1: 0, m2: 3 },
};

export interface Range {
  min: number;
  max: number;
  step: number;
}

// Ranges validated by the manga-worker (manga-config.go). m1 has no upper bound there.
export const mangaRanges = {
  gamma: { min: 0.1, max: 3, step: 0.05 },
  brightness: { min: -1, max: 1, step: 0.05 },
  contrast: { min: 0, max: 2, step: 0.05 },
  blackPointThreshold: { min: 0, max: 1, step: 0.01 },
  whitePointThreshold: { min: 0, max: 1, step: 0.01 },
  cropTolerance: { min: 0, max: 1, step: 0.01 },
  quality: { min: 1, max: 100, step: 1 },
  radius: { min: 1, max: 5, step: 1 },
  x1: { min: 1, max: 3, step: 0.1 },
  y2: { min: 0, max: 50, step: 1 },
  y3: { min: 0, max: 50, step: 1 },
  m1: { min: 0, max: 10, step: 0.1 },
  m2: { min: 1, max: 5, step: 0.1 },
} satisfies Record<string, Range>;

export type MangaConfigError = 'blackAboveWhite' | 'outOfRange';

export function validateMangaConfig(c: MangaConfig): MangaConfigError | null {
  const r = mangaRanges;
  const inRange = (v: number, range: Range) => v >= range.min && v <= range.max;
  const checks = [
    inRange(c.gamma, r.gamma),
    inRange(c.brightness, r.brightness),
    inRange(c.contrast, r.contrast),
    inRange(c.blackPointThreshold, r.blackPointThreshold),
    inRange(c.whitePointThreshold, r.whitePointThreshold),
    inRange(c.quality, r.quality),
    c.cropTolerance === cropDisabled || inRange(c.cropTolerance, r.cropTolerance),
  ];
  if (c.sharpen.enabled) {
    checks.push(
      inRange(c.sharpen.radius, r.radius),
      inRange(c.sharpen.x1, r.x1),
      inRange(c.sharpen.y2, r.y2),
      inRange(c.sharpen.y3, r.y3),
      c.sharpen.m1 >= 0,
      inRange(c.sharpen.m2, r.m2)
    );
  }
  if (checks.some((ok) => !ok)) return 'outOfRange';
  if (c.blackPointThreshold >= c.whitePointThreshold) return 'blackAboveWhite';
  return null;
}

export function toConfigDTO(config: MangaConfig, ereaderKey: string): ConfigDTO {
  return { ...config, sharpen: { ...config.sharpen }, ereaderKey };
}
