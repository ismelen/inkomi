import React from 'react';
import { Pressable, Text, View } from 'react-native';
import ActionCard from '../../src/components/home/action-card';
import SText from '../../src/components/shared/SText';
import { colors } from '../../src/theme/colors';
import SButton from '../../src/components/shared/SButton';

export default function HomePage() {
  return (
    <View style={{ gap: 32 }}>
      <View style={{ gap: 16 }}>
        <ActionCard
          icon="menu_book"
          title="Send Comic"
          subtitle="Convert .cbz to .epub and send to device"
          tag=".cbz to .epub"
          onClick={() => {}}
        />
        <ActionCard
          icon="book"
          title="Send Book"
          subtitle="Manage and transfer .epub files"
          tag=".epub management"
          onClick={() => {}}
        />
      </View>

      <View>
        <MonitoredFoldersTitle onClick={() => {}} />
      </View>
    </View>
  );
}

function MonitoredFoldersTitle({ onClick }: { onClick(): void }) {
  return (
    <View style={{ flexDirection: 'row', alignItems: 'center', justifyContent: 'space-between' }}>
      <SText style={{ fontFamily: 'semibold', fontSize: 20 }}>Monitored folders</SText>
      <SButton
        onPress={onClick}
        style={{
          paddingHorizontal: 8,
          paddingVertical: 4,
          borderRadius: 8,
        }}
      >
        <SText style={{ color: colors.primary, fontFamily: 'semibold', fontSize: 12 }}>
          + ADD FOLDER
        </SText>
      </SButton>
    </View>
  );
}
