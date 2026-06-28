import React, { useState } from 'react';
import { StyleSheet, View } from 'react-native';
import SText from '../src/components/shared/SText';
import { router, Stack } from 'expo-router';
import { colors } from '../src/theme/colors';
import SourceSelector from '../src/components/senders/source-selector';
import DestinationSelector from '../src/components/senders/destination-selector';
import OptionCardChecker from '../src/components/senders/option-card-checker';
import SButton from '../src/components/shared/SButton';
import { useQueue } from '../src/hooks/useQueue';
import { TransactionRequest } from '../src/models/transaction-request';
import SSelect from '../src/components/shared/SSelect';
import { useSettings } from '../src/hooks/useSettings';
import { useShallow } from 'zustand/react/shallow';
import { eReaderProfiles } from '../src/constants';

export default function SendBookPage() {
  const send = useQueue((s) => s.send);
  const { model, setModel } = useSettings(
    useShallow((s) => ({ model: s.model, setModel: s.setModel }))
  );
  const [req, setReq] = useState<TransactionRequest>({
    deleteOrigin: false,
    merge: false,
    destination: 'local',
    mode: 'no-select',
    sources: [],
    author: '',
    title: '',
  });

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
            <SText style={styles.title}>SOURCE</SText>
            <SourceSelector
              initSources={req.sources}
              onChange={(srcs) => setReq((s) => ({ ...s, sources: srcs }))}
              onModeChange={(mode) => setReq((s) => ({ ...s, mode: mode }))}
            />
          </View>

          <View>
            <SText style={styles.title}>READER MODEL</SText>
            <SSelect
              value={model}
              options={eReaderProfiles}
              onOptionChange={(opt) => setModel(opt.value)}
            />
          </View>

          <View>
            <SText style={styles.title}>DESTINATION</SText>
            <DestinationSelector
              initDestination={req.destination}
              onChange={(dest) => setReq((s) => ({ ...s, destination: dest }))}
            />
          </View>

          <View>
            <SText style={styles.title}>OPTIONS</SText>
            <OptionCardChecker
              initialChecked={req.deleteOrigin}
              label="Delete source"
              text="Remove original after successful upload"
              onChange={(checked) => setReq((s) => ({ ...s, deleteOrigin: checked }))}
            />
          </View>
        </View>

        <SButton
          onPress={async () => {
            const done = await send(req);
            if (done) router.navigate('/(tabs)/queue');
          }}
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
    marginBottom: 5,
  },
});
