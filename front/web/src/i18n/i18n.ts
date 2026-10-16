import i18n from 'i18next';
import { initReactI18next } from 'react-i18next';
import en from './locales/en.json';
import es from './locales/es.json';

export const supportedLanguages = ['es', 'en'] as const;
export type SupportedLanguage = (typeof supportedLanguages)[number];

const browserLanguage = navigator.language.split('-')[0];
const initialLanguage: SupportedLanguage = supportedLanguages.includes(
  browserLanguage as SupportedLanguage
)
  ? (browserLanguage as SupportedLanguage)
  : 'es';

// Spanish is the base language: it is the fallback for any missing key.
void i18n.use(initReactI18next).init({
  resources: { es: { translation: es }, en: { translation: en } },
  lng: initialLanguage,
  fallbackLng: 'es',
  interpolation: { escapeValue: false },
});

i18n.on('languageChanged', (lng) => {
  document.documentElement.lang = lng;
});
document.documentElement.lang = initialLanguage;

export default i18n;
