import Constants from 'expo-constants';

export const WEB_CLIENT_ID = Constants.expoConfig!.extra!.webClientId;
export const WEB_CLIENT_SECRET = Constants.expoConfig!.extra!.webClientSecret;
export const BACKENDD_URL = Constants.expoConfig!.extra!.backendUrl;

export const DROPBOX_API_KEY = Constants.expoConfig!.extra!.dropboxApiKey;

export const eReaderProfiles = [
  { value: 'K1', label: 'Kindle 1' },
  { value: 'K2', label: 'Kindle 2' },
  { value: 'KDX', label: 'Kindle DX/DXG' },
  { value: 'K34', label: 'Kindle Keyboard/Touch' },
  { value: 'K57', label: 'Kindle 5/7' },
  { value: 'KPW', label: 'Kindle Paperwhite 1/2' },
  { value: 'KV', label: 'Kindle Voyage' },
  { value: 'KPW34', label: 'Kindle Paperwhite 3/4/Oasis' },
  { value: 'K810', label: 'Kindle 8/10' },
  { value: 'KO', label: 'Kindle Oasis 2/3/Paperwhite 12' },
  { value: 'K11', label: 'Kindle 11' },
  { value: 'KPW5', label: 'Kindle Paperwhite 5/Signature Edition' },
  { value: 'KS', label: 'Kindle Scribe' },
  { value: 'KCS', label: 'Kindle Colorsoft' },

  // Kobo
  { value: 'KoMT', label: 'Kobo Mini/Touch' },
  { value: 'KoG', label: 'Kobo Glo' },
  { value: 'KoGHD', label: 'Kobo Glo HD' },
  { value: 'KoA', label: 'Kobo Aura' },
  { value: 'KoAHD', label: 'Kobo Aura HD' },
  { value: 'KoAH2O', label: 'Kobo Aura H2O' },
  { value: 'KoAO', label: 'Kobo Aura ONE' },
  { value: 'KoN', label: 'Kobo Nia' },
  { value: 'KoC', label: 'Kobo Clara HD/Kobo Clara 2E' },
  { value: 'KoCC', label: 'Kobo Clara Colour' },
  { value: 'KoL', label: 'Kobo Libra H2O/Kobo Libra 2' },
  { value: 'KoLC', label: 'Kobo Libra Colour' },
  { value: 'KoF', label: 'Kobo Forma' },
  { value: 'KoS', label: 'Kobo Sage' },
  { value: 'KoE', label: 'Kobo Elipsa' },
];
