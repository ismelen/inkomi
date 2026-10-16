import { useState } from 'react';
import { useTranslation } from 'react-i18next';
import { cropDisabled, MangaConfig, mangaRanges } from '../../models/manga-config';
import { SectionLabel } from '../ui/field';
import { Slider } from '../ui/slider';
import { Switch } from '../ui/switch';

interface Props {
  config: MangaConfig;
  onChange: (config: MangaConfig) => void;
}

type BoolKey = 'rotateDoublePage' | 'enlarge' | 'embed' | 'grayscale' | 'dithering' | 'palette8bit';
type SharpenKey = 'radius' | 'x1' | 'y2' | 'y3' | 'm1' | 'm2';

// The design shows these; "Mostrar todas las opciones" adds the rest.
const mainToggles: { key: BoolKey; label: string }[] = [
  { key: 'rotateDoublePage', label: 'options.rotateDoublePage' },
  { key: 'grayscale', label: 'options.grayscale' },
  { key: 'dithering', label: 'options.dithering' },
];

const moreToggles: { key: BoolKey; label: string }[] = [
  { key: 'enlarge', label: 'options.enlarge' },
  { key: 'embed', label: 'options.embed' },
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

function ToggleRow({
  label,
  checked,
  onChange,
}: {
  label: string;
  checked: boolean;
  onChange: (value: boolean) => void;
}) {
  return (
    <div className="flex min-h-[26px] items-center gap-3">
      <span className="flex-1 text-[13px]">{label}</span>
      <Switch label={label} checked={checked} onChange={onChange} />
    </div>
  );
}

/** IMAGEN and AJUSTES of the manga options (cbz, pdf and groups). */
export function MangaOptionsForm({ config, onChange }: Props) {
  const { t } = useTranslation();
  const [showAll, setShowAll] = useState(false);
  const set = <K extends keyof MangaConfig>(key: K, value: MangaConfig[K]) =>
    onChange({ ...config, [key]: value });
  const cropOn = config.cropTolerance !== cropDisabled;

  return (
    <div>
      <SectionLabel className="mt-4 mb-1">{t('options.image')}</SectionLabel>
      <Slider
        label={t('options.gamma')}
        value={config.gamma}
        range={mangaRanges.gamma}
        onChange={(v) => set('gamma', v)}
      />
      <Slider
        label={t('options.brightness')}
        value={config.brightness}
        range={mangaRanges.brightness}
        onChange={(v) => set('brightness', v)}
      />
      <Slider
        label={t('options.contrast')}
        value={config.contrast}
        range={mangaRanges.contrast}
        onChange={(v) => set('contrast', v)}
      />
      <Slider
        label={t('options.blackPoint')}
        value={config.blackPointThreshold}
        range={mangaRanges.blackPointThreshold}
        onChange={(v) => set('blackPointThreshold', v)}
      />
      <Slider
        label={t('options.whitePoint')}
        value={config.whitePointThreshold}
        range={mangaRanges.whitePointThreshold}
        onChange={(v) => set('whitePointThreshold', v)}
      />
      {showAll && (
        <Slider
          label={t('options.cropTolerance')}
          value={cropOn ? config.cropTolerance : 0}
          range={mangaRanges.cropTolerance}
          disabled={!cropOn}
          onChange={(v) => set('cropTolerance', v)}
          accessory={
            <Switch
              label={t('options.crop')}
              checked={cropOn}
              onChange={(on) => set('cropTolerance', on ? 0.1 : cropDisabled)}
            />
          }
        />
      )}
      <Slider
        label={t('options.quality')}
        value={config.quality}
        range={mangaRanges.quality}
        format={(v) => `${v}%`}
        onChange={(v) => set('quality', v)}
      />

      <SectionLabel className="mt-4 mb-1.5">{t('options.settings')}</SectionLabel>
      <div className="flex flex-col gap-1">
        {mainToggles.map(({ key, label }) => (
          <ToggleRow
            key={key}
            label={t(label)}
            checked={config[key]}
            onChange={(v) => set(key, v)}
          />
        ))}
        {showAll &&
          moreToggles.map(({ key, label }) => (
            <ToggleRow
              key={key}
              label={t(label)}
              checked={config[key]}
              onChange={(v) => set(key, v)}
            />
          ))}
        <ToggleRow
          label={t('options.sharpen')}
          checked={config.sharpen.enabled}
          onChange={(enabled) => set('sharpen', { ...config.sharpen, enabled })}
        />
      </div>
      {showAll && config.sharpen.enabled && (
        <div className="mt-1 border-l-2 border-surface pl-3">
          {sharpenFields.map(({ key, label }) => (
            <Slider
              key={key}
              label={t(label)}
              value={config.sharpen[key]}
              range={mangaRanges[key]}
              onChange={(v) => set('sharpen', { ...config.sharpen, [key]: v })}
            />
          ))}
        </div>
      )}
      <button
        type="button"
        onClick={() => setShowAll((s) => !s)}
        className="mt-3 cursor-pointer text-sm text-info hover:underline"
      >
        {showAll ? t('options.showLess') : t('options.showAll')}
      </button>
    </div>
  );
}
