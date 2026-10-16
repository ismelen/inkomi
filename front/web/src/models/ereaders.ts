// Mirror of back/transaction-manager/internal/domain/models/ereader.go.

export interface EReader {
  key: string;
  name: string;
  width: number;
  height: number;
  /** Kobo devices: the server converts the result to kepub. */
  kepub: boolean;
}

export const ereaders: EReader[] = [
  { key: 'K1', name: 'Kindle 1', width: 600, height: 670, kepub: false },
  { key: 'K2', name: 'Kindle 2', width: 600, height: 670, kepub: false },
  { key: 'KDX', name: 'Kindle DX/DXG', width: 824, height: 1000, kepub: false },
  { key: 'K34', name: 'Kindle Keyboard/Touch', width: 600, height: 800, kepub: false },
  { key: 'K57', name: 'Kindle 5/7', width: 600, height: 800, kepub: false },
  { key: 'KPW', name: 'Kindle Paperwhite 1/2', width: 758, height: 1024, kepub: false },
  { key: 'KV', name: 'Kindle Voyage', width: 1072, height: 1448, kepub: false },
  { key: 'KPW34', name: 'Kindle Paperwhite 3/4/Oasis', width: 1072, height: 1448, kepub: false },
  { key: 'K810', name: 'Kindle 8/10', width: 600, height: 800, kepub: false },
  { key: 'KO', name: 'Kindle Oasis 2/3/Paperwhite 12', width: 1264, height: 1680, kepub: false },
  { key: 'K11', name: 'Kindle 11', width: 1072, height: 1448, kepub: false },
  { key: 'KPW5', name: 'Kindle Paperwhite 5/Signature', width: 1236, height: 1648, kepub: false },
  { key: 'KS', name: 'Kindle Scribe', width: 1860, height: 2480, kepub: false },
  { key: 'KCS', name: 'Kindle Colorsoft', width: 1264, height: 1680, kepub: false },
  { key: 'KoMT', name: 'Kobo Mini/Touch', width: 600, height: 800, kepub: true },
  { key: 'KoG', name: 'Kobo Glo', width: 768, height: 1024, kepub: true },
  { key: 'KoGHD', name: 'Kobo Glo HD', width: 1072, height: 1448, kepub: true },
  { key: 'KoA', name: 'Kobo Aura', width: 758, height: 1024, kepub: true },
  { key: 'KoAHD', name: 'Kobo Aura HD', width: 1080, height: 1440, kepub: true },
  { key: 'KoAH2O', name: 'Kobo Aura H2O', width: 1080, height: 1430, kepub: true },
  { key: 'KoAO', name: 'Kobo Aura ONE', width: 1404, height: 1872, kepub: true },
  { key: 'KoN', name: 'Kobo Nia', width: 758, height: 1024, kepub: true },
  { key: 'KoC', name: 'Kobo Clara HD/2E', width: 1072, height: 1448, kepub: true },
  { key: 'KoCC', name: 'Kobo Clara Colour', width: 1072, height: 1448, kepub: true },
  { key: 'KoL', name: 'Kobo Libra H2O/2', width: 1264, height: 1680, kepub: true },
  { key: 'KoLC', name: 'Kobo Libra Colour', width: 1264, height: 1680, kepub: true },
  { key: 'KoF', name: 'Kobo Forma', width: 1440, height: 1920, kepub: true },
  { key: 'KoS', name: 'Kobo Sage', width: 1440, height: 1920, kepub: true },
  { key: 'KoE', name: 'Kobo Elipsa', width: 1404, height: 1872, kepub: true },
];

export const defaultEReaderKey = 'KPW5';

export function findEReader(key: string): EReader {
  return ereaders.find((e) => e.key === key) ?? ereaders.find((e) => e.key === defaultEReaderKey)!;
}
