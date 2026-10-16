import React from 'react';
import { Pressable, View } from 'react-native';
import { useTranslation } from 'react-i18next';
import { colors, radii, spacing } from '../../theme';
import { Group } from '../../models/group';
import { SIcon } from '../ui/s-icon';
import { SIconButton } from '../ui/s-icon-button';
import { SText } from '../ui/s-text';

interface Props {
  group: Group;
  selecting: boolean;
  selected: boolean;
  onPress: () => void;
  onLongPress: () => void;
  onSettings: () => void;
  onToggleCollapsed: () => void;
}

export function GroupRow({
  group,
  selecting,
  selected,
  onPress,
  onLongPress,
  onSettings,
  onToggleCollapsed,
}: Props) {
  const { t } = useTranslation();
  const details = [
    t('common.files', { count: group.itemIds.length }),
    group.join ? t('group.join') : null,
  ]
    .filter(Boolean)
    .join(' · ');

  return (
    <Pressable
      onPress={onPress}
      onLongPress={onLongPress}
      android_ripple={{ color: colors.ripple }}
      style={{
        flexDirection: 'row',
        alignItems: 'center',
        gap: spacing.sm,
        paddingVertical: spacing.sm,
        paddingLeft: spacing.md,
        paddingRight: spacing.xs,
        borderRadius: radii.md,
        backgroundColor: colors.surface,
        borderWidth: selected ? 1 : 0,
        borderColor: colors.text,
      }}
    >
      {selecting && (
        <SIcon
          name={selected ? 'radio_button_checked' : 'radio_button_unchecked'}
          size={20}
          color={selected ? colors.text : colors.textSecondary}
        />
      )}
      <SIcon name="folder" size={20} color={colors.textSecondary} />
      <View style={{ flex: 1, gap: spacing.xxs }}>
        <SText variant="label" numberOfLines={1}>
          {group.title}
        </SText>
        <SText variant="small" tone={group.join ? 'info' : 'textSecondary'} numberOfLines={1}>
          {details}
        </SText>
      </View>
      {!selecting && (
        <SIconButton
          icon="settings"
          label={t('group.title')}
          size={20}
          color={colors.textSecondary}
          onPress={onSettings}
        />
      )}
      <SIconButton
        icon={group.collapsed ? 'expand_more' : 'expand_less'}
        label={group.title}
        size={20}
        color={colors.textSecondary}
        onPress={onToggleCollapsed}
      />
    </Pressable>
  );
}
