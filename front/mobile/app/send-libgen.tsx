import React from 'react';
import { useShallow } from 'zustand/react/shallow';
import { colors } from '../src/theme/colors';
import SIcon from '../src/components/icons/SIcon';
import { StyleSheet, View } from 'react-native';
import { useSafeAreaInsets } from 'react-native-safe-area-context';
import SText from '../src/components/shared/SText';
import SSelect from '../src/components/shared/SSelect';
import { eReaderProfiles } from '../src/constants';
import DestinationSelector from '../src/components/senders/destination-selector';
import SButton from '../src/components/shared/SButton';
import { ScrollView } from 'react-native-gesture-handler';
import SearchedBookCard from '../src/components/search/searched-book-card';
import { useLibgen } from '../src/hooks/useLibgen';
import LoadingScreen from '../src/components/shared/loading-screen';
import { useSender } from '../src/hooks/useSender';
import { Stack, usePathname } from 'expo-router';
import { useTranslation } from 'react-i18next';
import { useCloud } from '../src/hooks/useCloud';
import SConfirmDialog from '../src/components/shared/SConfirmDialog';

export default function SendLibgen() {
  const insets = useSafeAreaInsets();
  const { sending, config, setConfig, send, clear: clearSender } = useSender('md5');
  const { t } = useTranslation();
  const { oauth, folder, showAuthConfirm, resolveAuthConfirm } = useCloud(
    useShallow((s) => ({
      oauth: s.oauth,
      folder: s.folder,
      showAuthConfirm: s.showAuthConfirm,
      resolveAuthConfirm: s.resolveAuthConfirm,
    }))
  );

  const isSendDisabled = config.toCloud ? !oauth?.email || !folder : false;

  const { selectedBooks, onDelete, clear, setBooks } = useLibgen(
    useShallow((s) => ({
      selectedBooks: s.selected,
      onDelete: s.selectBook,
      clear: s.clear,
      setBooks: s.setBooks,
    }))
  );

  React.useEffect(() => {
    if (config.files && config.files.length > 0 && Object.keys(selectedBooks).length === 0) {
      setBooks(config.files);
    }
  }, [config.files]);

  if (sending)
    return <LoadingScreen title={t('loading.sendingBook')} subtitle={t('loading.optimizing')} />;

  return (
    <>
      <Stack.Screen
        options={{
          headerShown: true,
          title: t('sendLibgen.title'),
          headerTitleStyle: { fontFamily: 'semibold', fontSize: 20, color: colors.on_background },
          headerTitleAlign: 'center',
          headerStyle: {
            backgroundColor: colors.background,
          },
          headerTintColor: colors.primary,
          headerRight: () => (
            <SButton
              onPress={() => {
                clearSender();
                clear();
              }}
            >
              <SIcon name="delete" size={24} color={colors.primary} type="outlined" />
            </SButton>
          ),
        }}
      />
      <View style={{ flex: 1, paddingBottom: 24 + insets.bottom }}>
        <ScrollView style={{ flex: 1 }} contentContainerStyle={{ paddingHorizontal: 24, gap: 32 }}>
          <View style={styles.section}>
            <SText style={styles.title}>{t('sendLibgen.books')}</SText>
            {Object.values(selectedBooks).map((e) => (
              <SearchedBookCard
                key={e.src}
                book={e}
                onSelect={() => onDelete(e)}
                selected={false}
                deleteMode
              />
            ))}
          </View>

          <View>
            <SText style={styles.title}>{t('sendLibgen.readerModel')}</SText>
            <SSelect
              value={config.model}
              options={eReaderProfiles}
              onOptionChange={(opt) => setConfig((s) => ({ ...s, model: opt.value }))}
            />
          </View>

          <View style={styles.section}>
            <SText style={styles.title}>{t('sendLibgen.destination')}</SText>
            <DestinationSelector
              toCloud={config.toCloud}
              onChange={(toCloud) => setConfig((s) => ({ ...s, toCloud: toCloud }))}
            />
          </View>
        </ScrollView>

        <SButton
          onPress={async () => {
            const finalFiles = Object.values(selectedBooks);
            setConfig((s) => ({ ...s, files: finalFiles }));
            if (await send(finalFiles)) clear();
          }}
          disabled={isSendDisabled}
          style={{
            backgroundColor: isSendDisabled ? colors.surface_variant : colors.primary_container,
            marginHorizontal: 24,
            paddingVertical: 12,
            alignItems: 'center',
            justifyContent: 'center',
            borderRadius: 12,
            boxShadow: colors.boxShadow,
          }}
        >
          <SText
            style={{
              fontFamily: 'semibold',
              color: isSendDisabled ? colors.on_surface_variant : colors.on_primary,
            }}
          >
            {t('sendLibgen.send')}
          </SText>
        </SButton>
      </View>
      <SConfirmDialog
        visible={showAuthConfirm}
        title={t('common.authConfirmTitle')}
        message={t('common.authConfirmMessage')}
        confirmText={t('common.authConfirmOk')}
        cancelText={t('common.cancel')}
        onCancel={() => resolveAuthConfirm(false)}
        onConfirm={() => resolveAuthConfirm(true)}
      />
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
