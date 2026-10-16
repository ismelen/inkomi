import i18n from 'i18next';
import { initReactI18next } from 'react-i18next';
import * as Localization from 'expo-localization';
import en from './locales/en.json';
import es from './locales/es.json';

export const supportedLanguages = ['es', 'en'] as const;
export type SupportedLanguage = (typeof supportedLanguages)[number];

const deviceLanguage = Localization.getLocales()[0]?.languageCode ?? 'es';
const initialLanguage: SupportedLanguage = supportedLanguages.includes(
  deviceLanguage as SupportedLanguage
)
  ? (deviceLanguage as SupportedLanguage)
  : 'es';

// Spanish is the base language: it is the fallback for any missing key.
void i18n.use(initReactI18next).init({
  resources: { es: { translation: es }, en: { translation: en } },
  lng: initialLanguage,
  fallbackLng: 'es',
  interpolation: { escapeValue: false },
  compatibilityJSON: 'v4',
});

export default i18n;
