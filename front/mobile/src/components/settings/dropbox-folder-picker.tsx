import React, { useCallback, useEffect, useState } from 'react';
import { ActivityIndicator, FlatList, Pressable, View } from 'react-native';
import { useTranslation } from 'react-i18next';
import { errorKey } from '../../api/api-error';
import { DropboxFolderEntry } from '../../api/types/users';
import { usersApi } from '../../api/users-api';
import { colors, radii, spacing } from '../../theme';
import { SButton } from '../ui/s-button';
import { SIcon } from '../ui/s-icon';
import { SSheet } from '../ui/s-sheet';
import { SText } from '../ui/s-text';
import { STextInput } from '../ui/s-text-input';

interface Props {
  visible: boolean;
  /** Folder to open with, "" being the root. */
  initial: string;
  onSelect: (folder: string) => Promise<void>;
  onClose: () => void;
}

function parentOf(path: string): string {
  return path.slice(0, Math.max(0, path.lastIndexOf('/')));
}

/** Browses the user's Dropbox folder by folder and picks the destination of the books. */
export function DropboxFolderPicker({ visible, initial, onSelect, onClose }: Props) {
  const { t } = useTranslation();
  const [path, setPath] = useState(initial);
  const [folders, setFolders] = useState<DropboxFolderEntry[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string>();
  const [creating, setCreating] = useState(false);
  const [newName, setNewName] = useState('');
  const [busy, setBusy] = useState(false);

  const open = useCallback(async (target: string) => {
    setLoading(true);
    setError(undefined);
    try {
      const res = await usersApi.listDropboxFolders(target);
      setPath(res.path);
      setFolders(res.folders);
    } catch (err) {
      setError(errorKey(err));
      // a folder that no longer exists falls back to the root
      if (target !== '') {
        try {
          const res = await usersApi.listDropboxFolders('');
          setPath(res.path);
          setFolders(res.folders);
        } catch {
          // the first error is the one shown
        }
      }
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    if (!visible) return;
    setCreating(false);
    setNewName('');
    void open(initial);
  }, [visible, initial, open]);

  const select = async () => {
    setBusy(true);
    try {
      await onSelect(path);
      onClose();
    } catch (err) {
      setError(errorKey(err));
    } finally {
      setBusy(false);
    }
  };

  const create = async () => {
    const name = newName.trim().replace(/\//g, '');
    if (!name) return;
    setBusy(true);
    setError(undefined);
    try {
      const folder = await usersApi.createDropboxFolder(`${path}/${name}`);
      setCreating(false);
      setNewName('');
      await open(folder.path);
    } catch (err) {
      setError(errorKey(err));
    } finally {
      setBusy(false);
    }
  };

  const crumbs = path === '' ? [] : path.split('/').filter(Boolean);

  return (
    <SSheet visible={visible} title={t('settings.folderPicker.title')} onClose={onClose}>
      <View style={{ flexDirection: 'row', alignItems: 'center', gap: spacing.sm }}>
        <Pressable
          disabled={path === '' || loading}
          onPress={() => void open(parentOf(path))}
          hitSlop={8}
          style={{ opacity: path === '' ? 0.3 : 1 }}
        >
          <SIcon name="chevron_left" size={24} />
        </Pressable>
        <SText variant="body" numberOfLines={1} style={{ flex: 1 }}>
          {crumbs.length === 0 ? t('settings.folderPicker.root') : `/${crumbs.join('/')}`}
        </SText>
        <Pressable onPress={() => setCreating((c) => !c)} hitSlop={8}>
          <SIcon name="add" size={24} />
        </Pressable>
      </View>

      {creating && (
        <View style={{ flexDirection: 'row', alignItems: 'flex-end', gap: spacing.sm }}>
          <View style={{ flex: 1 }}>
            <STextInput
              value={newName}
              onChangeText={setNewName}
              placeholder={t('settings.folderPicker.newName')}
              autoFocus
              autoCapitalize="none"
              autoCorrect={false}
              onSubmitEditing={() => void create()}
            />
          </View>
          <SButton
            compact
            label={t('settings.folderPicker.create')}
            loading={busy}
            disabled={!newName.trim()}
            onPress={() => void create()}
          />
        </View>
      )}

      {!!error && <SText tone="danger">{t(error)}</SText>}

      {loading ? (
        <ActivityIndicator color={colors.primary} style={{ marginVertical: spacing.lg }} />
      ) : (
        <FlatList
          data={folders}
          keyExtractor={(f) => f.path}
          style={{ maxHeight: 360 }}
          ListEmptyComponent={
            <SText tone="textSecondary" style={{ textAlign: 'center', marginVertical: spacing.md }}>
              {t('settings.folderPicker.empty')}
            </SText>
          }
          renderItem={({ item }) => (
            <Pressable
              onPress={() => void open(item.path)}
              style={{
                height: 48,
                flexDirection: 'row',
                alignItems: 'center',
                gap: spacing.sm,
                paddingHorizontal: spacing.md,
                borderRadius: radii.md,
              }}
            >
              <SIcon name="folder" size={20} />
              <SText variant="body" numberOfLines={1} style={{ flex: 1 }}>
                {item.name}
              </SText>
              <SIcon name="chevron_right" size={20} />
            </Pressable>
          )}
        />
      )}

      <SButton
        icon="check"
        label={t('settings.folderPicker.select', {
          folder: crumbs.length === 0 ? t('settings.folderPicker.root') : crumbs[crumbs.length - 1],
        })}
        loading={busy && !creating}
        disabled={loading}
        onPress={() => void select()}
      />
    </SSheet>
  );
}
