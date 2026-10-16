import { useMemo, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { isManga } from '../../models/list-item';
import { defaultMangaConfig, MangaConfig, validateMangaConfig } from '../../models/manga-config';
import { useList } from '../../stores/use-list';
import { useSettings } from '../../stores/use-settings';
import { useToast } from '../../stores/use-toast';
import { Button } from '../ui/button';
import { SectionLabel, TextInput } from '../ui/field';
import { IconButton } from '../ui/icon-button';
import { Segmented } from '../ui/segmented';
import { Switch } from '../ui/switch';
import { MangaOptionsForm } from './manga-options-form';

export const defaultOptionsId = 'default';

interface Props {
  /** A list item, a group, or "default" (the default manga of the settings). */
  targetId: string;
  onClose?: () => void;
  onSaved?: () => void;
}

/**
 * Options of the focused element of the list (or the default manga). Remount it with a key per
 * target: the draft state starts from the target.
 */
export function OptionsPanel({ targetId, onClose, onSaved }: Props) {
  const { t } = useTranslation();
  const item = useList((s) => s.items[targetId]);
  const group = useList((s) => s.groups[targetId]);
  const parent = useList((s) => (item?.groupId ? s.groups[item.groupId] : undefined));
  const { defaultManga, setDefaultManga } = useSettings();
  const isDefault = targetId === defaultOptionsId;

  // What the target uses when it has no options of its own.
  const inherited = useMemo<MangaConfig>(() => {
    if (isDefault) return defaultMangaConfig;
    if (group) return defaultManga;
    return parent?.config ?? defaultManga;
  }, [isDefault, group, parent, defaultManga]);

  const own = isDefault ? defaultManga : (group?.config ?? item?.config);
  const [config, setConfig] = useState<MangaConfig>(own ?? inherited);
  const [touched, setTouched] = useState(!!own);
  const [title, setTitle] = useState(item?.title ?? group?.title ?? '');
  const [rtl, setRtl] = useState(item?.rtl ?? group?.rtl ?? false);
  const [join, setJoin] = useState(group?.join ?? false);

  if (!isDefault && !item && !group) return null;

  const showManga = isDefault || !!group || (item && isManga(item.ext));
  const error = showManga ? validateMangaConfig(config) : null;
  const subtitle = item?.kind === 'file' ? item.name : (item?.title ?? group?.title);

  const change = (next: MangaConfig) => {
    setConfig(next);
    setTouched(true);
  };

  const reset = () => {
    setConfig(inherited);
    setTouched(false);
    if (item) setRtl(parent?.rtl ?? false);
  };

  const save = () => {
    const ownConfig = showManga && touched ? config : undefined;
    if (isDefault) setDefaultManga(config);
    else if (group) {
      useList.getState().updateGroup(group.id, {
        title: title.trim() || group.title,
        join,
        rtl,
        config: ownConfig,
      });
    } else if (item) {
      useList.getState().updateItem(item.id, {
        title: title.trim() || item.title,
        rtl,
        config: ownConfig,
      });
    }
    useToast.getState().show(t('history.saved'));
    onSaved?.();
  };

  return (
    <div className="flex h-full min-h-0 flex-col">
      <div className="flex items-start gap-2 px-5 pt-5">
        <div className="min-w-0 flex-1">
          <h2 className="text-xl font-semibold">
            {isDefault ? t('options.defaultTitle') : group ? t('group.title') : t('options.title')}
          </h2>
          {subtitle && <p className="truncate text-sm text-muted">{subtitle}</p>}
        </div>
        <button
          type="button"
          onClick={reset}
          className="mt-1 cursor-pointer text-sm font-semibold text-primary hover:underline"
        >
          {t('options.reset')}
        </button>
        {onClose && <IconButton icon="close" label={t('common.close')} onClick={onClose} />}
      </div>

      <div className="min-h-0 flex-1 overflow-y-auto px-5 pt-4 pb-4">
        {!isDefault && (
          <>
            <TextInput
              label={t('common.title')}
              value={title}
              onChange={(e) => setTitle(e.target.value)}
              className="[&_div]:h-10"
            />
            {group && (
              <div className="mt-4 flex items-center gap-3">
                <span className="flex-1 text-sm">{t('group.join')}</span>
                <Switch label={t('group.join')} checked={join} onChange={setJoin} />
              </div>
            )}
            <SectionLabel className="mt-4 mb-1.5">{t('common.readingDirection')}</SectionLabel>
            <Segmented
              label={t('common.readingDirection')}
              value={rtl ? 'rtl' : 'ltr'}
              onChange={(v) => setRtl(v === 'rtl')}
              options={[
                { value: 'ltr', label: t('common.ltr') },
                { value: 'rtl', label: t('common.rtl') },
              ]}
            />
          </>
        )}
        {group && <SectionLabel className="mt-5 -mb-2">{t('group.mangaOptions')}</SectionLabel>}
        {isDefault && <p className="text-sm text-muted">{t('options.defaultHint')}</p>}
        {showManga ? (
          <MangaOptionsForm config={config} onChange={change} />
        ) : (
          <p className="mt-4 text-sm text-muted">{t('options.epubHint')}</p>
        )}
      </div>

      <div className="pb-safe px-5 pt-2 pb-5">
        {error && (
          <p role="alert" className="mb-2 text-center text-sm text-danger">
            {t(`options.${error}`)}
          </p>
        )}
        <Button size="lg" className="w-full" disabled={!!error} onClick={save}>
          {t('common.save')}
        </Button>
      </div>
    </div>
  );
}
