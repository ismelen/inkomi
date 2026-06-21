import React, { useState } from 'react';
import { StyleSheet, View } from 'react-native';
import { ScrollView } from 'react-native-gesture-handler';
import { Source } from '../../models/source';
import { UploadSettings } from '../../models/upload';
import { FilesystemService } from '../../services/filesystem-service';
import { colors } from '../../theme/colors';
import SButton from '../shared/SButton';
import SColumn from '../shared/SColumn';
import SDivider from '../shared/SDivider';
import SText from '../shared/SText';
import CloudField from './cloud-field';
import ConfigFilesList from './config-files-list';
import ConfigFolderCard from './config-folder-card';
import ConfigToggleField from './config-toogle-field';
import FilesOrFolderSelector from './files-or-folder-selector';

interface Props {
  isMonitored: boolean;
  initSources: Source[];
  settings: UploadSettings;
  onSaveSettings(settings: UploadSettings, sources: Source[]): void;
  onProcess(settings: UploadSettings, sources: Source[], filesMode: boolean): void;
}

export default function KepubifySettingsPage({
  isMonitored,
  initSources,
  settings,
  onProcess,
  onSaveSettings,
}: Props) {
  const [filesMode, setFilesMode] = useState(!isMonitored);
  const [sources, setSources] = useState<Source[]>(initSources);

  return (
    <View style={{ flex: 1 }}>
      <ScrollView style={{ paddingHorizontal: 20 }}>
        <SText style={styles.sectionTitle}>SOURCE</SText>
        {sources.length === 0 && (
          <FilesOrFolderSelector
            onFilesAdd={async () => {
              const files = await FilesystemService.pickFiles();
              setSources(files);
              setFilesMode(true);
            }}
            onFolderAdd={async () => {
              const folder = await FilesystemService.pickFolder();
              if (!folder) return;
              setSources([folder]);
              setFilesMode(false);
            }}
          />
        )}
        {sources.length !== 0 && filesMode && (
          <ConfigFilesList sources={sources} onChange={(srcs) => setSources(srcs)} />
        )}
        {sources.length !== 0 && !filesMode && (
          <ConfigFolderCard
            source={sources[0]}
            onChange={(src) => {
              if (!isMonitored) {
                setSources(src ? [src] : []);
                return;
              }

              onProcess(settings, [], filesMode);
            }}
            isMonitorized={isMonitored}
          />
        )}

        <SText style={styles.sectionTitle}>CONFIGURATION</SText>
        <SColumn removeCellPadding>
          <ConfigToggleField
            initial={settings.deleteFilesAfterUpload}
            label="Delete files after upload"
            onChange={(value) => (settings.deleteFilesAfterUpload = value)}
          />
          <CloudField toCloud={settings.cloud} onChange={(dest) => (settings.cloud = dest)} />
          {isMonitored && (
            <ConfigToggleField
              initial={settings.autoUpload}
              label="Auto upload"
              onChange={(value) => (settings.autoUpload = value)}
            />
          )}
        </SColumn>
        <View style={{ height: 30 }} />
      </ScrollView>
      <SDivider />
      <View style={{ paddingVertical: 20, gap: 20 }}>
        {isMonitored && (
          <SButton
            onPress={() => onSaveSettings(settings, sources)}
            style={{ marginHorizontal: 20, backgroundColor: colors.card }}
          >
            <SText style={{ fontSize: 18, fontWeight: '700' }}>Save settings</SText>
          </SButton>
        )}
        <SButton
          onPress={() => onProcess(settings, sources, filesMode)}
          style={{ marginHorizontal: 20 }}
        >
          <SText style={{ fontSize: 18, fontWeight: '700' }}>Start Processing</SText>
        </SButton>
      </View>
    </View>
  );
}

const styles = StyleSheet.create({
  sectionTitle: { color: colors.onCard, marginBottom: 10, marginTop: 20 },
  label: { fontWeight: '500', fontSize: 16 },
});
