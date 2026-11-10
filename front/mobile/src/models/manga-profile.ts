import { defaultMangaConfig, MangaConfig } from './manga-config';

/** A named manga config the user saved to reuse. */
export interface MangaProfile {
  id: string;
  name: string;
  config: MangaConfig;
}

export function newProfileId(): string {
  return `${Date.now().toString(36)}-${Math.random().toString(36).slice(2, 8)}`;
}

/** The config with every field, so profiles saved before a new option still compare. */
function complete(config: MangaConfig): MangaConfig {
  return {
    ...defaultMangaConfig,
    ...config,
    sharpen: { ...defaultMangaConfig.sharpen, ...config.sharpen },
  };
}

export function sameConfig(a: MangaConfig, b: MangaConfig): boolean {
  return JSON.stringify(complete(a)) === JSON.stringify(complete(b));
}

/** The profile whose config is exactly config, if any. */
export function matchingProfile(
  profiles: MangaProfile[],
  config: MangaConfig
): MangaProfile | undefined {
  return profiles.find((p) => sameConfig(p.config, config));
}
