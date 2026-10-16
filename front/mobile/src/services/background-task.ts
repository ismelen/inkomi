import { Platform } from 'react-native';
import BackgroundService from 'react-native-background-actions';

/**
 * Runs task inside an Android foreground service so uploads keep going with the screen off.
 * On iOS it just runs the task.
 */
export async function runInBackground<T>(title: string, desc: string, task: () => Promise<T>) {
  if (Platform.OS !== 'android') return task();

  let result: T | undefined;
  let error: unknown;
  let finish!: () => void;
  const finished = new Promise<void>((resolve) => (finish = resolve));

  await BackgroundService.start(
    async () => {
      try {
        result = await task();
      } catch (err) {
        error = err;
      } finally {
        finish();
      }
    },
    {
      taskName: 'inkomi-upload',
      taskTitle: title,
      taskDesc: desc,
      taskIcon: { name: 'ic_launcher', type: 'mipmap' },
      foregroundServiceType: ['dataSync'],
    } as Parameters<typeof BackgroundService.start>[1]
  );
  await finished;
  await BackgroundService.stop();

  if (error) throw error;
  return result as T;
}
