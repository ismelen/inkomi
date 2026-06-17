import * as IntentLauncher from 'expo-intent-launcher';
import * as Linking from 'expo-linking';
import { Platform } from 'react-native';

interface IStorageManager {
  isExternalStorageManager: () => Promise<boolean>;
  requestManageAllFiles: () => Promise<void>;
  getPackageName: () => Promise<string>;
}

const StorageManager: IStorageManager = {
  /**
   * Verifica si la app tiene el permiso MANAGE_EXTERNAL_STORAGE
   * Equivalente a Environment.isExternalStorageManager()
   */
  isExternalStorageManager: async (): Promise<boolean> => {
    if (Platform.OS !== 'android') {
      return false;
    }

    // Para Android 11+ (API 30+)
    if (Platform.Version >= 30) {
      try {
        // Intenta acceder a una ruta que solo está disponible con MANAGE_EXTERNAL_STORAGE
        const RNFS = await import('react-native-fs');

        // Si podemos listar el directorio raíz, tenemos el permiso
        try {
          await RNFS.default.readDir('/storage/emulated/0/');
          return true;
        } catch (error) {
          return false;
        }
      } catch {
        // Fallback: asumimos que no tenemos permiso
        return false;
      }
    }

    return true; // En versiones anteriores a Android 11, no se necesita
  },

  /**
   * Abre la configuración del sistema para "Administrar todos los archivos"
   * Equivalente a ACTION_MANAGE_ALL_FILES_ACCESS_PERMISSION
   */
  requestManageAllFiles: async (): Promise<void> => {
    if (Platform.OS !== 'android') {
      console.warn('Esta función solo está disponible en Android');
      return;
    }

    if (Platform.Version < 30) {
      console.warn('MANAGE_EXTERNAL_STORAGE solo está disponible en Android 11+');
      return;
    }

    try {
      const packageName = await StorageManager.getPackageName();
      
      // Usar IntentLauncher para abrir la configuración específica
      await IntentLauncher.startActivityAsync(
        IntentLauncher.ActivityAction.MANAGE_ALL_FILES_ACCESS_PERMISSION,
        {
          data: `package:${packageName}`,
        }
      );
    } catch (error) {
      console.error('Error abriendo configuración:', error);
      // Fallback: abrir configuración general de la app
      Linking.openSettings();
    }
  },

  /**
   * Obtiene el nombre del paquete de la app
   */
  getPackageName: async (): Promise<string> => {
    try {
      const Constants = await import('expo-constants');
      return Constants.default?.expoConfig?.android?.package || 'com.tuapp.nombre';
    } catch {
      return 'com.tuapp.nombre';
    }
  },
};

export default StorageManager;