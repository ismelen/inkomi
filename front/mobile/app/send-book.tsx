import React from 'react';
import { View } from 'react-native';
import SText from '../src/components/shared/SText';
import { Stack } from 'expo-router';
import { colors } from '../src/theme/colors';
import SourceSelector from '../src/components/senders/source-selector';
import DestinationSelector from '../src/components/senders/destination-selector';
import OptionCardChecker from '../src/components/senders/option-card-checker';
import SButton from '../src/components/shared/SButton';

export default function SendBookPage() {
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
      <View style={{ flex: 1, paddingBottom: 24, paddingHorizontal: 24 }}>
        <View style={{ flex: 1, gap: 32 }}>
          <View>
            <SText
              style={{
                fontFamily: 'semibold',
                fontSize: 14,
                color: colors.on_surface_variant,
                marginBottom: 5,
              }}
            >
              SOURCE
            </SText>
            <SourceSelector initSources={[]} onChange={(srcs) => {}} />
          </View>

          <View>
            <SText
              style={{
                fontFamily: 'semibold',
                fontSize: 14,
                color: colors.on_surface_variant,
                marginBottom: 5,
              }}
            >
              DESTINATION
            </SText>
            <DestinationSelector initDestination="local" onChange={(dest) => {}} />
          </View>

          <View>
            <SText
              style={{
                fontFamily: 'semibold',
                fontSize: 14,
                color: colors.on_surface_variant,
                marginBottom: 5,
              }}
            >
              OPTIONS
            </SText>
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
