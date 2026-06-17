import { router, Stack, useLocalSearchParams } from 'expo-router';
import { useEffect, useState } from 'react';
import { View } from 'react-native';
import ConfigToggleField from '../../../../src/components/convert-settings/config-toogle-field';
import ConvertLoading from '../../../../src/components/convert-settings/convert-loading';
import ConvertSettingsPage from '../../../../src/components/convert-settings/convert-settings-page';
import KepubifySettingsPage from '../../../../src/components/convert-settings/kepubify-settings-page';
import SDivider from '../../../../src/components/shared/SDivider';
import { useMonitoredFolders } from '../../../../src/hooks/use-monitored-folders';
import { useQueue } from '../../../../src/hooks/use-queue';
import { QueueRequest } from '../../../../src/models/queue';
import { Source } from '../../../../src/models/source';
import { UploadSettings } from '../../../../src/models/upload';
import { KepubifyService } from '../../../../src/services/kepubify-service';
import { MangaConvertService } from '../../../../src/services/manga-convert-service';
import { colors } from '../../../../src/theme/colors';

export default function index() {
  let { idx, kepubify } = useLocalSearchParams();
  const [kepubifyVal, setKepubifyVal] = useState(kepubify === true.toString());
  const updateFolder = useMonitoredFolders((s) => s.updateFolder);
  const deleteFolder = useMonitoredFolders((s) => s.deleteFolder);
  const updateFolderSettings = useMonitoredFolders((s) => s.updateFolderSettings);
  const [loading, setLoading] = useState(false);

  let sources: Source[] = [];

  let settings = UploadSettings.default('', '');
  const isMonitored = idx !== '-1';

  if (isMonitored) {
    const folder = useMonitoredFolders.getState().folders[Number(idx)];
    if (!folder) router.back();
    sources.push(folder.source);
    settings = folder.settings;
    kepubify = folder.kepubify.toString();
  }

  useEffect(() => {
    if (isMonitored) {
      setKepubifyVal(kepubify === true.toString());
    }
  }, []);

  const handleSaveSettings = (newSettings: UploadSettings, newSources: Source[]) => {
    if (newSources.length === 0) {
      deleteFolder(Number(idx));
      router.back();
      return;
    }
    updateFolderSettings(newSettings, kepubifyVal, Number(idx));
  };

  const handleProcess = async (
    newSettings: UploadSettings,
    newSources: Source[],
    isFilesMode: boolean,
    convert: (paths: string[], settings: UploadSettings) => Promise<QueueRequest | undefined>
  ) => {
    if (isMonitored) {
      handleSaveSettings(newSettings, newSources);
    }

    if (newSources.length === 0) {
      alert('Nothing to upload!!');
      return;
    }

    sources = newSources;
    settings = newSettings;

    let paths: string[] = [];
    if (isFilesMode) {
      paths = newSources.map((e) => e.path);
    } else {
      paths = newSources[0].children!;
    }

    setLoading(true);
    const request = await convert(paths, settings);
    setLoading(false);

    if (!request) return;

    const handleUpdateFolder = () => {
      updateFolder(
        {
          source: {
            ...newSources[0],
            children: settings.deleteFilesAfterUpload ? [] : newSources[0].children,
          },
          settings: settings,
          uploaded: true,
          kepubify: kepubifyVal,
          lastUploadedPahts: paths,
        },
        Number(idx)
      );
    };

    if (kepubifyVal) {
      if (isMonitored) {
        handleUpdateFolder();
      }

      router.replace('/(tabs)');
      return;
    }

    request.sources = sources;
    if (isMonitored) {
      settings.initialVolume! += request.times.length;
      handleUpdateFolder();
    }

    useQueue.getState().add(request);
    router.replace('/(tabs)/queue');
  };

  if (loading) return <ConvertLoading />;

  return (
    <View style={{ flex: 1 }}>
      <Stack.Screen
        options={{
          headerShown: true,
          title: 'Settings',
          headerTintColor: colors.primary,
          headerTitleStyle: {
            color: colors.white,
          },
          headerTitleAlign: 'center',
        }}
      />
      <View
        style={{
          borderRadius: 14,
          backgroundColor: colors.card,
          marginHorizontal: 20,
          marginVertical: 10,
        }}
      >
        <ConfigToggleField
          initial={kepubify === true.toString()}
          label="Kepubify"
          onChange={(value) => setKepubifyVal(value)}
        />
      </View>
      <SDivider />
      {kepubifyVal && (
        <KepubifySettingsPage
          isMonitored={isMonitored}
          initSources={sources}
          settings={settings}
          onProcess={async (newSettings, newSources, isFilesMode) => {
            handleProcess(newSettings, newSources, isFilesMode, KepubifyService.convert);
          }}
          onSaveSettings={(newSettings, newSources) => {
            handleSaveSettings(newSettings, newSources);
            router.back();
          }}
        />
      )}
      {!kepubifyVal && (
        <ConvertSettingsPage
          isMonitored={isMonitored}
          initSources={sources}
          settings={settings}
          onProcess={async (newSettings, newSources, isFilesMode) => {
            handleProcess(newSettings, newSources, isFilesMode, MangaConvertService.convert);
          }}
          onSaveSettings={(newSettings, newSources) => {
            handleSaveSettings(newSettings, newSources);
            router.back();
          }}
        />
      )}
    </View>
  );
}
