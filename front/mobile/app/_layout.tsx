import '../src/i18n/i18n';
import React, { useEffect } from 'react';
import { AppState } from 'react-native';
import { SplashScreen, Stack } from 'expo-router';
import { StatusBar } from 'expo-status-bar';
import { GestureHandlerRootView } from 'react-native-gesture-handler';
import {
  Inter_400Regular,
  Inter_500Medium,
  Inter_600SemiBold,
  Inter_700Bold,
  Inter_900Black,
  useFonts,
} from '@expo-google-fonts/inter';
import { UpdateModal } from '../src/components/update/update-modal';
import { useAppUpdate } from '../src/stores/use-app-update';
import { colors } from '../src/theme';
import { useSession } from '../src/stores/use-session';

void SplashScreen.preventAutoHideAsync();

const updateCheckEveryMs = 6 * 60 * 60_000;

export default function RootLayout() {
  const status = useSession((s) => s.status);
  const [fontsLoaded] = useFonts({
    regular: Inter_400Regular,
    medium: Inter_500Medium,
    semibold: Inter_600SemiBold,
    bold: Inter_700Bold,
    black: Inter_900Black,
  });

  useEffect(() => {
    void useSession.getState().hydrate();
    void useAppUpdate.getState().check();
    const sub = AppState.addEventListener('change', (state) => {
      const { lastCheck, check } = useAppUpdate.getState();
      if (state === 'active' && Date.now() - lastCheck > updateCheckEveryMs) void check();
    });
    return () => sub.remove();
  }, []);

  const ready = fontsLoaded && status !== 'loading';
  useEffect(() => {
    if (ready) void SplashScreen.hideAsync();
  }, [ready]);

  if (!ready) return null;

  const signedIn = status === 'signedIn';
  return (
    <GestureHandlerRootView style={{ flex: 1, backgroundColor: colors.background }}>
      <StatusBar style="dark" />
      <Stack
        screenOptions={{ headerShown: false, contentStyle: { backgroundColor: colors.background } }}
      >
        <Stack.Protected guard={signedIn}>
          <Stack.Screen name="(app)" />
        </Stack.Protected>
        <Stack.Protected guard={!signedIn}>
          <Stack.Screen name="(auth)" />
        </Stack.Protected>
        <Stack.Screen name="oauth" />
      </Stack>
      <UpdateModal />
    </GestureHandlerRootView>
  );
}
