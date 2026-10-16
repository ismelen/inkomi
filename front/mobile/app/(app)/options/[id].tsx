import React, { useMemo, useState } from 'react';
import { ScrollView, View } from 'react-native';
import { router, useLocalSearchParams } from 'expo-router';
import { useSafeAreaInsets } from 'react-native-safe-area-context';
import { useTranslation } from 'react-i18next';
import { MangaOptionsForm } from '../../../src/components/options/manga-options-form';
import { Header } from '../../../src/components/ui/header';
import { Screen } from '../../../src/components/ui/screen';
import { SButton } from '../../../src/components/ui/s-button';
import { SSegmented } from '../../../src/components/ui/s-segmented';
import { SText } from '../../../src/components/ui/s-text';
import { STextInput } from '../../../src/components/ui/s-text-input';
import { isManga } from '../../../src/models/list-item';
import {
  defaultMangaConfig,
  MangaConfig,
  validateMangaConfig,
} from '../../../src/models/manga-config';
import { useList } from '../../../src/stores/use-list';
import { useSettings } from '../../../src/stores/use-settings';
import { spacing } from '../../../src/theme';

/** Default manga, group ("Opciones manga (todos)") or item options; "default" is the settings one. */
export default function OptionsScreen() {
  const { t } = useTranslation();
  const insets = useSafeAreaInsets();
  const { id } = useLocalSearchParams<{ id: string }>();
  const item = useList((s) => s.items[id]);
  const group = useList((s) => s.groups[id]);
  const parent = useList((s) => (item?.groupId ? s.groups[item.groupId] : undefined));
  const { updateItem, updateGroup } = useList.getState();
  const { defaultManga, setDefaultManga } = useSettings();

  // What the target uses when it has no options of its own.
  const inherited = useMemo<MangaConfig>(() => {
    if (id === 'default') return defaultMangaConfig;
    if (group) return defaultManga;
    return parent?.config ?? defaultManga;
  }, [id, group, parent, defaultManga]);

  const own = id === 'default' ? defaultManga : (group?.config ?? item?.config);
  const [config, setConfig] = useState<MangaConfig>(own ?? inherited);
  const [touched, setTouched] = useState(!!own);
  const [title, setTitle] = useState(item?.title ?? '');
  const [rtl, setRtl] = useState(item?.rtl ?? false);

  if (id !== 'default' && !item && !group) {
    return (
      <Screen>
        <Header title={t('options.title')} />
      </Screen>
    );
  }

  const showManga = id === 'default' || !!group || (item && isManga(item.ext));
  const error = showManga ? validateMangaConfig(config) : null;
  const screenTitle =
    id === 'default'
      ? t('options.defaultTitle')
      : group
        ? t('options.groupTitle', { title: group.title })
        : t('options.title');

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
    if (id === 'default') setDefaultManga(config);
    else if (group) updateGroup(group.id, { config: touched ? config : undefined });
    else if (item) {
      updateItem(item.id, {
        title: title.trim() || item.title,
        rtl,
        config: showManga && touched ? config : undefined,
      });
    }
    router.back();
  };

  return (
    <Screen>
      <Header
        title={screenTitle}
        right={<SButton variant="ghost" compact label={t('options.reset')} onPress={reset} />}
      />
      <ScrollView
        contentContainerStyle={{
          paddingHorizontal: spacing.lg,
          paddingBottom: spacing.xl,
          gap: spacing.sm,
        }}
      >
        {item && (
          <>
            <STextInput label={t('common.title')} value={title} onChangeText={setTitle} />
            <SText variant="section" style={{ marginTop: spacing.sm }}>
              {t('common.readingDirection')}
            </SText>
            <SSegmented
              value={rtl}
              onChange={setRtl}
              options={[
                { value: false, label: t('common.ltr') },
                { value: true, label: t('common.rtl') },
              ]}
            />
          </>
        )}
        {showManga ? (
          <MangaOptionsForm config={config} onChange={change} />
        ) : (
          <SText variant="small" tone="textSecondary" style={{ marginTop: spacing.md }}>
            {t('options.epubHint')}
          </SText>
        )}
      </ScrollView>
      <View
        style={{
          paddingHorizontal: spacing.lg,
          paddingTop: spacing.sm,
          paddingBottom: spacing.md + insets.bottom,
          gap: spacing.sm,
        }}
      >
        {!!error && (
          <SText variant="small" tone="danger" style={{ textAlign: 'center' }}>
            {t(`options.${error}`)}
          </SText>
        )}
        <SButton label={t('common.save')} onPress={save} disabled={!!error} />
      </View>
    </Screen>
  );
}
