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

/** Lets the UI paint (e.g. a loading indicator) before the next blocking native call. */
export const nextFrame = () => new Promise<void>((resolve) => setTimeout(resolve, 16));

/** Files whose metadata is read between two frames: each read is a blocking native call. */
const filesPerFrame = 8;

/**
 * Every supported file directly inside a folder, sorted by name. Null when cancelled.
 * onLoading is called once the user picked the folder, before the slow part (listing it and
 * reading every file's size, which on Android goes through the storage provider), and then with
 * the progress of that read.
 */
export async function pickBookFolder(
  onLoading?: (read: number, total: number) => void
): Promise<PickedFolder | null> {
  let uri: string;
  try {
    uri = (await pickDirectory()).uri;
  } catch (err) {
    if (isCancel(err)) return null;
    throw err;
  }

  onLoading?.(0, 0);
  await nextFrame();
  const name = decodeURIComponent(uri).split(/[/:]/).filter(Boolean).pop() ?? '';
  const entries = new Directory(uri)
    .list()
    .filter((entry): entry is File => entry instanceof File && formatOf(entry.name) !== null);

  const files: PickedFile[] = [];
  for (let i = 0; i < entries.length; i++) {
    if (i % filesPerFrame === 0) {
      onLoading?.(i, entries.length);
      await nextFrame();
    }
    const file = entries[i];
    files.push({
      uri: file.uri,
      name: file.name,
      size: file.size ?? 0,
      mime: file.type ?? undefined,
      ext: formatOf(file.name)!,
    });
  }
  onLoading?.(entries.length, entries.length);
  return { name, files: files.sort(byName) };
}
