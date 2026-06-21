import React, { useState } from 'react';
import { Pressable, View } from 'react-native';
import { Source } from '../../models/source';
import { FilesystemService } from '../../services/filesystem-service';
import { colors } from '../../theme/colors';
import AddCircleIcon from '../icons/add-cirlce-icon';
import DeleteForeverIcon from '../icons/delete-forever-icon';
import DocsIcon from '../icons/docs-icon';
import SColumn from '../shared/SColumn';
import SText from '../shared/SText';

interface Props {
  sources: Source[];
  onChange(sources: Source[]): void;
}

export default function ConfigFilesList({ sources, onChange }: Props) {
  const [files, setFiles] = useState<Source[]>(sources);

  return (
    <SColumn
      footer={
        <Pressable
          style={{
            flexDirection: 'row',
            gap: 8,
            alignItems: 'center',
          }}
          onPress={async () => {
            let newFiles = await FilesystemService.pickFiles();
            newFiles = [...files, ...newFiles];
            setFiles(newFiles);
            onChange(newFiles);
          }}
        >
          <AddCircleIcon size="22px" color={colors.primary} />
          <SText style={{ color: colors.primary, fontWeight: '600', fontSize: 16 }}>
            Add More Files
          </SText>
        </Pressable>
      }
    >
      {files.map((e, i) => (
        <View style={{ flexDirection: 'row', gap: 12, alignItems: 'center' }}>
          <DocsIcon size="27px" color={colors.onCard} />
          <SText
            style={{ flex: 1, fontSize: 16, fontWeight: 500, overflow: 'scroll' }}
            numberOfLines={2}
          >
            {e.name}
          </SText>
          <Pressable
            style={{ padding: 3 }}
            onPress={() => {
              const newFiles = files.filter((_, idx) => idx != i);
              setFiles([...newFiles]);
              onChange(newFiles);
            }}
          >
            <DeleteForeverIcon size="24px" color={colors.onCard} />
          </Pressable>
        </View>
      ))}
    </SColumn>
  );
}
