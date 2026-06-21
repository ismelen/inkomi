import { router, Stack } from 'expo-router';
import React, { useEffect } from 'react';
import { Pressable, StyleSheet, View } from 'react-native';
import ActionButton from '../../src/components/dashboard/action-button';
import MonitoredFolderCard from '../../src/components/dashboard/monitored-folder-card';
import AddCircleIcon from '../../src/components/icons/add-cirlce-icon';
import BookIcon from '../../src/components/icons/book-icon';
import ImageIcon from '../../src/components/icons/image-icon';
import SColumn from '../../src/components/shared/SColumn';
import SDivider from '../../src/components/shared/SDivider';
import SText from '../../src/components/shared/SText';
import { useMonitoredFolders } from '../../src/hooks/use-monitored-folders';
import { colors } from '../../src/theme/colors';

export default function index() {
  const fetchMonitoredFolders = useMonitoredFolders((s) => s.fetchMonitoredFolders);
  const folders = useMonitoredFolders((s) => s.folders);
  const addFolder = useMonitoredFolders((s) => s.addFolder);

  useEffect(() => {
    fetchMonitoredFolders();
  }, []);

  return (
    <View>
      <Stack.Screen
        options={{
          headerShown: false,
          contentStyle: {
            backgroundColor: colors.background,
            paddingTop: 50,
          },
        }}
      />
      <View style={[styles.padding]}>
        <SText style={[styles.title, { fontSize: 20 }]}>Dashboard</SText>
        <SText style={{ fontSize: 14, color: colors.onCard }}>CONVERSION HUB</SText>
      </View>
      <SDivider />
      <View style={[styles.padding, { gap: 14, paddingTop: 20 }]}>
        <ActionButton
          text="Convert Manga to EPUB"
          icon={(size, color) => <ImageIcon size={size} color={color} />}
          onPress={() =>
            router.push({
              pathname: '/convert-settings/[idx]/[kepubify]',
              params: {
                idx: -1,
              },
            })
          }
        />
        <ActionButton
          text="Kepubify"
          icon={(size, color) => <BookIcon size={size} color={color} />}
          onPress={() =>
            router.push({
              pathname: '/convert-settings/[idx]/[kepubify]',
              params: {
                idx: -1,
                kepubify: true.toString(),
              },
            })
          }
        />
        <View style={[styles.monitoredTitle]}>
          <SText style={[styles.title, { fontSize: 18 }]}>FOLDER MONITOR</SText>
          <View style={{ flexDirection: 'row', gap: 10, alignItems: 'center' }}>
            <View
              style={{ backgroundColor: colors.onCard, width: 8, height: 8, borderRadius: 10 }}
            />
            <SText style={{ color: colors.onCard }}>LIVE</SText>
          </View>
        </View>

        <SColumn
          footer={
            <Pressable style={[styles.addFolders]} onPress={addFolder}>
              <AddCircleIcon size="22px" color={colors.primary} />
              <SText style={{ color: colors.primary, fontWeight: '600', fontSize: 16 }}>
                Add Folder
              </SText>
            </Pressable>
          }
        >
          {folders.map((e, i) => (
            <MonitoredFolderCard
              key={i}
              folder={e}
              onPress={() =>
                router.push({
                  pathname: '/convert-settings/[idx]/[kepubify]',
                  params: {
                    idx: i,
                  },
                })
              }
            />
          ))}
        </SColumn>
      </View>
    </View>
  );
}

const styles = StyleSheet.create({
  padding: {
    paddingHorizontal: 20,
    paddingVertical: 10,
  },
  title: {
    fontWeight: '500',
  },
  monitoredTitle: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    marginTop: 15,
  },
  addFolders: {
    flexDirection: 'row',
    gap: 8,
    alignItems: 'center',
  },
});
