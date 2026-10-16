import React, { useEffect } from 'react';
import { Stack } from 'expo-router';
import { colors } from '../../src/theme';
import { SendingModal } from '../../src/components/upload/sending-modal';
import { disablePushListeners, enablePush } from '../../src/services/push-service';
import { startSync, stopSync } from '../../src/services/sync-service';
import { useActivity } from '../../src/stores/use-activity';
import { useProfile } from '../../src/stores/use-profile';
import { useSettings } from '../../src/stores/use-settings';

/** Mounted only with a session: realtime activity, push and profile live here. */
export default function AppLayout() {
  useEffect(() => {
    void useActivity.getState().reload();
    startSync();
    useProfile
      .getState()
      .load()
      .catch((err) => console.warn('profile', err));
    if (useSettings.getState().notifications) {
      enablePush().catch((err) => console.warn('push', err));
    }
    return () => {
      stopSync();
      disablePushListeners();
    };
  }, []);

  return (
    <>
      <Stack
        screenOptions={{ headerShown: false, contentStyle: { backgroundColor: colors.background } }}
      />
      <SendingModal />
    </>
  );
}
