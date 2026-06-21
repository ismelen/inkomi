import { Stack } from 'expo-router';
import { StatusBar } from 'expo-status-bar';
import React from 'react';
import { useWindowDimensions, View } from 'react-native';
import { GestureHandlerRootView } from 'react-native-gesture-handler';
import ConvertLoading from '../src/components/convert-settings/convert-loading';
import EReaderProfilePickerModalRoot from '../src/components/modals/e-reader-profile-picker-modal';
import DriveFolderPickerModalRoot from '../src/components/modals/google-drive-picker-modal';
import { useMonitoredFolders } from '../src/hooks/use-monitored-folders';
import { colors } from '../src/theme/colors';

export default function RootLayout() {
  const loading = useMonitoredFolders((s) => s.loading);
  const { width, height } = useWindowDimensions();

  return (
    <GestureHandlerRootView style={{ flex: 1 }}>
      <StatusBar style="dark" backgroundColor={colors.background} />
      <Stack
        screenOptions={{
          headerShown: false,
          headerStyle: {
            backgroundColor: colors.background,
          },
          headerShadowVisible: false,
          contentStyle: {
            backgroundColor: colors.background,
          },
        }}
      />
      <DriveFolderPickerModalRoot />
      <EReaderProfilePickerModalRoot />
      {loading && (
        <View style={{ width: width, height: height, position: 'absolute' }}>
          <ConvertLoading />
        </View>
      )}
    </GestureHandlerRootView>
  );
}
