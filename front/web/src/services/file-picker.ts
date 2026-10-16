import { BookFormat } from '../api/types/books';
import { formatOf, supportedFormats } from '../models/list-item';

export interface PickedFile {
  file: File;
  name: string;
  size: number;
  mime?: string;
  ext: BookFormat;
}

export interface PickedFolder {
  name: string;
  files: PickedFile[];
}

export interface DroppedEntries {
  files: PickedFile[];
  folders: PickedFolder[];
}

const byName = (a: PickedFile, b: PickedFile) =>
  a.name.localeCompare(b.name, undefined, { numeric: true });

export const acceptedExtensions = supportedFormats.map((f) => `.${f}`).join(',');

/** epub, cbz and pdf files; the rest are skipped (cbz has no reliable mime type). */
export function toPickedFiles(files: Iterable<File>): PickedFile[] {
  return [...files]
    .map((file): PickedFile | null => {
      const ext = formatOf(file.name);
      return ext
        ? { file, name: file.name, size: file.size, mime: file.type || undefined, ext }
        : null;
    })
    .filter((f): f is PickedFile => f !== null)
    .sort(byName);
}

function openInput(configure: (input: HTMLInputElement) => void): Promise<File[] | null> {
  return new Promise((resolve) => {
    const input = document.createElement('input');
    input.type = 'file';
    configure(input);
    input.addEventListener('change', () => resolve(input.files ? [...input.files] : []));
    input.addEventListener('cancel', () => resolve(null));
    input.click();
  });
}

export async function pickBookFiles(): Promise<PickedFile[]> {
  const files = await openInput((input) => {
    input.multiple = true;
    input.accept = acceptedExtensions;
  });
  return files ? toPickedFiles(files) : [];
}

/** Supported files directly inside the picked folder (like the app). Null when cancelled. */
export async function pickBookFolder(): Promise<PickedFolder | null> {
  const files = await openInput((input) => {
    input.webkitdirectory = true;
  });
  if (!files || files.length === 0) return null;

  const name = files[0].webkitRelativePath.split('/')[0] ?? '';
  const direct = files.filter((f) => f.webkitRelativePath.split('/').length === 2);
  return { name, files: toPickedFiles(direct) };
}

function readFile(entry: FileSystemFileEntry): Promise<File> {
  return new Promise((resolve, reject) => entry.file(resolve, reject));
}

async function readDirectory(entry: FileSystemDirectoryEntry): Promise<File[]> {
  const reader = entry.createReader();
  const entries: FileSystemEntry[] = [];
  // readEntries returns the children in batches until an empty one.
  for (;;) {
    const batch = await new Promise<FileSystemEntry[]>((resolve, reject) =>
      reader.readEntries(resolve, reject)
    );
    if (batch.length === 0) break;
    entries.push(...batch);
  }
  return Promise.all(
    entries.filter((e): e is FileSystemFileEntry => e.isFile).map((e) => readFile(e))
  );
}

/** Files and folders dropped from the system. Folders keep only their direct files. */
export async function readDrop(data: DataTransfer): Promise<DroppedEntries> {
  // The DataTransfer is emptied after the drop event: read it before the first await.
  const plainFiles = [...data.files];
  const entries = [...data.items]
    .filter((item) => item.kind === 'file')
    .map((item) => item.webkitGetAsEntry());

  const looseFiles: File[] = [];
  const folders: PickedFolder[] = [];
  for (const entry of entries) {
    if (!entry) continue;
    if (entry.isDirectory) {
      const files = await readDirectory(entry as FileSystemDirectoryEntry);
      folders.push({ name: entry.name, files: toPickedFiles(files) });
    } else if (entry.isFile) {
      looseFiles.push(await readFile(entry as FileSystemFileEntry));
    }
  }
  if (entries.every((e) => !e)) looseFiles.push(...plainFiles);
  return { files: toPickedFiles(looseFiles), folders };
}
