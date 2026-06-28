import React from 'react';
import { StyleSheet, Text, View } from 'react-native';
import { ScrollView } from 'react-native-gesture-handler';
import SText from '../../src/components/shared/SText';
import { colors } from '../../src/theme/colors';
import SSelect from '../../src/components/shared/SSelect';
import { eReaderProfiles } from '../../src/constants';
import { useSettings } from '../../src/hooks/useSettings';
import { useShallow } from 'zustand/react/shallow';
import SIcon from '../../src/components/icons/SIcon';
import { useCloud } from '../../src/hooks/useCloud';
import SButton from '../../src/components/shared/SButton';

export default function SettingsPage() {
  const { model, setModel } = useSettings(
    useShallow((s) => ({ model: s.model, setModel: s.setModel }))
  );

  const { email, getToken, getFolder, folder, logout } = useCloud(
    useShallow((s) => ({
      email: s.email,
      getToken: s.getToken,
      getFolder: s.getFolder,
      folder: s.folderPath,
      logout: s.logout,
    }))
  );

  return (
    <ScrollView style={{ flex: 1, paddingHorizontal: 24 }}>
      <Text style={{ fontFamily: 'bold', fontSize: 28 }}>Settings</Text>
      <View style={{ marginTop: 14, gap: 4 }}>
        <SText style={styles.title}>READER MODEL</SText>
        <SSelect
          value={model}
          options={eReaderProfiles}
          onOptionChange={(opt) => setModel(opt.value)}
        />
      </View>

      <View style={{ marginTop: 32, gap: 4 }}>
        <SText style={styles.title}>CLOUD SYNCHRONIZATION</SText>
        <View
          style={{
            boxShadow: colors.boxShadow,
            borderRadius: 12,
            backgroundColor: colors.surface_container_lowest,
            padding: 10,
          }}
        >
          <View style={{ flexDirection: 'row', gap: 10, alignItems: 'center' }}>
            <SIcon name="cloud" color={colors.primary} size={32} type="outlined" />
            <SText style={{ fontSize: 18, flex: 1, fontFamily: 'semibold' }}>Dropbox</SText>
            <SButton
              onPress={() => (email ? logout() : getToken(true))}
              style={{
                borderWidth: 1,
                borderColor: colors.primary_fixed,
                alignSelf: 'flex-start',
                paddingHorizontal: 14,
                paddingVertical: 7,
                borderRadius: 12,
              }}
            >
              <SText style={{ fontFamily: 'semibold', color: colors.primary }}>
                {email ? 'Disconnect' : 'Connect'}
              </SText>
            </SButton>
          </View>
          {email && (
            <View
              style={{
                flexDirection: 'row',
                gap: 5,
                alignItems: 'center',
                marginTop: 5,
              }}
            >
              <SIcon name="check_circle" color={colors.primary} size={14} />
              <SText style={{ fontSize: 12, color: colors.primary, fontFamily: 'semibold' }}>
                Connected as {email}
              </SText>
            </View>
          )}
        </View>
      </View>

      {email && (
        <View
          style={{
            marginTop: 10,
            boxShadow: colors.boxShadow,
            backgroundColor: colors.surface_container_lowest,
            borderRadius: 12,
            padding: 10,
          }}
        >
          <SText style={{ fontSize: 14, fontFamily: 'semibold', marginBottom: 2 }}>Folder</SText>
          <View style={{ flexDirection: 'row', gap: 10 }}>
            <View
              style={{
                flex: 1,
                borderWidth: 0.5,
                borderColor: colors.outline,
                borderRadius: 12,
                padding: 10,
              }}
            >
              <SText>{folder ?? 'Select folder'}</SText>
            </View>
            <SButton
              onPress={() => getFolder(true)}
              style={{
                backgroundColor: colors.primary,
                paddingHorizontal: 14,
                paddingVertical: 7,
                borderRadius: 12,
                justifyContent: 'center',
              }}
            >
              <SText
                style={{
                  color: colors.on_primary,
                  fontFamily: 'semibold',
                }}
              >
                Browse
              </SText>
            </SButton>
          </View>
        </View>
      )}
    </ScrollView>
  );
}

const styles = StyleSheet.create({
  title: {
    fontFamily: 'semibold',
    fontSize: 14,
    color: colors.on_surface_variant,
  },
});
