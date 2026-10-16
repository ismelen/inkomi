import React, { useState } from 'react';
import { Pressable, View } from 'react-native';
import { router, useLocalSearchParams } from 'expo-router';
import DraggableFlatList, {
  RenderItemParams,
  ScaleDecorator,
} from 'react-native-draggable-flatlist';
import { useSafeAreaInsets } from 'react-native-safe-area-context';
import { useTranslation } from 'react-i18next';
import { Header } from '../../../../src/components/ui/header';
import { Screen } from '../../../../src/components/ui/screen';
import { SButton } from '../../../../src/components/ui/s-button';
import { SIcon } from '../../../../src/components/ui/s-icon';
import { SText } from '../../../../src/components/ui/s-text';
import { ListItem } from '../../../../src/models/list-item';
import { useList } from '../../../../src/stores/use-list';
import { colors, radii, spacing } from '../../../../src/theme';
import { formatBytes } from '../../../../src/utils/format';

export default function ReorderScreen() {
  const { t } = useTranslation();
  const insets = useSafeAreaInsets();
  const { id } = useLocalSearchParams<{ id: string }>();
  const group = useList((s) => s.groups[id]);
  const items = useList((s) => s.items);
  const reorderGroup = useList((s) => s.reorderGroup);
  const [order, setOrder] = useState<ListItem[]>(
    () => group?.itemIds.map((itemId) => items[itemId]).filter(Boolean) ?? []
  );

  if (!group) {
    return (
      <Screen>
        <Header title={t('reorder.title')} />
      </Screen>
    );
  }

  const sortByName = () =>
    setOrder((current) =>
      [...current].sort((a, b) => a.title.localeCompare(b.title, undefined, { numeric: true }))
    );

  const save = () => {
    reorderGroup(
      group.id,
      order.map((i) => i.id)
    );
    router.back();
  };

  const renderItem = ({ item, drag, isActive, getIndex }: RenderItemParams<ListItem>) => (
    <ScaleDecorator>
      <Pressable
        onLongPress={drag}
        delayLongPress={150}
        disabled={isActive}
        style={{
          flexDirection: 'row',
          alignItems: 'center',
          gap: spacing.md,
          padding: spacing.md,
          marginBottom: spacing.xs,
          borderRadius: radii.md,
          borderWidth: 1,
          borderColor: isActive ? colors.text : colors.border,
          backgroundColor: colors.background,
        }}
      >
        <SText variant="label" tone="textSecondary" style={{ width: 22, textAlign: 'center' }}>
          {(getIndex() ?? 0) + 1}
        </SText>
        <View style={{ flex: 1, gap: spacing.xxs }}>
          <SText variant="label" numberOfLines={1}>
            {item.title}
          </SText>
          <SText variant="small" tone="textSecondary">
            {[item.ext.toUpperCase(), item.kind === 'file' ? formatBytes(item.size) : '']
              .filter(Boolean)
              .join(' · ')}
          </SText>
        </View>
        <Pressable onPressIn={drag} hitSlop={8}>
          <SIcon name="drag_indicator" size={22} color={colors.textSecondary} />
        </Pressable>
      </Pressable>
    </ScaleDecorator>
  );

  return (
    <Screen>
      <Header title={t('reorder.title')} />
      <View style={{ paddingHorizontal: spacing.lg, gap: spacing.xs, paddingBottom: spacing.sm }}>
        <SText variant="label">
          {t('reorder.subtitle', {
            title: group.title,
            files: t('common.files', { count: order.length }),
          })}
        </SText>
        <SText variant="small" tone="textSecondary">
          {t('reorder.hint')}
        </SText>
      </View>
      <DraggableFlatList
        data={order}
        keyExtractor={(i) => i.id}
        onDragEnd={({ data }) => setOrder(data)}
        renderItem={renderItem}
        containerStyle={{ flex: 1 }}
        contentContainerStyle={{ paddingHorizontal: spacing.lg }}
      />
      <View
        style={{
          paddingHorizontal: spacing.lg,
          paddingTop: spacing.sm,
          paddingBottom: spacing.md + insets.bottom,
          gap: spacing.sm,
        }}
      >
        <SButton
          variant="secondary"
          icon="sort_by_alpha"
          label={t('reorder.sortByName')}
          onPress={sortByName}
        />
        <SButton label={t('common.save')} onPress={save} />
      </View>
    </Screen>
  );
}
