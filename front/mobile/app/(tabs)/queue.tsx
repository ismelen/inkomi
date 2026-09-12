import { useShallow } from 'zustand/react/shallow';
import { StyleSheet, View, SectionList, ActivityIndicator } from 'react-native';
import { useSafeAreaInsets } from 'react-native-safe-area-context';
import { useQueue } from '../../src/hooks/useQueue';
import SIcon from '../../src/components/icons/SIcon';
import { colors } from '../../src/theme/colors';
import QueueItemCard from '../../src/components/queue/queue-item-card';
import SText from '../../src/components/shared/SText';
import SButton from '../../src/components/shared/SButton';
import { router } from 'expo-router';
import { useTranslation } from 'react-i18next';
import { useState, useEffect, useMemo } from 'react';
import { Transaction } from '../../src/models/transaction';

export default function QueuePage() {
  const insets = useSafeAreaInsets();
  const [isReady, setIsReady] = useState(false);

  useEffect(() => {
    const timeout = setTimeout(() => {
      setIsReady(true);
    }, 300);
    return () => clearTimeout(timeout);
  }, []);

  const { transactions, loadMore } = useQueue(
    useShallow((s: any) => ({
      transactions: s.transactions as Transaction[],
      loadMore: s.loadMore,
    }))
  );
  const { t } = useTranslation();

  const groupedSections = useMemo(() => {
    const today = new Date();
    today.setHours(0, 0, 0, 0);
    const startOfToday = today.getTime();
    const startOfYesterday = startOfToday - 86400000;
    const startOfThisWeek = startOfToday - 7 * 86400000;
    const startOfThisMonth = startOfToday - 30 * 86400000;

    const groups: { titleKey: string; data: { item: Transaction; globalIndex: number }[] }[] = [
      { titleKey: 'queue.sections.today', data: [] },
      { titleKey: 'queue.sections.yesterday', data: [] },
      { titleKey: 'queue.sections.thisWeek', data: [] },
      { titleKey: 'queue.sections.thisMonth', data: [] },
      { titleKey: 'queue.sections.older', data: [] },
    ];

    transactions.forEach((t, i) => {
      const msTimestamp = t.timestamp * 1000;
      if (msTimestamp >= startOfToday) {
        groups[0].data.push({ item: t, globalIndex: i });
      } else if (msTimestamp >= startOfYesterday) {
        groups[1].data.push({ item: t, globalIndex: i });
      } else if (msTimestamp >= startOfThisWeek) {
        groups[2].data.push({ item: t, globalIndex: i });
      } else if (msTimestamp >= startOfThisMonth) {
        groups[3].data.push({ item: t, globalIndex: i });
      } else {
        groups[4].data.push({ item: t, globalIndex: i });
      }
    });

    return groups.filter((g) => g.data.length > 0);
  }, [transactions]);

  const renderEmpty = () => (
    <View
      style={{
        alignItems: 'center',
        gap: 35,
        flex: 1,
        justifyContent: 'center',
        marginTop: 50,
      }}
    >
      <View style={{ backgroundColor: colors.primary_fixed, borderRadius: 999, padding: 15 }}>
        <SIcon name="cloud_off" color={colors.primary} size={60} type="outlined" />
      </View>
      <View>
        <SText style={{ fontSize: 16, textAlign: 'center', fontFamily: 'semibold' }}>
          {t('queue.empty')}
        </SText>
        <SText style={{ fontSize: 16, textAlign: 'center' }}>{t('queue.emptySubtitle1')}</SText>
        <SText style={{ fontSize: 16, textAlign: 'center' }}>{t('queue.emptySubtitle2')}</SText>
      </View>
      <SButton
        onPress={() => router.navigate('/(tabs)/')}
        style={{
          backgroundColor: colors.primary_container,
          paddingVertical: 14,
          paddingHorizontal: 30,
          borderRadius: 12,
        }}
      >
        <SText style={{ color: colors.on_primary, fontFamily: 'semibold', fontSize: 16 }}>
          {t('queue.sendSomething')}
        </SText>
      </SButton>
    </View>
  );

  if (!isReady) {
    return (
      <View style={{ flex: 1, paddingHorizontal: 24 }}>
        <SText style={{ fontFamily: 'bold', fontSize: 28, marginBottom: 16 }}>
          {t('queue.title')}
        </SText>
        <View style={{ flex: 1, justifyContent: 'center', alignItems: 'center' }}>
          <ActivityIndicator size="large" color={colors.primary} />
        </View>
      </View>
    );
  }

  return (
    <View style={{ flex: 1, paddingHorizontal: 24 }}>
      <SText style={{ fontFamily: 'bold', fontSize: 28, marginBottom: 16 }}>
        {t('queue.title')}
      </SText>
      <SectionList
        sections={groupedSections}
        keyExtractor={(item) => item.item.id}
        renderItem={({ item }) => <QueueItemCard data={item.item} idx={item.globalIndex} />}
        renderSectionHeader={({ section: { titleKey } }) => (
          <SText
            style={{
              fontFamily: 'bold',
              fontSize: 18,
              marginVertical: 12,
              color: colors.on_surface,
            }}
          >
            {t(titleKey)}
          </SText>
        )}
        ListEmptyComponent={renderEmpty}
        contentContainerStyle={{ gap: 10, paddingBottom: 16 + insets.bottom }}
        showsVerticalScrollIndicator={false}
        onEndReached={loadMore}
        onEndReachedThreshold={0.5}
      />
    </View>
  );
}
