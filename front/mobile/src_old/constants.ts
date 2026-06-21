import Constants from 'expo-constants';

export const eReaderProfiles: Record<string, string> = {
  K1: 'Kindle 1',
  K2: 'Kindle 2',
  KDX: 'Kindle DX/DXG',
  K34: 'Kindle Keyboard/Touch',
  K57: 'Kindle 5/7',
  KPW: 'Kindle Paperwhite 1/2',
  KV: 'Kindle Voyage',
  KPW34: 'Kindle Paperwhite 3/4/Oasis',
  K810: 'Kindle 8/10',
  KO: 'Kindle Oasis 2/3/Paperwhite 12',
  K11: 'Kindle 11',
  KPW5: 'Kindle Paperwhite 5/Signature Edition',
  KS: 'Kindle Scribe',
  KCS: 'Kindle Colorsoft',

  // Kobo
  KoMT: 'Kobo Mini/Touch',
  KoG: 'Kobo Glo',
  KoGHD: 'Kobo Glo HD',
  KoA: 'Kobo Aura',
  KoAHD: 'Kobo Aura HD',
  KoAH2O: 'Kobo Aura H2O',
  KoAO: 'Kobo Aura ONE',
  KoN: 'Kobo Nia',
  KoC: 'Kobo Clara HD/Kobo Clara 2E',
  KoCC: 'Kobo Clara Colour',
  KoL: 'Kobo Libra H2O/Kobo Libra 2',
  KoLC: 'Kobo Libra Colour',
  KoF: 'Kobo Forma',
  KoS: 'Kobo Sage',
  KoE: 'Kobo Elipsa',
};

export const WEB_CLIENT_ID = Constants.expoConfig!.extra!.webClientId;
export const WEB_CLIENT_SECRET = Constants.expoConfig!.extra!.webClientSecret;
export const BACKENDD_URL = Constants.expoConfig!.extra!.backendUrl;
