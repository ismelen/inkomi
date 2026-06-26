import React, { useEffect } from 'react';
import { View } from 'react-native';
import { colors } from '../../theme/colors';
import SButton from '../shared/SButton';
import SIcon from '../icons/SIcon';
import SText from '../shared/SText';
import { SourceMode, useSource } from '../../hooks/useSource';
import { Source } from '../../models/source';

interface Props {
  initSources: Source[];
  onChange(sources: Source[]): void;
  onModeChange(mode: SourceMode): void;
}

export default function SourceSelector({ initSources, onChange, onModeChange }: Props) {
  const { addFiles, addFolder, mode, sources, deleteSource } = useSource(initSources);

  useEffect(() => {
    onChange(sources);
  }, [sources]);

  useEffect(() => {
    onModeChange(mode);
  }, [mode]);

  switch (mode) {
    case 'files':
      return (
        <View style={{ gap: 10 }}>
          <SourcesViewer sources={sources} deleteSource={deleteSource} mode={mode} />
          <AddMore onClick={addFiles} />
        </View>
      );

    case 'folder':
      return <SourcesViewer sources={sources} deleteSource={deleteSource} mode={mode} />;
  }

  return (
    <View style={{ flexDirection: 'row', gap: 10 }}>
      <Option label="Add Files" icon="upload_file" onClick={addFiles} />
      <Option label="Add Folder" icon="folder_open" onClick={addFolder} />
    </View>
  );
}

function SourcesViewer({
  sources,
  deleteSource,
  mode,
}: {
  sources: Source[];
  deleteSource(index: number): void;
  mode: 'files' | 'folder' | 'no-select';
}) {
  return (
    <View
      style={{
        borderRadius: 12,
        backgroundColor: colors.surface_container_lowest,
        boxShadow: colors.boxShadow,
      }}
    >
      {sources.map((src, idx) => (
        <View
          style={{
            flexDirection: 'row',
            paddingHorizontal: 8,
            paddingVertical: 10,
            alignItems: 'center',
            justifyContent: 'space-between',
            overflow: 'hidden',
            borderTopWidth: 0.5,
            borderTopColor:
              idx !== 0 && sources.length > 0 ? colors.outline_variant : 'transparent',
          }}
          key={`${idx}${src.path}`}
        >
          <View
            style={{
              flexDirection: 'row',
              gap: 8,
              alignItems: 'center',
              flex: 1,
              minWidth: 0,
            }}
          >
            <SIcon name={mode === 'files' ? 'docs' : 'folder'} size={24} color={colors.primary} />
            <SText
              style={{ fontFamily: 'semibold', flex: 1 }}
              numberOfLines={1}
              ellipsizeMode="tail"
            >
              {src.name}
            </SText>
          </View>
          <SButton onPress={() => deleteSource(idx)}>
            <SIcon name="delete" size={24} color={colors.primary} type="outlined" />
          </SButton>
        </View>
      ))}
    </View>
  );
}

function AddMore({ onClick }: { onClick(): void }) {
  return (
    <SButton
      onPress={onClick}
      style={{
        borderRadius: 12,
        paddingVertical: 10,
        backgroundColor: colors.primary_container,
        boxShadow: colors.boxShadow,
        overflow: 'hidden',
        alignItems: 'center',
      }}
    >
      <SText
        style={{
          color: colors.on_primary,
          fontFamily: 'semibold',
          fontSize: 16,
        }}
      >
        + Add more
      </SText>
    </SButton>
  );
}

function Option({ icon, label, onClick }: { icon: string; label: string; onClick(): void }) {
  return (
    <SButton
      onPress={onClick}
      style={{
        borderRadius: 12,
        backgroundColor: colors.surface_container_lowest,
        boxShadow: colors.boxShadow,
        paddingVertical: 20,
        overflow: 'hidden',
        alignItems: 'center',
        flex: 1,
        gap: 5,
      }}
    >
      <SIcon name={icon} size={32} color={colors.primary} />
      <SText style={{ color: colors.primary, fontFamily: 'semibold', fontSize: 14 }}>{label}</SText>
    </SButton>
  );
}
