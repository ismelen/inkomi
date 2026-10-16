import { MangaConfig } from './manga-config';

/** A folder of the list: sent as a "folder" source whose items keep this order. */
export interface Group {
  id: string;
  title: string;
  join: boolean;
  rtl: boolean;
  /** "Opciones manga (todos)". */
  config?: MangaConfig;
  itemIds: string[];
  collapsed: boolean;
}
