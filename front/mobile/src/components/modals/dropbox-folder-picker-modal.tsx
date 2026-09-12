import React, { useEffect, useRef, useState } from 'react';
import {
  Animated,
  Dimensions,
  FlatList,
  Modal,
  Pressable,
  StyleSheet,
  TouchableOpacity,
  View,
  ActivityIndicator,
} from 'react-native';
import { useSafeAreaInsets } from 'react-native-safe-area-context';
import { useCloud } from '../../hooks/useCloud';
import { usePathname } from 'expo-router';
import { colors } from '../../theme/colors';
import SText from '../shared/SText';
import SButton from '../shared/SButton';
import SIcon from '../icons/SIcon';

const SCREEN_HEIGHT = Dimensions.get('window').height;

interface DropboxFolder {
  id: string;
  name: string;
  path_lower: string;
  path_display: string;
}

export function DropboxFolderPickerModal() {
  const insets = useSafeAreaInsets();
  const DRAWER_HEIGHT = SCREEN_HEIGHT * 0.85 - insets.top;
  const pathname = usePathname();
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
      const token = await getToken(pathname);
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
      <Animated.View style={[styles.backdrop, { opacity: backdropOpacity }]}>
        <Pressable style={StyleSheet.absoluteFill} onPress={close} />
      </Animated.View>

      <Animated.View
        style={[styles.drawer, { height: DRAWER_HEIGHT, transform: [{ translateY }] }]}
      >
        <View style={styles.handle} />

        <View style={styles.header}>
          <SText style={styles.title}>Seleccionar carpeta</SText>
          <TouchableOpacity onPress={close} hitSlop={12} style={{ padding: 5 }}>
            <SIcon name="close" size={24} color={colors.on_surface_variant} />
          </TouchableOpacity>
        </View>

        <View style={styles.breadcrumbRow}>
          <TouchableOpacity onPress={() => navigateTo('', '')}>
            <SText style={[styles.breadcrumb, currentPath === '' && styles.breadcrumbActive]}>
              Dropbox
            </SText>
          </TouchableOpacity>
          {breadcrumbs.map((b, i) => (
            <React.Fragment key={b.path}>
              <SText style={styles.breadcrumbSep}>›</SText>
              <TouchableOpacity onPress={() => navigateTo(b.path, b.name)}>
                <SText
                  style={[
                    styles.breadcrumb,
                    i === breadcrumbs.length - 1 && styles.breadcrumbActive,
                  ]}
                >
                  {b.name}
                </SText>
              </TouchableOpacity>
            </React.Fragment>
          ))}
        </View>

        <View style={{ flex: 1 }}>
          {loading ? (
            <ActivityIndicator style={styles.loader} color={colors.primary} size="large" />
          ) : error ? (
            <View style={styles.errorContainer}>
              <SText style={styles.errorText}>{error}</SText>
              <SButton onPress={() => navigateTo(currentPath, '')} style={styles.retryBtn}>
                <SText style={{ color: colors.on_primary, fontFamily: 'semibold' }}>
                  Reintentar
                </SText>
              </SButton>
            </View>
          ) : (
            <FlatList
              data={folders}
              keyExtractor={(item) => item.id}
              contentContainerStyle={styles.listContent}
              ListEmptyComponent={<SText style={styles.emptyText}>Esta carpeta está vacía</SText>}
              renderItem={({ item }) => (
                <TouchableOpacity
                  style={styles.folderRow}
                  onPress={() => navigateTo(item.path_lower, item.name)}
                >
                  <SIcon name="folder" size={28} color={colors.primary} />
                  <SText style={styles.folderName} numberOfLines={1}>
                    {item.name}
                  </SText>
                  <SIcon
                    name="chevron_right"
                    size={24}
                    color={colors.outline_variant}
                    type="outlined"
                  />
                </TouchableOpacity>
              )}
            />
          )}
        </View>

        <View style={[styles.bottomBar, { paddingBottom: 20 + insets.bottom }]}>
          <SButton
            onPress={handleSelectCurrent}
            style={[
              styles.selectBtn,
              currentPath === '' && { backgroundColor: colors.surface_variant },
            ]}
            disabled={currentPath === ''}
          >
            <SText
              style={[
                styles.selectBtnText,
                currentPath === '' && { color: colors.on_surface_variant },
              ]}
            >
              {currentPath
                ? `Usar "${breadcrumbs.at(-1)?.name ?? currentPath}"`
                : 'No se puede usar la raíz'}
            </SText>
            <SIcon
              name="check"
              size={20}
              color={currentPath === '' ? colors.on_surface_variant : colors.on_primary}
              type="outlined"
            />
          </SButton>
        </View>
      </Animated.View>
    </Modal>
  );
}

const styles = StyleSheet.create({
  backdrop: {
    ...StyleSheet.absoluteFillObject,
    backgroundColor: 'rgba(0,0,0,0.5)',
  },
  drawer: {
    position: 'absolute',
    bottom: 0,
    left: 0,
    right: 0,
    backgroundColor: colors.surface_container_lowest,
    borderTopLeftRadius: 24,
    borderTopRightRadius: 24,
  },
  handle: {
    width: 40,
    height: 5,
    borderRadius: 3,
    backgroundColor: colors.outline_variant,
    alignSelf: 'center',
    marginTop: 12,
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
    fontSize: 18,
    fontFamily: 'bold',
    color: colors.on_surface,
  },
  breadcrumbRow: {
    flexDirection: 'row',
    alignItems: 'center',
    flexWrap: 'wrap',
    paddingHorizontal: 20,
    paddingBottom: 16,
    gap: 4,
    borderBottomWidth: 1,
    borderBottomColor: colors.surface_container,
  },
  breadcrumb: {
    fontSize: 14,
    color: colors.primary,
    fontFamily: 'semibold',
  },
  breadcrumbActive: {
    color: colors.on_surface_variant,
    fontFamily: 'bold',
  },
  breadcrumbSep: {
    fontSize: 14,
    color: colors.outline_variant,
    marginHorizontal: 4,
  },
  loader: {
    marginTop: 40,
  },
  errorContainer: {
    alignItems: 'center',
    marginTop: 40,
    gap: 16,
  },
  errorText: {
    fontSize: 15,
    color: colors.error,
    textAlign: 'center',
    paddingHorizontal: 24,
    fontFamily: 'semibold',
  },
  retryBtn: {
    backgroundColor: colors.primary,
    paddingHorizontal: 20,
    paddingVertical: 10,
    borderRadius: 12,
  },
  listContent: {
    paddingBottom: 20,
  },
  folderRow: {
    flexDirection: 'row',
    alignItems: 'center',
    paddingVertical: 16,
    paddingHorizontal: 20,
    borderBottomWidth: StyleSheet.hairlineWidth,
    borderColor: colors.surface_container_high,
    gap: 16,
  },
  folderName: {
    flex: 1,
    fontSize: 16,
    color: colors.on_surface,
    fontFamily: 'medium',
  },
  emptyText: {
    textAlign: 'center',
    color: colors.outline,
    fontSize: 15,
    marginTop: 40,
  },
  bottomBar: {
    padding: 20,
    borderTopWidth: 1,
    borderTopColor: colors.surface_container,
    backgroundColor: colors.surface_container_lowest,
  },
  selectBtn: {
    backgroundColor: colors.primary,
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'center',
    paddingVertical: 14,
    borderRadius: 14,
    gap: 8,
    boxShadow: colors.boxShadow,
  },
  selectBtnText: {
    color: colors.on_primary,
    fontSize: 16,
    fontFamily: 'bold',
  },
});
