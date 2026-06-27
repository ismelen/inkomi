import * as Notifications from 'expo-notifications';
import { Platform } from 'react-native';

export class NotificationService {
  static async requestNotificationPermission(): Promise<boolean> {
    if (Platform.OS !== 'android') return true;

    const { status } = await Notifications.requestPermissionsAsync();
    return status === 'granted';
  }
}
