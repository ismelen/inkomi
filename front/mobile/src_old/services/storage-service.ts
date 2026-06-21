import AsyncStorage from '@react-native-async-storage/async-storage';
import * as SecureStorage from 'expo-secure-store';

export class StorageService {
  static async GetSecureAsync(key: string): Promise<string | undefined> {
    const result = await SecureStorage.getItemAsync(key);
    if (result === null) return undefined;
    return result;
  }

  static async SetSecureAsync(key: string, value: string): Promise<void> {
    await SecureStorage.setItemAsync(key, value);
  }

  static async GetAsync<T>(key: string): Promise<T | undefined> {
    const json = await AsyncStorage.getItem(key);
    if (json === null) return undefined;

    return JSON.parse(json);
  }

  static async SetAsync<T>(key: string, value: T): Promise<void> {
    const json = JSON.stringify(value);
    await AsyncStorage.setItem(key, json);
  }
}
