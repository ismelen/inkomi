import React from 'react';
import { StyleSheet, Text, View } from 'react-native';
import { useQueue } from '../../src/hooks/useQueue';
import SIcon from '../../src/components/icons/SIcon';
import { colors } from '../../src/theme/colors';
import QueueItemCard from '../../src/components/queue/queue-item-card';

export default function QueuePage() {
  const transactions = useQueue((s) => s.transactions);
  const completedTransactions = useQueue((s) => s.completedTransactions);

  return (
    <View style={{ flex: 1 }}>
      <Text style={{ fontFamily: 'bold', fontSize: 28 }}>Transaction queue</Text>
      <View style={styles.section}>
        <SIcon name="pending_actions" color={colors.primary} size={24} />
        <Text style={styles.label}>ACTIVE</Text>
      </View>

      <View style={{ marginTop: 16, gap: 10 }}>
        {transactions.map((e, i) => (
          <QueueItemCard key={e.id} data={e} idx={i} autoCheck />
        ))}
      </View>

      <View style={[styles.section, { marginTop: 20 }]}>
        <SIcon name="check_circle" color={colors.ok} size={24} type="outlined" />
        <Text style={styles.label}>COMPLETED</Text>
      </View>

      <View style={{ marginTop: 16, gap: 10 }}>
        {completedTransactions.map((e, i) => (
          <QueueItemCard key={e.id} data={e} idx={i} />
        ))}
      </View>
    </View>
  );
}

const styles = StyleSheet.create({
  section: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: 5,
    marginTop: 16,
  },
  label: { fontSize: 14, fontFamily: 'semibold', color: colors.on_surface_variant },
});
