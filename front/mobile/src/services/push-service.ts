import { Platform } from 'react-native';
import * as Notifications from 'expo-notifications';
import {
  FirebaseMessagingTypes,
  getMessaging,
  getToken,
  onMessage,
  onTokenRefresh,
} from '@react-native-firebase/messaging';
import { usersApi } from '../api/users-api';
import { SentPushData } from '../api/types/events';
import { handleSentPush } from './sync-service';

let unsubscribe: (() => void)[] = [];

function sentData(message: FirebaseMessagingTypes.RemoteMessage): SentPushData | null {
  const data = message.data as Partial<SentPushData> | undefined;
  return data?.type === 'sent' && data.id ? (data as SentPushData) : null;
}

/** Applies a "sent" push to the local activity, in foreground or from the background handler. */
export async function handlePushMessage(message: FirebaseMessagingTypes.RemoteMessage) {
  const data = sentData(message);
  if (data) await handleSentPush(data);
}

async function hasPermission(): Promise<boolean> {
  const current = await Notifications.getPermissionsAsync();
  if (current.granted) return true;
  if (Platform.OS === 'android' && Platform.Version < 33) return true;
  const asked = await Notifications.requestPermissionsAsync();
  return asked.granted;
}

/** Sends the FCM token to the auth-service and keeps it updated. False if not allowed. */
export async function enablePush(): Promise<boolean> {
  disablePushListeners();
  if (!(await hasPermission())) return false;

  const messaging = getMessaging();
  await usersApi.setFcmToken(await getToken(messaging));
  unsubscribe = [
    onTokenRefresh(messaging, (token) => {
      usersApi.setFcmToken(token).catch((err) => console.warn('fcm token refresh', err));
    }),
    onMessage(messaging, (message) => handlePushMessage(message)),
  ];
  return true;
}

export function disablePushListeners() {
  unsubscribe.forEach((fn) => fn());
  unsubscribe = [];
}

/** Stops the notifications of this user: an empty token deletes it in the auth-service. */
export async function disablePush(): Promise<void> {
  disablePushListeners();
  await usersApi.setFcmToken('');
}
