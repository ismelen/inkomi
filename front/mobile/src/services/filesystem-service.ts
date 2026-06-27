import { pick, pickDirectory, types } from '@react-native-documents/picker';
import { Source } from '../models/source';
import { Directory, File } from 'expo-file-system';

export class FilesystemService {
  static async pickFolder(): Promise<Source | undefined> {
    const dir = await pickDirectory();
    if (!dir || !dir.uri) return;

    const files = new Directory(dir.uri).list();
    const source: Source = {
      name: decodeURIComponent(dir.uri).split('/').pop() ?? '',
      path: dir.uri,
      children: [],
    };

    files.forEach((file) => {
      if (file instanceof Directory) return;
      source.children!.push({ name: file.name, path: file.uri, size: file.size, mime: file.type });
    });

    return source;
  }

  static async pickFiles(): Promise<Source[] | undefined> {
    const files = await pick({
      allowMultiSelection: true,
      type: [types.allFiles],
    });

    if (!files || files.length === 0) return;

    return files.map(
      (file) =>
        ({
          name: file.name,
          path: file.uri,
          size: file.size,
          mime: file.type,
        }) as Source
    );
  }

  static async deleteFile(path: string) {
    new File(path).delete();
  }
}
