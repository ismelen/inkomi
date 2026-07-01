import { useShallow } from 'zustand/react/shallow';
import { useQueue } from '../src/hooks/useQueue';
import { useSettings } from '../src/hooks/useSettings';
import { useState } from 'react';
import { TransactionRequest } from '../src/models/transaction-request';
import { router, Stack } from 'expo-router';
import { colors } from '../src/theme/colors';
import { StyleSheet, View } from 'react-native';
import SText from '../src/components/shared/SText';
import SSelect from '../src/components/shared/SSelect';
import { eReaderProfiles } from '../src/constants';
import DestinationSelector from '../src/components/senders/destination-selector';
import OptionCardChecker from '../src/components/senders/option-card-checker';
import SButton from '../src/components/shared/SButton';
import { ScrollView } from 'react-native-gesture-handler';
import SearchedBookCard from '../src/components/search/searched-book-card';
import { useLibgen } from '../src/hooks/useLibgen';
import { LibgenTransactionRequest } from '../src/models/libgen-transaction-request';

export default function SendLibgen() {
  const send = useQueue((s) => s.send);
  const { model, setModel } = useSettings(
    useShallow((s) => ({ model: s.model, setModel: s.setModel }))
  );
  const [req, setReq] = useState<LibgenTransactionRequest>({
    deleteOrigin: false,
    merge: false,
    destination: 'local',
    mode: 'no-select',
    sources: [],
    author: '',
    title: '',
    books: [],
  });

  const { selectedBooks, onDelete } = useLibgen(
    useShallow((s) => ({
      selectedBooks: s.selected,
      onDelete: s.selectBook,
    }))
  );

  return (
    <>
      <Stack.Screen
        options={{
          headerShown: true,
          title: 'Send Books',
          headerTitleStyle: { fontFamily: 'semibold', fontSize: 20, color: colors.on_background },
          headerTitleAlign: 'center',
          headerStyle: {
            backgroundColor: colors.background,
          },
          headerTintColor: colors.primary,
        }}
      />
      <View style={{ flex: 1, paddingBottom: 24, paddingHorizontal: 24 }}>
        <ScrollView style={{ flex: 1, gap: 32 }}>
          <View style={styles.section}>
            <SText style={styles.title}>BOOKS</SText>
            {Object.values(selectedBooks).map((e) => (
              <SearchedBookCard
                key={e.md5}
                book={e}
                onSelect={() => onDelete(e)}
                selected={false}
                deleteMode
              />
            ))}
          </View>

          <View style={styles.section}>
            <SText style={styles.title}>READER MODEL</SText>
            <SSelect
              value={model}
              options={eReaderProfiles}
              onOptionChange={(opt) => setModel(opt.value)}
            />
          </View>

          <View style={styles.section}>
            <SText style={styles.title}>DESTINATION</SText>
            <DestinationSelector
              initDestination={req.destination}
              onChange={(dest) => setReq((s) => ({ ...s, destination: dest }))}
            />
          </View>
        </ScrollView>

        <SButton
          onPress={async () => {
            req.books = Object.values(selectedBooks);
            if (req.books.length === 0) return;

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
  section: {
    marginBottom: 32,
  },
});
