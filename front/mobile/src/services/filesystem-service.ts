import { pickDirectory } from "@react-native-documents/picker";
import { Source } from "../models/source";
import * as FS from 'expo-file-system/legacy';
import { getDocumentAsync } from "expo-document-picker";


export class FilesystemService {
  static async pickFolder(): Promise<Source | undefined> {
    const dir = await pickDirectory({
      requestLongTermAccess: true
    })
    if (dir.bookmarkStatus !== "success") return

    const name = decodeURIComponent(dir.uri).split("/").pop()
    if (!name) return

    const files = await FS.StorageAccessFramework.readDirectoryAsync(dir.uri)

    return {
      name,
      path: dir.uri,
      children: files,
    }
  }

  static async pickFiles(): Promise<Source[] |undefined> {
    const result = await getDocumentAsync({
      copyToCacheDirectory: false,
      multiple: true,
    })
    if(result.canceled) return

    return result.assets.map(file => ({
      name: file.name, 
      path: file.uri, 
      size: file.size
    } as Source))
  }
}