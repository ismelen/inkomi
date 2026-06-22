import { Stack } from 'expo-router';
import React from 'react';
import { StyleSheet, View } from 'react-native';
import { colors } from '../src/theme/colors';
import SText from '../src/components/shared/SText';
import SourceSelector from '../src/components/senders/source-selector';
import DestinationSelector from '../src/components/senders/destination-selector';
import OptionCardChecker from '../src/components/senders/option-card-checker';
import SButton from '../src/components/shared/SButton';
import MetadataSection from '../src/components/senders/metadata-section';

export default function SendComicPage() {
  return (
    <>
      <Stack.Screen
        options={{
          headerShown: true,
          title: 'Send Book',
          headerTitleStyle: { fontFamily: 'semibold', fontSize: 20, color: colors.on_background },
          headerTitleAlign: 'center',
          headerStyle: {
            backgroundColor: colors.background,
          },
          headerTintColor: colors.primary,
        }}
      />
      <View style={{ flex: 1, paddingBottom: 24 }}>
        <View style={{ flex: 1, gap: 32 }}>
          <View style={styles.section}>
            <SText style={styles.title}>SOURCE</SText>
            <SourceSelector initSources={[]} onChange={(srcs) => {}} />
          </View>

          <View style={styles.section}>
            <SText style={styles.title}>METADATA</SText>
            <MetadataSection initialMetadata={{}} onChange={(meta) => {}} />
          </View>

          <View style={styles.section}>
            <SText style={styles.title}>DESTINATION</SText>
            <DestinationSelector initDestination="local" onChange={(dest) => {}} />
          </View>

          <View style={{ gap: 5 }}>
            <SText style={styles.title}>OPTIONS</SText>
            <OptionCardChecker
              initialChecked={false}
              label="Merge chapters"
              text="Combine multiple chapters into a single volume"
              onChange={(checked) => {}}
            />
            <OptionCardChecker
              initialChecked={false}
              label="Delete source"
              text="Remove original after successful upload"
              onChange={(checked) => {}}
            />
          </View>
        </View>

        <SButton
          onPress={() => {}} //TODO: Send
          style={{
            backgroundColor: colors.primary_container,
            paddingVertical: 12,
            alignItems: 'center',
            justifyContent: 'center',
            borderRadius: 12,
            boxShadow: colors.boxShadow,
          }}
        >
          <SText style={{ fontFamily: 'semibold', color: colors.on_primary }}>Send</SText>
        </SButton>
      </View>
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
