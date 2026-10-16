import { getMessaging, setBackgroundMessageHandler } from '@react-native-firebase/messaging';
import { handlePushMessage } from './push-service';

// Registered from index.ts, outside React, so it also runs in the headless background task.
setBackgroundMessageHandler(getMessaging(), async (message) => {
  try {
    await handlePushMessage(message);
  } catch (err) {
    console.warn('background push', err);
  }
});
