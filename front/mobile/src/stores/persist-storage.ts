import AsyncStorage from '@react-native-async-storage/async-storage';
import { createJSONStorage } from 'zustand/middleware';

/** AsyncStorage for preferences and the list draft only; secrets go to SecureStore. */
export const asyncStorage = createJSONStorage(() => AsyncStorage);
