import React, { useCallback, useEffect, useState } from 'react';
import {
  ActivityIndicator,
  Pressable,
  StyleSheet,
  Text,
  TouchableOpacity,
  useWindowDimensions,
  View,
} from 'react-native';
import { create } from 'zustand';
import { colors } from '../../theme/colors';
import ChevronRightIcon from '../icons/chevron-right-icon';
import CloseIcon from '../icons/close-icon';
import FolderIcon from '../icons/folder-icon';
import SColumn from '../shared/SColumn';
import SText from '../shared/SText';

interface State {
  token?: string;
  resolver?: (value?: { id: string; name: string }) => void;
  close(folderId: string, folderName: string): void;
  show(token: string): Promise<{ id: string; name: string } | undefined>;
}

export const useGoogleDrivePicker = create<State>((set, get) => ({
  show(token: string): Promise<{ id: string; name: string } | undefined> {
    return new Promise<{ id: string; name: string } | undefined>((resolve) => {
      set({ token: token, resolver: resolve });
    });
  },

  close(folderId: string, folderName: string) {
    const { resolver } = get();
    set({ token: undefined, resolver: undefined });

    if (!folderId || !folderName) {
      resolver?.(undefined);
      return;
    }

    resolver?.({ id: folderId, name: folderName });
  },
}));

export default function DriveFolderPickerModalRoot() {
  const token = useGoogleDrivePicker((s) => s.token);
  const close = useGoogleDrivePicker((s) => s.close);
  const { width, height } = useWindowDimensions();

  if (!token) return null;

  return (
    <View
      style={{
        flex: 1,
        paddingTop: 50,
        width: width,
        height: height,
        bottom: 0,
        position: 'absolute',
      }}
    >
      <View
        style={{
          backgroundColor: 'black',
          opacity: 0.4,
          width: width,
          height: height,
          position: 'absolute',
        }}
      />
      <View
        style={{
          backgroundColor: colors.card,
          borderTopLeftRadius: 14,
          borderTopRightRadius: 14,
          width: width,
          // height: height / 2,
          position: 'absolute',
          bottom: 0,
          paddingBottom: 30,
        }}
      >
        <View
          style={{
            flexDirection: 'row',
            alignItems: 'center',
            paddingHorizontal: 20,
            paddingVertical: 14,
            justifyContent: 'space-between',
          }}
        >
          <SText style={{ fontWeight: 600, fontSize: 20 }}>Cloud Folders</SText>
          <Pressable onPress={() => close('', '')}>
            <CloseIcon size="28px" color={colors.onCard} />
          </Pressable>
        </View>

        <DriveFolderList token={token} onSelect={(id, name) => close(id, name)} />
      </View>
    </View>
  );
}

export interface GoogleDriveFile {
  id: string;
  name: string;
  mimeType: string;
}

export interface GoogleDriveResponse {
  files: GoogleDriveFile[];
  nextPageToken?: string;
}
interface Props {
  token: string;
  onSelect: (folderId: string, folderName: string) => void;
}

function DriveFolderList({ token, onSelect }: Props) {
  const [folders, setFolders] = useState<GoogleDriveFile[]>([]);
  const [loading, setLoading] = useState<boolean>(true);
  const [error, setError] = useState<string | null>(null);

  const fetchFolders = useCallback(async () => {
    try {
      setLoading(true);
      const query = encodeURIComponent(
        "mimeType='application/vnd.google-apps.folder' and trashed=false and 'me' in owners"
      );

      const response = await fetch(
        `https://www.googleapis.com/drive/v3/files?q=${query}&fields=files(id,name,mimeType)&orderBy=name`,
        {
          method: 'GET',
          headers: {
            Authorization: `Bearer ${token}`,
            'Content-Type': 'application/json',
          },
        }
      );

      if (!response.ok) throw new Error('Error al conectar con Google Drive');

      const data: GoogleDriveResponse = await response.json();
      setFolders(data.files);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Error desconocido');
    } finally {
      setLoading(false);
    }
  }, [token]);

  useEffect(() => {
    fetchFolders();
  }, [fetchFolders]);

  if (loading) {
    return (
      <View style={styles.center}>
        <ActivityIndicator size="large" color="#4285F4" />
        <Text style={styles.loadingText}>Cargando carpetas...</Text>
      </View>
    );
  }

  if (error) {
    return (
      <View style={styles.center}>
        <Text style={styles.errorText}>{error}</Text>
        <TouchableOpacity onPress={fetchFolders} style={styles.retryButton}>
          <Text style={styles.retryText}>Reintentar</Text>
        </TouchableOpacity>
      </View>
    );
  }

  return (
    <SColumn>
      {folders.map((e, i) => (
        <Pressable
          style={{ flexDirection: 'row', alignItems: 'center', gap: 10 }}
          key={i}
          onPress={() => onSelect(e.id, e.name)}
        >
          <FolderIcon size="24px" color={colors.onCard} />
          <SText style={{ flex: 1, fontWeight: 500, fontSize: 16 }}>{e.name}</SText>
          <ChevronRightIcon size="24px" color={colors.onCard} />
        </Pressable>
      ))}
    </SColumn>
  );
}

const styles = StyleSheet.create({
  center: { flex: 1, justifyContent: 'center', alignItems: 'center', padding: 20 },
  listContent: { paddingBottom: 20 },
  loadingText: { marginTop: 10, color: '#666' },
  errorText: { color: 'red', textAlign: 'center', marginBottom: 10 },
  retryButton: { padding: 10, backgroundColor: '#4285F4', borderRadius: 5 },
  retryText: { color: 'white', fontWeight: 'bold' },
  folderItem: {
    flexDirection: 'row',
    alignItems: 'center',
    padding: 15,
    borderBottomWidth: 1,
    borderBottomColor: '#eee',
  },
  folderIcon: { fontSize: 24, marginRight: 15 },
  folderName: { fontSize: 16, color: '#333', flex: 1 },
  emptyText: { textAlign: 'center', marginTop: 50, color: '#999' },
});
