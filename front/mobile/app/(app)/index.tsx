import React, { useCallback, useMemo, useState } from 'react';
import { Alert, BackHandler, FlatList, Pressable, View } from 'react-native';
import { router, useFocusEffect } from 'expo-router';
import { useSafeAreaInsets } from 'react-native-safe-area-context';
import { useTranslation } from 'react-i18next';
import { errorKey } from '../../src/api/api-error';
import { AddSheet } from '../../src/components/list/add-sheet';
import { GroupModal } from '../../src/components/list/group-modal';
import { GroupRow } from '../../src/components/list/group-row';
import { ListRow } from '../../src/components/list/list-row';
import { MoveModal } from '../../src/components/list/move-modal';
import { Screen } from '../../src/components/ui/screen';
import { SButton } from '../../src/components/ui/s-button';
import { SIcon } from '../../src/components/ui/s-icon';
import { SIconButton } from '../../src/components/ui/s-icon-button';
import { SModal } from '../../src/components/ui/s-modal';
import { SSwitch } from '../../src/components/ui/s-switch';
import { SText } from '../../src/components/ui/s-text';
import { Group } from '../../src/models/group';
import { ListItem } from '../../src/models/list-item';
import { pickBookFiles, pickBookFolder } from '../../src/services/file-picker';
import { useActivity } from '../../src/stores/use-activity';
import { useList } from '../../src/stores/use-list';
import { useProfile } from '../../src/stores/use-profile';
import { useSettings } from '../../src/stores/use-settings';
import { useUpload } from '../../src/stores/use-upload';
import { colors, radii, spacing } from '../../src/theme';

type Row =
  | { type: 'group'; group: Group }
  | { type: 'item'; item: ListItem; nested: boolean; groupId?: string };

function showError(title: string, err: unknown, t: (key: string) => string) {
  Alert.alert(title, t(errorKey(err)));
}

export default function HomeScreen() {
  const { t } = useTranslation();
  const insets = useSafeAreaInsets();
  const { items, groups, rootOrder, selecting, selection } = useList();
  const list = useList.getState();
  const queueCount = useActivity((s) => s.queue.length);
  const toCloudDefault = useSettings((s) => s.toCloudDefault);
  const dropboxLinked = useProfile((s) => s.profile?.dropbox.linked ?? false);
  const sendUpload = useUpload((s) => s.send);
  const sending = useUpload((s) => s.phase === 'requesting' || s.phase === 'uploading');

  const [toCloud, setToCloud] = useState(toCloudDefault);
  const [linking, setLinking] = useState(false);
  const [addOpen, setAddOpen] = useState(false);
  const [editingGroup, setEditingGroup] = useState<Group | null>(null);
  const [confirmDelete, setConfirmDelete] = useState(false);
  const [moveOpen, setMoveOpen] = useState(false);

  useFocusEffect(
    useCallback(() => {
      const sub = BackHandler.addEventListener('hardwareBackPress', () => {
        if (!useList.getState().selecting) return false;
        useList.getState().clearSelection();
        return true;
      });
      return () => sub.remove();
    }, [])
  );

  const rows = useMemo<Row[]>(
    () =>
      rootOrder.flatMap((id): Row[] => {
        const group = groups[id];
        if (group) {
          const children = group.collapsed
            ? []
            : group.itemIds
                .filter((itemId) => items[itemId])
                .map((itemId) => ({
                  type: 'item' as const,
                  item: items[itemId],
                  nested: true,
                  groupId: id,
                }));
          return [{ type: 'group', group }, ...children];
        }
        return items[id] ? [{ type: 'item', item: items[id], nested: false }] : [];
      }),
    [rootOrder, groups, items]
  );

  const pressRow = (id: string, open: () => void) => {
    if (selecting) list.toggleSelected(id);
    else open();
  };

  const longPressRow = (id: string) => {
    if (selecting) list.toggleSelected(id);
    else list.startSelection(id);
  };

  const toggleCloud = async (value: boolean) => {
    if (!value || dropboxLinked) {
      setToCloud(value);
      return;
    }
    setLinking(true);
    try {
      if (await useProfile.getState().linkDropbox()) setToCloud(true);
    } catch (err) {
      showError(t('settings.dropbox'), err, t);
    } finally {
      setLinking(false);
    }
  };

  const addFiles = async () => {
    setAddOpen(false);
    try {
      list.addFiles(await pickBookFiles());
    } catch (err) {
      showError(t('add.files'), err, t);
    }
  };

  const addFolder = async () => {
    setAddOpen(false);
    try {
      const folder = await pickBookFolder();
      if (!folder) return;
      if (folder.files.length === 0) Alert.alert(t('add.folder'), t('add.emptyFolder'));
      else list.addFolder(folder);
    } catch (err) {
      showError(t('add.folder'), err, t);
    }
  };

  const openOptions = (id: string) => router.push({ pathname: '/options/[id]', params: { id } });
  const openReorder = (id: string) =>
    router.push({ pathname: '/group/[id]/reorder', params: { id } });

  const renderRow = ({ item: row }: { item: Row }) => {
    if (row.type === 'group') {
      const { group } = row;
      return (
        <GroupRow
          group={group}
          selecting={selecting}
          selected={selection.includes(group.id)}
          onPress={() => pressRow(group.id, () => setEditingGroup(group))}
          onLongPress={() => longPressRow(group.id)}
          onSettings={() => setEditingGroup(group)}
          onToggleCollapsed={() => list.updateGroup(group.id, { collapsed: !group.collapsed })}
        />
      );
    }
    const { item, nested, groupId } = row;
    return (
      <ListRow
        item={item}
        nested={nested}
        selecting={selecting}
        selected={selection.includes(item.id)}
        onPress={() => pressRow(item.id, () => openOptions(item.id))}
        onLongPress={() => longPressRow(item.id)}
        onOptions={() => openOptions(item.id)}
        onReorder={groupId ? () => openReorder(groupId) : undefined}
      />
    );
  };

  return (
    <Screen>
      {selecting ? (
        <View style={styles.header}>
          <SIconButton
            icon="close"
            label={t('home.cancelSelection')}
            onPress={list.clearSelection}
          />
          <SText variant="title" style={{ flex: 1 }}>
            {t('home.selected', { count: selection.length })}
          </SText>
          <SIconButton
            icon="checklist"
            label={t('home.toggleAll')}
            onPress={list.toggleSelectAll}
          />
        </View>
      ) : (
        <View style={styles.header}>
          <SText variant="brand" style={{ flex: 1, paddingLeft: spacing.sm }}>
            INKOMI
          </SText>
          <SIconButton
            icon="history"
            label={t('home.history')}
            onPress={() => router.push('/history')}
          />
          <SIconButton
            icon="progress_activity"
            label={t('home.activity')}
            badge={queueCount}
            onPress={() => router.push('/activity')}
          />
          <SIconButton
            icon="settings"
            label={t('home.settings')}
            onPress={() => router.push('/settings')}
          />
        </View>
      )}

      {!selecting && (
        <View style={{ flexDirection: 'row', gap: spacing.sm, paddingHorizontal: spacing.lg }}>
          <Pressable
            onPress={() => router.push('/search')}
            style={{
              flex: 1,
              flexDirection: 'row',
              alignItems: 'center',
              gap: spacing.sm,
              height: 44,
              paddingHorizontal: spacing.md,
              borderRadius: radii.md,
              borderWidth: 1,
              borderColor: colors.border,
            }}
          >
            <SIcon name="search" size={20} color={colors.textSecondary} />
            <SText tone="textSecondary">{t('home.searchPlaceholder')}</SText>
          </Pressable>
          <SIconButton
            icon="attach_file"
            label={t('add.title')}
            onPress={() => setAddOpen(true)}
            style={{
              borderWidth: 1,
              borderColor: colors.border,
              borderRadius: radii.md,
              width: 44,
              height: 44,
            }}
          />
        </View>
      )}

      <FlatList
        data={rows}
        keyExtractor={(row) => (row.type === 'group' ? `g-${row.group.id}` : row.item.id)}
        renderItem={renderRow}
        contentContainerStyle={{ padding: spacing.lg, gap: spacing.xs, flexGrow: 1 }}
        ListEmptyComponent={
          <View
            style={{ flex: 1, alignItems: 'center', justifyContent: 'center', padding: spacing.xl }}
          >
            <SText tone="textSecondary" style={{ textAlign: 'center' }}>
              {t('home.empty')}
            </SText>
          </View>
        }
      />

      <View
        style={{
          flexDirection: 'row',
          alignItems: 'center',
          gap: spacing.md,
          paddingHorizontal: spacing.lg,
          paddingTop: spacing.md,
          paddingBottom: spacing.md + insets.bottom,
          borderTopWidth: 1,
          borderTopColor: colors.surface,
        }}
      >
        {selecting ? (
          <>
            <SButton
              variant="secondary"
              icon="delete"
              label={t('common.delete')}
              onPress={() => setConfirmDelete(true)}
              disabled={selection.length === 0}
              style={{ flex: 1 }}
            />
            <SButton
              variant="secondary"
              icon="drive_file_move"
              label={t('common.move')}
              onPress={() => setMoveOpen(true)}
              disabled={selection.length === 0}
              style={{ flex: 1 }}
            />
          </>
        ) : (
          <>
            <View style={{ flexDirection: 'row', alignItems: 'center', gap: spacing.sm }}>
              <SSwitch
                value={toCloud}
                onValueChange={toggleCloud}
                disabled={linking}
                accessibilityLabel={t('home.dropbox')}
              />
              <SText variant="label">{t('home.dropbox')}</SText>
            </View>
            <SButton
              label={t('home.send')}
              onPress={() => void sendUpload(toCloud)}
              disabled={rows.length === 0 || sending}
              style={{ flex: 1 }}
            />
          </>
        )}
      </View>

      <AddSheet
        visible={addOpen}
        onClose={() => setAddOpen(false)}
        onFiles={addFiles}
        onFolder={addFolder}
      />
      <GroupModal group={editingGroup} onClose={() => setEditingGroup(null)} />
      <MoveModal visible={moveOpen} count={selection.length} onClose={() => setMoveOpen(false)} />
      <SModal
        visible={confirmDelete}
        title={t('delete.title', { count: selection.length })}
        message={t('delete.message')}
        onClose={() => setConfirmDelete(false)}
        actions={[
          {
            label: t('common.cancel'),
            onPress: () => setConfirmDelete(false),
            variant: 'secondary',
          },
          {
            label: t('common.delete'),
            variant: 'danger',
            onPress: () => {
              list.remove(selection);
              list.clearSelection();
              setConfirmDelete(false);
            },
          },
        ]}
      />
    </Screen>
  );
}

const styles = {
  header: {
    height: 56,
    flexDirection: 'row' as const,
    alignItems: 'center' as const,
    paddingHorizontal: spacing.sm,
    gap: spacing.xs,
  },
};
