import { SplashScreen, Stack } from 'expo-router';
import React, { useCallback, useEffect } from 'react';
import { GestureHandlerRootView } from 'react-native-gesture-handler';
import { StatusBar } from 'expo-status-bar';
import { colors } from '../src/theme/colors';
import {
  useFonts,
  Inter_400Regular,
  Inter_500Medium,
  Inter_600SemiBold,
  Inter_700Bold,
  Inter_900Black,
} from '@expo-google-fonts/inter';
import { useQueue } from '../src/hooks/useQueue';
import { useCloud } from '../src/hooks/useCloud';
import { DropboxFolderPickerModal } from '../src/components/modals/dropbox-folder-picker-modal';

SplashScreen.preventAutoHideAsync();

export default function RootLayout() {
  const initQueue = useQueue((s) => s.init);
  const initCloud = useCloud((s) => s.init);

  useEffect(() => {
    initQueue();
    initCloud();
  }, []);

  const [fontsLoaded] = useFonts({
    regular: Inter_400Regular,
    medium: Inter_500Medium,
    semibold: Inter_600SemiBold,
    bold: Inter_700Bold,
    black: Inter_900Black,
  });

  const onLayoutRootView = useCallback(async () => {
    if (fontsLoaded) {
      await SplashScreen.hideAsync();
    }
  }, [fontsLoaded]);

  if (!fontsLoaded) {
    return null;
  }

  return (
    <GestureHandlerRootView style={{ flex: 1 }} onLayout={onLayoutRootView}>
      <Stack
        screenOptions={{
          headerShown: false,
          headerShadowVisible: false,
          contentStyle: {
            backgroundColor: colors.background,
          },
        }}
      />
      <DropboxFolderPickerModal />
    </GestureHandlerRootView>
  );
}
