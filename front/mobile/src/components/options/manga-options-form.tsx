import React, { useState } from 'react';
import { View } from 'react-native';
import { useTranslation } from 'react-i18next';
import { spacing } from '../../theme';
import { cropDisabled, MangaConfig, mangaRanges } from '../../models/manga-config';
import { SIconButton } from '../ui/s-icon-button';
import { SSlider } from '../ui/s-slider';
import { SSwitch } from '../ui/s-switch';
import { SectionTitle, SettingRow } from '../ui/setting-row';

interface Props {
  config: MangaConfig;
  onChange: (config: MangaConfig) => void;
}

type BoolKey = 'rotateDoublePage' | 'enlarge' | 'embed' | 'grayscale' | 'dithering' | 'palette8bit';
type SharpenKey = 'radius' | 'x1' | 'y2' | 'y3' | 'm1' | 'm2';

const toggles: { key: BoolKey; label: string }[] = [
  { key: 'rotateDoublePage', label: 'options.rotateDoublePage' },
  { key: 'enlarge', label: 'options.enlarge' },
  { key: 'embed', label: 'options.embed' },
  { key: 'grayscale', label: 'options.grayscale' },
  { key: 'dithering', label: 'options.dithering' },
  { key: 'palette8bit', label: 'options.palette8bit' },
];

const sharpenFields: { key: SharpenKey; label: string }[] = [
  { key: 'radius', label: 'options.sharpenRadius' },
  { key: 'x1', label: 'options.sharpenX1' },
  { key: 'y2', label: 'options.sharpenY2' },
  { key: 'y3', label: 'options.sharpenY3' },
  { key: 'm1', label: 'options.sharpenM1' },
  { key: 'm2', label: 'options.sharpenM2' },
];

/** IMAGEN and AJUSTES sections of the manga options (cbz, pdf and groups). */
export function MangaOptionsForm({ config, onChange }: Props) {
  const { t } = useTranslation();
  const [sharpenOpen, setSharpenOpen] = useState(false);
  const set = <K extends keyof MangaConfig>(key: K, value: MangaConfig[K]) =>
    onChange({ ...config, [key]: value });
  const cropOn = config.cropTolerance !== cropDisabled;

  return (
    <View>
      <SectionTitle>{t('options.image')}</SectionTitle>
      <SSlider
        label={t('options.gamma')}
        value={config.gamma}
        range={mangaRanges.gamma}
        onChange={(v) => set('gamma', v)}
      />
      <SSlider
        label={t('options.brightness')}
        value={config.brightness}
        range={mangaRanges.brightness}
        onChange={(v) => set('brightness', v)}
      />
      <SSlider
        label={t('options.contrast')}
        value={config.contrast}
        range={mangaRanges.contrast}
        onChange={(v) => set('contrast', v)}
      />
      <SSlider
        label={t('options.blackPoint')}
        value={config.blackPointThreshold}
        range={mangaRanges.blackPointThreshold}
        onChange={(v) => set('blackPointThreshold', v)}
      />
      <SSlider
        label={t('options.whitePoint')}
        value={config.whitePointThreshold}
        range={mangaRanges.whitePointThreshold}
        onChange={(v) => set('whitePointThreshold', v)}
      />
      <SSlider
        label={t('options.cropTolerance')}
        value={cropOn ? config.cropTolerance : 0}
        range={mangaRanges.cropTolerance}
        disabled={!cropOn}
        onChange={(v) => set('cropTolerance', v)}
        accessory={
          <SSwitch
            value={cropOn}
            accessibilityLabel={t('options.crop')}
            onValueChange={(on) => set('cropTolerance', on ? 0.1 : cropDisabled)}
          />
        }
      />
      <SSlider
        label={t('options.quality')}
        value={config.quality}
        range={mangaRanges.quality}
        format={(v) => `${v}%`}
        onChange={(v) => set('quality', v)}
      />

      <SectionTitle>{t('options.settings')}</SectionTitle>
      {toggles.map(({ key, label }) => (
        <SettingRow
          key={key}
          label={t(label)}
          right={<SSwitch value={config[key]} onValueChange={(v) => set(key, v)} />}
        />
      ))}
      <SettingRow
        label={t('options.sharpen')}
        right={
          <View style={{ flexDirection: 'row', alignItems: 'center', gap: spacing.xs }}>
            {config.sharpen.enabled && (
              <SIconButton
                icon={sharpenOpen ? 'expand_less' : 'expand_more'}
                label={t('options.sharpen')}
                size={20}
                onPress={() => setSharpenOpen((o) => !o)}
              />
            )}
            <SSwitch
              value={config.sharpen.enabled}
              onValueChange={(enabled) => {
                set('sharpen', { ...config.sharpen, enabled });
                setSharpenOpen(enabled);
              }}
            />
          </View>
        }
      />
      {config.sharpen.enabled && sharpenOpen && (
        <View style={{ paddingLeft: spacing.md }}>
          {sharpenFields.map(({ key, label }) => (
            <SSlider
              key={key}
              label={t(label)}
              value={config.sharpen[key]}
              range={mangaRanges[key]}
              onChange={(v) => set('sharpen', { ...config.sharpen, [key]: v })}
            />
          ))}
        </View>
      )}
    </View>
  );
}
