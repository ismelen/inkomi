import React, { useState } from 'react';
import { FlatList, Pressable, View } from 'react-native';
import { useTranslation } from 'react-i18next';
import { colors, radii, spacing } from '../../theme';
import { MangaConfig } from '../../models/manga-config';
import { MangaProfile } from '../../models/manga-profile';
import { useSettings } from '../../stores/use-settings';
import { SIcon } from '../ui/s-icon';
import { SIconButton } from '../ui/s-icon-button';
import { SModal } from '../ui/s-modal';
import { SSheet } from '../ui/s-sheet';
import { SText } from '../ui/s-text';
import { STextInput } from '../ui/s-text-input';

interface Props {
  visible: boolean;
  /** Called with the picked profile's config; without it the sheet only manages the profiles. */
  onPick?: (config: MangaConfig) => void;
  /** The profile matching the form, marked in the list. */
  currentId?: string;
  onClose: () => void;
}

/** Lists the saved manga profiles to load one, rename it or delete it. */
export function MangaProfileSheet({ visible, onPick, currentId, onClose }: Props) {
  const { t } = useTranslation();
  const profiles = useSettings((s) => s.mangaProfiles);
  const { renameMangaProfile, deleteMangaProfile } = useSettings.getState();
  const [renaming, setRenaming] = useState<MangaProfile | null>(null);
  const [name, setName] = useState('');
  const [deleting, setDeleting] = useState<MangaProfile | null>(null);

  const startRename = (profile: MangaProfile) => {
    setName(profile.name);
    setRenaming(profile);
  };

  const rename = () => {
    const next = name.trim();
    if (renaming && next) renameMangaProfile(renaming.id, next);
    setRenaming(null);
  };

  return (
    <>
      <SSheet visible={visible} title={t('profiles.title')} onClose={onClose}>
        <FlatList
          data={profiles}
          keyExtractor={(p) => p.id}
          ListEmptyComponent={
            <SText tone="textSecondary" style={{ textAlign: 'center', marginVertical: spacing.md }}>
              {t('profiles.empty')}
            </SText>
          }
          renderItem={({ item }) => (
            <View
              style={{
                minHeight: 48,
                flexDirection: 'row',
                alignItems: 'center',
                gap: spacing.xs,
                borderRadius: radii.md,
                backgroundColor: item.id === currentId ? colors.surface : 'transparent',
              }}
            >
              <Pressable
                disabled={!onPick}
                onPress={() => {
                  onPick?.(item.config);
                  onClose();
                }}
                style={{
                  flex: 1,
                  flexDirection: 'row',
                  alignItems: 'center',
                  gap: spacing.sm,
                  paddingHorizontal: spacing.md,
                  paddingVertical: spacing.sm,
                }}
              >
                <SText variant="body" numberOfLines={1} style={{ flex: 1 }}>
                  {item.name}
                </SText>
                {item.id === currentId && <SIcon name="check" size={18} />}
              </Pressable>
              <SIconButton
                icon="edit"
                label={t('profiles.rename')}
                size={20}
                onPress={() => startRename(item)}
              />
              <SIconButton
                icon="delete"
                label={t('common.delete')}
                size={20}
                onPress={() => setDeleting(item)}
              />
            </View>
          )}
        />
      </SSheet>
      <SModal
        visible={!!renaming}
        title={t('profiles.rename')}
        onClose={() => setRenaming(null)}
        actions={[
          { label: t('common.cancel'), onPress: () => setRenaming(null), variant: 'secondary' },
          { label: t('common.save'), onPress: rename, disabled: !name.trim() },
        ]}
      >
        <STextInput value={name} onChangeText={setName} autoFocus onSubmitEditing={rename} />
      </SModal>
      <SModal
        visible={!!deleting}
        title={t('profiles.deleteTitle', { name: deleting?.name ?? '' })}
        onClose={() => setDeleting(null)}
        actions={[
          { label: t('common.cancel'), onPress: () => setDeleting(null), variant: 'secondary' },
          {
            label: t('common.delete'),
            variant: 'danger',
            onPress: () => {
              if (deleting) deleteMangaProfile(deleting.id);
              setDeleting(null);
            },
          },
        ]}
      />
    </>
  );
}
