import { router, SplashScreen, Stack } from 'expo-router';
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
import { TRANSACTIONS_KEY, useQueue } from '../src/hooks/useQueue';
import { useCloud } from '../src/hooks/useCloud';
import { DropboxFolderPickerModal } from '../src/components/modals/dropbox-folder-picker-modal';
import NewVersionModal from '../src/components/modals/new-version-modal';
import { useSettings } from '../src/hooks/useSettings';
import { useMonitoredFolders } from '../src/hooks/useMonitoredFolders';
import {
  addNotificationReceivedListener,
  addNotificationResponseReceivedListener,
  AndroidNotificationPriority,
  NotificationBehavior,
  setNotificationHandler,
} from 'expo-notifications';
import { defineTask } from 'expo-task-manager';
import { StorageService } from '../src/services/storage-service';
import { useVersionChecker } from '../src/hooks/useVersionhecker';
import { useShallow } from 'zustand/react/shallow';
import { Transaction } from '../src/models/transaction';
import SConfirmDialog from '../src/components/shared/SConfirmDialog';
import { useTranslation } from 'react-i18next';

SplashScreen.preventAutoHideAsync();

setNotificationHandler({
  handleNotification: async () =>
    ({
      shouldShowAlert: true,
      shouldPlaySound: true,
      shouldSetBadge: false,
      priority: AndroidNotificationPriority.HIGH,
    }) as NotificationBehavior,
});

const BACKGROUND_NOTIFICATION_TASK = 'BACKGROUND-NOTIFICATION-TASK';

defineTask(BACKGROUND_NOTIFICATION_TASK, async ({ data, error }) => {
  if (error || !data) {
    return;
  }

  if (data) {
    const notiData = (data as any).notification?.request?.content?.data;
    await processNotification(notiData);
  }
});

async function processNotification(data: any): Promise<Transaction[] | undefined> {
  const raw = data['data'];

  const tran: Transaction = JSON.parse(raw);

  let transactions = (await StorageService.GetAsync<Transaction[]>(TRANSACTIONS_KEY)) ?? [tran];

  const idx = transactions.findIndex((e) => e.id !== tran.id);
  if (idx === -1) {
    transactions.unshift(tran);
  } else {
    transactions[idx] = tran;
  }

  await StorageService.SetAsync(TRANSACTIONS_KEY, transactions);
  return transactions;
}

export default function RootLayout() {
  const { t } = useTranslation();
  const initQueue = useQueue((s) => s.init);
  const initCloud = useCloud((s) => s.init);
  const initSettings = useSettings((s) => s.init);
  const initMonitoredFolders = useMonitoredFolders((s) => s.init);
  const initVersionChecker = useVersionChecker((s) => s.init);

  useEffect(() => {
    initVersionChecker();
    initMonitoredFolders();
    initSettings();
    initQueue();
    initCloud();
  }, []);

  useEffect(() => {
    const foregroundSubscription = addNotificationReceivedListener(async (notification) => {
      const data = notification.request.content.data;
      const transactions = await processNotification(data);
      if (!transactions) return;

      useQueue.setState({ transactions: transactions });
    });

    const responseSubscription = addNotificationResponseReceivedListener(() => {
      router.navigate('/(tabs)/queue');
    });

    return () => {
      foregroundSubscription.remove();
      responseSubscription.remove();
    };
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

  const showAuthError = useCloud((s) => s.showAuthError);

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
      <NewVersionModal />
      <SConfirmDialog
        visible={showAuthError}
        title={t('error.auth_failed_title', 'Authentication Error')}
        message={t(
          'error.auth_failed_message',
          'Your session has expired or the token is invalid. Please log in again to continue.'
        )}
        confirmText={t('common.ok', 'OK')}
        onConfirm={() => useCloud.getState().setShowAuthError(false)}
      />
    </GestureHandlerRootView>
  );
}
