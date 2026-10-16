import React from 'react';
import { FlatList, View } from 'react-native';
import { useTranslation } from 'react-i18next';
import { ActivityCard } from '../../src/components/activity/activity-card';
import { Header } from '../../src/components/ui/header';
import { Screen } from '../../src/components/ui/screen';
import { SText } from '../../src/components/ui/s-text';
import { useActivity } from '../../src/stores/use-activity';
import { colors, radii, spacing } from '../../src/theme';

export default function ActivityScreen() {
  const { t } = useTranslation();
  const queue = useActivity((s) => s.queue);

  return (
    <Screen bottomInset>
      <Header title={t('activity.title')} />
      <FlatList
        data={queue}
        keyExtractor={(tree) => tree.root.id}
        renderItem={({ item }) => <ActivityCard tree={item} />}
        contentContainerStyle={{ padding: spacing.lg, gap: spacing.md, flexGrow: 1 }}
        ListEmptyComponent={
          <View style={{ flex: 1, alignItems: 'center', justifyContent: 'center' }}>
            <SText tone="textSecondary">{t('activity.empty')}</SText>
          </View>
        }
      />
      <View
        style={{
          flexDirection: 'row',
          alignItems: 'center',
          justifyContent: 'center',
          gap: spacing.xs,
          padding: spacing.md,
        }}
      >
        <View
          style={{ width: 6, height: 6, borderRadius: radii.pill, backgroundColor: colors.success }}
        />
        <SText variant="small" tone="textSecondary">
          {t('activity.realtime')}
        </SText>
      </View>
    </Screen>
  );
}
