import React from 'react';
import { Pressable, Text, View } from 'react-native';
import ActionCard from '../../src/components/home/action-card';
import SText from '../../src/components/shared/SText';
import { colors } from '../../src/theme/colors';
import SButton from '../../src/components/shared/SButton';
import { router, Tabs } from 'expo-router';
import AppHeader from '../../src/components/app-header';

export default function HomePage() {
  return (
    <>
      <Tabs.Screen options={{ headerShown: true, header: () => <AppHeader /> }} />
      <View style={{ gap: 32, paddingHorizontal: 24 }}>
        <View style={{ gap: 16 }}>
          <ActionCard
            icon="menu_book"
            title="Send Comic"
            subtitle="Convert .cbz to .epub and send to device"
            tag=".cbz to .epub"
            onClick={() => router.push('/send-comic')}
          />
          <ActionCard
            icon="book"
            title="Send Book"
            subtitle="Manage and transfer .epub files"
            tag=".epub management"
            onClick={() => router.push('/send-book')}
          />
        </View>

        <View>
          <MonitoredFoldersTitle onClick={() => {}} />
        </View>
      </View>
    </>
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
