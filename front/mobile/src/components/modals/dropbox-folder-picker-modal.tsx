// components/DropboxFolderPicker.tsx
import React, { useEffect, useRef, useState } from 'react';
import {
  Animated,
  Dimensions,
  FlatList,
  Modal,
  Pressable,
  StyleSheet,
  Text,
  TouchableOpacity,
  View,
  ActivityIndicator,
} from 'react-native';
import { useCloud } from '../../hooks/useCloud';

const SCREEN_HEIGHT = Dimensions.get('window').height;
const DRAWER_HEIGHT = SCREEN_HEIGHT * 0.75;

interface DropboxFolder {
  id: string;
  name: string;
  path_lower: string;
  path_display: string;
}

export function DropboxFolderPickerModal() {
  const showDialog = useCloud((s) => s.showDialog);
  const onFolderSelect = useCloud((s) => s.onFolderSelect);
  const getToken = useCloud((s) => s.getToken);

  const [folders, setFolders] = useState<DropboxFolder[]>([]);
  const [currentPath, setCurrentPath] = useState('');
  const [breadcrumbs, setBreadcrumbs] = useState<{ name: string; path: string }[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const translateY = useRef(new Animated.Value(DRAWER_HEIGHT)).current;
  const backdropOpacity = useRef(new Animated.Value(0)).current;

  useEffect(() => {
    if (showDialog) {
      navigateTo('', '');
      Animated.parallel([
        Animated.spring(translateY, {
          toValue: 0,
          useNativeDriver: true,
          bounciness: 0,
          speed: 14,
        }),
        Animated.timing(backdropOpacity, {
          toValue: 1,
          duration: 250,
          useNativeDriver: true,
        }),
      ]).start();
    } else {
      Animated.parallel([
        Animated.timing(translateY, {
          toValue: DRAWER_HEIGHT,
          duration: 220,
          useNativeDriver: true,
        }),
        Animated.timing(backdropOpacity, {
          toValue: 0,
          duration: 220,
          useNativeDriver: true,
        }),
      ]).start();
    }
  }, [showDialog]);

  const close = () => {
    onFolderSelect?.();
    setCurrentPath('');
    setBreadcrumbs([]);
    setFolders([]);
    setError(null);
  };

  const navigateTo = async (path: string, name: string) => {
    setLoading(true);
    setError(null);
    try {
      const token = await getToken();
      if (!token) throw new Error('No hay sesión de Dropbox activa');

      const res = await fetch('https://api.dropboxapi.com/2/files/list_folder', {
        method: 'POST',
        headers: {
          Authorization: `Bearer ${token}`,
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          path,
          recursive: false,
          include_non_downloadable_files: false,
        }),
      });

      if (!res.ok) throw new Error('Error al cargar carpetas');

      const data = await res.json();
      const onlyFolders: DropboxFolder[] = (data.entries ?? []).filter(
        (e: any) => e['.tag'] === 'folder'
      );

      setFolders(onlyFolders);
      setCurrentPath(path);

      // Actualiza breadcrumbs
      if (path === '') {
        setBreadcrumbs([]);
      } else {
        setBreadcrumbs((prev) => {
          const idx = prev.findIndex((b) => b.path === path);
          if (idx >= 0) return prev.slice(0, idx + 1);
          return [...prev, { name, path }];
        });
      }
    } catch (e: any) {
      setError(e.message ?? 'Error desconocido');
    } finally {
      setLoading(false);
    }
  };

  const handleSelectCurrent = () => {
    onFolderSelect?.(currentPath);
    close();
  };

  return (
    <Modal visible={showDialog} transparent animationType="none" onRequestClose={close}>
      {/* Backdrop */}
      <Animated.View style={[styles.backdrop, { opacity: backdropOpacity }]}>
        <Pressable style={StyleSheet.absoluteFill} onPress={close} />
      </Animated.View>

      {/* Drawer */}
      <Animated.View style={[styles.drawer, { transform: [{ translateY }] }]}>
        {/* Handle */}
        <View style={styles.handle} />

        {/* Header */}
        <View style={styles.header}>
          <Text style={styles.title}>Seleccionar carpeta</Text>
          <TouchableOpacity onPress={close} hitSlop={12}>
            <Text style={styles.closeBtn}>✕</Text>
          </TouchableOpacity>
        </View>

        {/* Breadcrumbs */}
        <View style={styles.breadcrumbRow}>
          <TouchableOpacity onPress={() => navigateTo('', '')}>
            <Text style={[styles.breadcrumb, currentPath === '' && styles.breadcrumbActive]}>
              Dropbox
            </Text>
          </TouchableOpacity>
          {breadcrumbs.map((b, i) => (
            <React.Fragment key={b.path}>
              <Text style={styles.breadcrumbSep}>›</Text>
              <TouchableOpacity onPress={() => navigateTo(b.path, b.name)}>
                <Text
                  style={[
                    styles.breadcrumb,
                    i === breadcrumbs.length - 1 && styles.breadcrumbActive,
                  ]}
                >
                  {b.name}
                </Text>
              </TouchableOpacity>
            </React.Fragment>
          ))}
        </View>

        {/* Botón seleccionar carpeta actual */}
        <TouchableOpacity style={styles.selectCurrentBtn} onPress={handleSelectCurrent}>
          <Text style={styles.selectCurrentIcon}>✓</Text>
          <Text style={styles.selectCurrentText}>
            {currentPath
              ? `Seleccionar "${breadcrumbs.at(-1)?.name ?? currentPath}"`
              : 'Seleccionar raíz de Dropbox'}
          </Text>
        </TouchableOpacity>

        {/* Lista de carpetas */}
        {loading ? (
          <ActivityIndicator style={styles.loader} color="#0061FE" />
        ) : error ? (
          <View style={styles.errorContainer}>
            <Text style={styles.errorText}>{error}</Text>
            <TouchableOpacity onPress={() => navigateTo(currentPath, '')}>
              <Text style={styles.retryText}>Reintentar</Text>
            </TouchableOpacity>
          </View>
        ) : (
          <FlatList
            data={folders}
            keyExtractor={(item) => item.id}
            contentContainerStyle={styles.listContent}
            ListEmptyComponent={<Text style={styles.emptyText}>Esta carpeta está vacía</Text>}
            renderItem={({ item }) => (
              <TouchableOpacity
                style={styles.folderRow}
                onPress={() => navigateTo(item.path_lower, item.name)}
              >
                <Text style={styles.folderIcon}>📁</Text>
                <Text style={styles.folderName} numberOfLines={1}>
                  {item.name}
                </Text>
                <Text style={styles.folderArrow}>›</Text>
              </TouchableOpacity>
            )}
          />
        )}
      </Animated.View>
    </Modal>
  );
}

const styles = StyleSheet.create({
  backdrop: {
    ...StyleSheet.absoluteFillObject,
    backgroundColor: 'rgba(0,0,0,0.45)',
  },
  drawer: {
    position: 'absolute',
    bottom: 0,
    left: 0,
    right: 0,
    height: DRAWER_HEIGHT,
    backgroundColor: '#fff',
    borderTopLeftRadius: 20,
    borderTopRightRadius: 20,
  },
  handle: {
    width: 36,
    height: 4,
    borderRadius: 2,
    backgroundColor: '#D1D1D6',
    alignSelf: 'center',
    marginTop: 10,
  },
  header: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'space-between',
    paddingHorizontal: 20,
    paddingTop: 16,
    paddingBottom: 12,
  },
  title: {
    fontSize: 17,
    fontWeight: '600',
    color: '#111',
  },
  closeBtn: {
    fontSize: 16,
    color: '#999',
    fontWeight: '500',
  },
  breadcrumbRow: {
    flexDirection: 'row',
    alignItems: 'center',
    flexWrap: 'wrap',
    paddingHorizontal: 20,
    paddingBottom: 12,
    gap: 4,
  },
  breadcrumb: {
    fontSize: 13,
    color: '#0061FE',
  },
  breadcrumbActive: {
    color: '#555',
    fontWeight: '500',
  },
  breadcrumbSep: {
    fontSize: 13,
    color: '#C7C7CC',
    marginHorizontal: 2,
  },
  selectCurrentBtn: {
    flexDirection: 'row',
    alignItems: 'center',
    marginHorizontal: 16,
    marginBottom: 8,
    paddingVertical: 11,
    paddingHorizontal: 14,
    backgroundColor: '#EEF4FF',
    borderRadius: 10,
    borderWidth: 1,
    borderColor: '#0061FE',
    gap: 8,
  },
  selectCurrentIcon: {
    fontSize: 15,
    color: '#0061FE',
    fontWeight: '700',
  },
  selectCurrentText: {
    fontSize: 14,
    color: '#0061FE',
    fontWeight: '500',
    flexShrink: 1,
  },
  loader: {
    marginTop: 40,
  },
  errorContainer: {
    alignItems: 'center',
    marginTop: 40,
    gap: 10,
  },
  errorText: {
    fontSize: 14,
    color: '#E24B4A',
    textAlign: 'center',
    paddingHorizontal: 24,
  },
  retryText: {
    fontSize: 14,
    color: '#0061FE',
    fontWeight: '500',
  },
  listContent: {
    paddingBottom: 40,
  },
  folderRow: {
    flexDirection: 'row',
    alignItems: 'center',
    paddingVertical: 14,
    paddingHorizontal: 20,
    borderBottomWidth: StyleSheet.hairlineWidth,
    borderColor: '#F0F0F5',
    gap: 12,
  },
  folderIcon: {
    fontSize: 20,
  },
  folderName: {
    flex: 1,
    fontSize: 15,
    color: '#111',
  },
  folderArrow: {
    fontSize: 20,
    color: '#C7C7CC',
  },
  emptyText: {
    textAlign: 'center',
    color: '#999',
    fontSize: 14,
    marginTop: 40,
  },
});
