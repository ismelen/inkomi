import { BookFormat } from '../api/types/books';
import { MangaConfig } from './manga-config';

interface BaseItem {
  id: string;
  title: string;
  ext: BookFormat;
  rtl: boolean;
  /** Overrides the group or default manga options. Ignored for epub. */
  config?: MangaConfig;
  groupId?: string;
}

/** A file picked or dropped in the browser, uploaded with a presigned PUT. */
export interface FileItem extends BaseItem {
  kind: 'file';
  /** Kept in IndexedDB with the draft, so the list survives a reload. */
  file: File;
  name: string;
  size: number;
  mime?: string;
}

/** A book found with /books/search; the server downloads it by md5. */
export interface LibraryItem extends BaseItem {
  kind: 'library';
  md5: string;
  author?: string;
}

export type ListItem = FileItem | LibraryItem;

export const supportedFormats: BookFormat[] = ['epub', 'cbz', 'pdf'];

export function formatOf(filename: string): BookFormat | null {
  const ext = filename.split('.').pop()?.toLowerCase();
  return supportedFormats.find((f) => f === ext) ?? null;
}

export function isManga(ext: BookFormat): boolean {
  return ext !== 'epub';
}

export function titleFromFilename(filename: string): string {
  const dot = filename.lastIndexOf('.');
  return dot > 0 ? filename.slice(0, dot) : filename;
}
