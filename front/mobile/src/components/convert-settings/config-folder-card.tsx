import React, { useState } from 'react';
import { Pressable, View } from 'react-native';
import { Source } from '../../models/source';
import { FilesystemService } from '../../services/filesystem-service';
import { colors } from '../../theme/colors';
import ChangeIcon from '../icons/change-icon';
import DeleteForeverIcon from '../icons/delete-forever-icon';
import FolderIcon from '../icons/folder-icon';
import SText from '../shared/SText';

interface Props {
  source: Source;
  onChange(source?: Source): void;
  isMonitorized: boolean;
}

export default function ConfigFolderCard({ source, onChange, isMonitorized }: Props) {
  const [folder, setFolder] = useState<Source | undefined>(source);

  return (
    <View
      style={{
        flexDirection: 'row',
        gap: 12,
        alignItems: 'center',
        backgroundColor: colors.card,
        borderRadius: 14,
        paddingHorizontal: 16,
        paddingVertical: 14,
      }}
    >
      <FolderIcon size="27px" color={colors.onCard} />
      <SText
        style={{ flex: 1, fontSize: 16, fontWeight: 500, overflow: 'scroll' }}
        numberOfLines={2}
      >
        {folder?.name}
      </SText>
      {!isMonitorized && (
        <Pressable
          style={{ padding: 3 }}
          onPress={async () => {
            const newFolder = await FilesystemService.pickFolder();
            if (!newFolder) return;

            setFolder(newFolder);
            onChange(newFolder);
          }}
        >
          <ChangeIcon size="24px" color={colors.primary} />
        </Pressable>
      )}
      <Pressable
        style={{ padding: 3 }}
        onPress={() => {
          setFolder(undefined);
          onChange(undefined);
        }}
      >
        <DeleteForeverIcon size="24px" color={colors.primary} />
      </Pressable>
    </View>
  );
}
