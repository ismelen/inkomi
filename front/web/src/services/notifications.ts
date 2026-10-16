import i18n from '../i18n/i18n';
import { useSettings } from '../stores/use-settings';

export function notificationsSupported(): boolean {
  return typeof Notification !== 'undefined';
}

/** Asks for the permission if needed. False when the browser blocks or lacks notifications. */
export async function requestNotifications(): Promise<boolean> {
  if (!notificationsSupported()) return false;
  if (Notification.permission === 'granted') return true;
  if (Notification.permission === 'denied') return false;
  return (await Notification.requestPermission()) === 'granted';
}

/** Shown by the tab that owns the websocket, only while the user is looking elsewhere. */
export function notifySent(title: string) {
  if (!useSettings.getState().notifications || !notificationsSupported()) return;
  if (Notification.permission !== 'granted' || document.hasFocus()) return;
  new Notification(i18n.t('notifications.sentTitle'), {
    body: i18n.t('notifications.sentBody', { title }),
    icon: '/icon.png',
    tag: `sent-${title}`,
  });
}
