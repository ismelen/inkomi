import React, { useEffect, useState } from 'react';
import { Pressable, StyleSheet, View } from 'react-native';
import ChevronRightIcon from '../../src/components/icons/chevron-right-icon';
import { useEReaderModelPicker } from '../../src/components/modals/e-reader-profile-picker-modal';
import SColumn from '../../src/components/shared/SColumn';
import SText from '../../src/components/shared/SText';
import { Cloud } from '../../src/models/cloud';
import { eReaderModel } from '../../src/models/e-reader-model';
import { colors } from '../../src/theme/colors';

export default function settings() {
  const [cloud, setCloud] = useState<Cloud | undefined>();
  const [profile, setProfile] = useState<eReaderModel | undefined>();
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    Promise.all([
      Cloud.instance().then((value) => {
        setCloud(value);
      }),
      eReaderModel.instance().then((value) => {
        setProfile(value);
      }),
    ]).then(() => setLoading(false));
  }, [loading]);

  return (
    <View style={{ paddingHorizontal: 20, paddingVertical: 10 }}>
      <SText style={styles.sectionTitle}>CLOUD</SText>
      <SColumn>
        <CloudConfigField
          label="Account"
          value={cloud?.email}
          onPress={async () => {
            await cloud?.setAccount();
            setLoading(true);
          }}
        />
        <CloudConfigField
          label="Folder"
          value={cloud?.folderName}
          onPress={async () => {
            await cloud?.setFolder();
            setLoading(true);
          }}
        />
      </SColumn>

      <SText style={styles.sectionTitle}>OPTIONS</SText>
      <SColumn>
        <CloudConfigField
          label="eReader"
          value={profile?.name}
          onPress={async () => {
            const model = await useEReaderModelPicker.getState().show();
            if (!model) return;
            profile!.setModel(model);
            setLoading(true);
          }}
        />
      </SColumn>
      {/*
       * TODO: Kepubify page
       */}
    </View>
  );
}

interface CloudConfigFieldProps {
  label: string;
  onPress?(): void;
  value?: string;
}

function CloudConfigField({ label, value, onPress }: CloudConfigFieldProps) {
  return (
    <Pressable
      onPress={onPress}
      style={{
        flexDirection: 'row',
        alignItems: 'center',
        gap: 8,
      }}
    >
      <SText style={{ flex: 1, fontSize: 16, fontWeight: 500 }}>{label}</SText>
      <SText style={{ color: colors.onCard, flex: 2, textAlign: 'right' }}>{value}</SText>
      <ChevronRightIcon size="24px" color={colors.onCard} />
    </Pressable>
  );
}

const styles = StyleSheet.create({
  sectionTitle: { color: colors.onCard, marginBottom: 10, marginTop: 20 },
});
