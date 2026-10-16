import React from 'react';
import { Pressable, View } from 'react-native';
import { useTranslation } from 'react-i18next';
import { colors, radii, spacing } from '../../theme';
import { IconName } from '../icons/icons';
import { SIcon } from '../ui/s-icon';
import { SSheet } from '../ui/s-sheet';
import { SText } from '../ui/s-text';

interface Props {
  visible: boolean;
  onClose: () => void;
  onFiles: () => void;
  onFolder: () => void;
}

function Option({
  icon,
  title,
  hint,
  onPress,
}: {
  icon: IconName;
  title: string;
  hint: string;
  onPress: () => void;
}) {
  return (
    <Pressable
      onPress={onPress}
      android_ripple={{ color: colors.ripple }}
      style={{
        flexDirection: 'row',
        alignItems: 'center',
        gap: spacing.md,
        padding: spacing.md,
        borderWidth: 1,
        borderColor: colors.border,
        borderRadius: radii.md,
      }}
    >
      <SIcon name={icon} size={24} />
      <View style={{ flex: 1, gap: spacing.xxs }}>
        <SText variant="label">{title}</SText>
        <SText variant="small" tone="textSecondary">
          {hint}
        </SText>
      </View>
    </Pressable>
  );
}

export function AddSheet({ visible, onClose, onFiles, onFolder }: Props) {
  const { t } = useTranslation();
  return (
    <SSheet visible={visible} title={t('add.title')} onClose={onClose}>
      <Option
        icon="description"
        title={t('add.files')}
        hint={t('add.filesHint')}
        onPress={onFiles}
      />
      <Option icon="folder" title={t('add.folder')} hint={t('add.folderHint')} onPress={onFolder} />
    </SSheet>
  );
}
