import { router, Stack } from 'expo-router';
import React from 'react';
import { ScrollView, StyleSheet, View } from 'react-native';
import { colors } from '../src/theme/colors';
import SText from '../src/components/shared/SText';
import SourceSelector from '../src/components/senders/source-selector';
import DestinationSelector from '../src/components/senders/destination-selector';
import OptionCardChecker from '../src/components/senders/option-card-checker';
import SButton from '../src/components/shared/SButton';
import MetadataSection from '../src/components/senders/metadata-section';
import {  TransactionRequest } from '../src/models/transaction-request';
import { useQueue } from '../src/hooks/useQueue';

export default function SendComicPage() {
  const req: Partial<TransactionRequest> = {}
  const send = useQueue(s => s.send)
  
  return (
    <>
      <Stack.Screen
        options={{
          headerShown: true,
          title: 'Send Comic',
          headerTitleStyle: { fontFamily: 'semibold', fontSize: 20, color: colors.on_background },
          headerTitleAlign: 'center',
          headerStyle: {
            backgroundColor: colors.background,
          },
          headerTintColor: colors.primary,
        }}
      />
      <ScrollView style={{ flex: 1, paddingBottom: 24, paddingHorizontal: 24 }}>
        <View style={{ flex: 1, gap: 32, paddingBottom: 24, }}>
          <View style={styles.section}>
            <SText style={styles.title}>SOURCE</SText>
            <SourceSelector initSources={req.sources ?? []} onChange={(srcs) => req.sources = srcs} />
          </View>

          <View style={styles.section}>
            <SText style={styles.title}>METADATA</SText>
            <MetadataSection initialMetadata={{
              title: req.title,
              author: req.author
            }} onChange={(meta) => {
              req.author = meta.author
              req.title = meta.title
            }} />
          </View>

          <View style={styles.section}>
            <SText style={styles.title}>DESTINATION</SText>
            <DestinationSelector initDestination={req.destination ?? 'local'} onChange={(dest) => req.destination = dest} />
          </View>

          <View style={{ gap: 5 }}>
            <SText style={styles.title}>OPTIONS</SText>
            <OptionCardChecker
              initialChecked={req.merge ?? false}
              label="Merge chapters"
              text="Combine multiple chapters into a single volume"
              onChange={(checked) => req.merge = checked}
            />
            <OptionCardChecker
              initialChecked={req.deleteOrigin ?? false}
              label="Delete source"
              text="Remove original after successful upload"
              onChange={(checked) => req.deleteOrigin = checked}
            />
          </View>
        </View>

      </ScrollView>
        <SButton
          disabled={!req.sources || req.sources.length == 0}
          onPress={async () => {
            const done = await send(req)
            if(done) router.navigate("/(tabs)/queue")
          }}
          style={{
            backgroundColor: colors.primary_container,
            margin: 24,
            paddingVertical: 12,
            alignItems: 'center',
            justifyContent: 'center',
            borderRadius: 12,
            boxShadow: colors.boxShadow,
          }}
        >
          <SText style={{ fontFamily: 'semibold', color: colors.on_primary }}>Send</SText>
        </SButton>
    </>
  );
}

const styles = StyleSheet.create({
  title: {
    fontFamily: 'semibold',
    fontSize: 14,
    color: colors.on_surface_variant,
  },
  section: {
    gap: 5,
  },
});
