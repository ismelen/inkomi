import { requireNativeModule } from 'expo-modules-core';

const FileHandler = requireNativeModule('FileHandler');

export async function copyToCache(contentUri: string, filename: string): Promise<string> {
  return await FileHandler.copyToCache(contentUri, filename);
}
