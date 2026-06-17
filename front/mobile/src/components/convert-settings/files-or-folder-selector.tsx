import React from 'react';
import { Pressable, View } from 'react-native';
import { colors } from '../../theme/colors';
import OpenFolderIcon from '../icons/open-folder-icon';
import UploadFileIcon from '../icons/upload-file-icon';
import SText from '../shared/SText';

interface Props {
  onFilesAdd(): void;
  onFolderAdd(): void;
}

interface Option {
  label: string;
  icon: React.JSX.Element;
  func: () => void;
}

export default function FilesOrFolderSelector({ onFilesAdd, onFolderAdd }: Props) {
  const options: Option[] = [
    {
      label: 'Add Files',
      icon: <UploadFileIcon size="35px" color={colors.onCard} />,
      func: onFilesAdd,
    },
    {
      label: 'Add Folder',
      icon: <OpenFolderIcon size="35px" color={colors.onCard} />,
      func: onFolderAdd,
    },
  ];

  return (
    <View style={{ flexDirection: 'row', gap: 14 }}>
      {options.map((e, i) => (
        <Pressable
          key={i}
          onPress={e.func}
          style={{
            flex: 1,
            justifyContent: 'center',
            alignItems: 'center',
            paddingVertical: 20,
            paddingHorizontal: 16,
            backgroundColor: colors.card,
            borderRadius: 14,
            gap: 10,
          }}
        >
          {e.icon}
          <SText style={{ color: colors.onCard, fontSize: 16 }}>{e.label}</SText>
        </Pressable>
      ))}
    </View>
  );
}
