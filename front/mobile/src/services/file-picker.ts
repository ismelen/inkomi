import { pick, pickDirectory, types } from '@react-native-documents/picker';
import { Directory, File } from 'expo-file-system';
import { BookFormat } from '../api/types/books';
import { formatOf } from '../models/list-item';

export interface PickedFile {
  uri: string;
  name: string;
  size: number;
  mime?: string;
  ext: BookFormat;
}

export interface PickedFolder {
  name: string;
  files: PickedFile[];
}

function isCancel(err: unknown): boolean {
  return (err as { code?: string })?.code === 'OPERATION_CANCELED';
}

const byName = (a: PickedFile, b: PickedFile) =>
  a.name.localeCompare(b.name, undefined, { numeric: true });

/** epub, cbz and pdf files; the rest are skipped (cbz has no reliable mime type). */
export async function pickBookFiles(): Promise<PickedFile[]> {
  try {
    const files = await pick({ allowMultiSelection: true, type: [types.allFiles] });
    return files
      .map((f): PickedFile | null => {
        const name = f.name ?? '';
        const ext = formatOf(name);
        return ext ? { uri: f.uri, name, size: f.size ?? 0, mime: f.type ?? undefined, ext } : null;
      })
      .filter((f): f is PickedFile => f !== null)
      .sort(byName);
  } catch (err) {
    if (isCancel(err)) return [];
    throw err;
  }
}

/** Every supported file directly inside a folder, sorted by name. Null when cancelled. */
export async function pickBookFolder(): Promise<PickedFolder | null> {
  let uri: string;
  try {
    uri = (await pickDirectory()).uri;
  } catch (err) {
    if (isCancel(err)) return null;
    throw err;
  }

  const name = decodeURIComponent(uri).split(/[/:]/).filter(Boolean).pop() ?? '';
  const files = new Directory(uri)
    .list()
    .filter((entry): entry is File => entry instanceof File)
    .map((file): PickedFile | null => {
      const ext = formatOf(file.name);
      return ext
        ? {
            uri: file.uri,
            name: file.name,
            size: file.size ?? 0,
            mime: file.type ?? undefined,
            ext,
          }
        : null;
    })
    .filter((f): f is PickedFile => f !== null)
    .sort(byName);
  return { name, files };
}
