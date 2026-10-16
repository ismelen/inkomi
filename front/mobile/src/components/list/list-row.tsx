import React from 'react';
import { Pressable, View } from 'react-native';
import { useTranslation } from 'react-i18next';
import { colors, radii, spacing } from '../../theme';
import { ListItem } from '../../models/list-item';
import { formatBytes } from '../../utils/format';
import { SIcon } from '../ui/s-icon';
import { SIconButton } from '../ui/s-icon-button';
import { SText } from '../ui/s-text';

interface Props {
  item: ListItem;
  /** Item inside a group: indented, with the drag handle. */
  nested?: boolean;
  selecting: boolean;
  selected: boolean;
  onPress: () => void;
  onLongPress: () => void;
  onOptions: () => void;
  onReorder?: () => void;
}

export function ListRow({
  item,
  nested,
  selecting,
  selected,
  onPress,
  onLongPress,
  onOptions,
  onReorder,
}: Props) {
  const { t } = useTranslation();
  const details = [
    item.ext.toUpperCase(),
    item.kind === 'file' ? formatBytes(item.size) : (item.author ?? t('home.library')),
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
        paddingLeft: nested ? spacing.lg : spacing.md,
        paddingRight: spacing.xs,
        borderRadius: radii.md,
        backgroundColor: selected ? colors.surface : 'transparent',
      }}
    >
      {selecting && (
        <SIcon
          name={selected ? 'radio_button_checked' : 'radio_button_unchecked'}
          size={20}
          color={selected ? colors.text : colors.textSecondary}
        />
      )}
      {nested && !selecting && onReorder && (
        <SIconButton
          icon="drag_indicator"
          label={t('reorder.title')}
          size={18}
          color={colors.textSecondary}
          onPress={onReorder}
          onLongPress={onReorder}
        />
      )}
      <View style={{ flex: 1, gap: spacing.xxs }}>
        <SText variant="label" numberOfLines={1}>
          {item.title}
        </SText>
        <SText variant="small" tone="textSecondary" numberOfLines={1}>
          {details}
        </SText>
      </View>
      {!selecting && (
        <SIconButton
          icon="tune"
          label={t('options.title')}
          size={20}
          color={colors.textSecondary}
          onPress={onOptions}
        />
      )}
    </Pressable>
  );
}
