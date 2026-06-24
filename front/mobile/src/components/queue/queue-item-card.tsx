import { View } from 'react-native';
import { colors } from '../../theme/colors';
import { QueueElement } from '../../models/queue-element';
import SText from '../shared/SText';
import { Destination } from '../../models/transaction-request';
import SIcon from '../icons/SIcon';
import PulseProgressBar from './pulse-progress-bar';
import SButton from '../shared/SButton';
import { useQueue } from '../../hooks/useQueue';
import { useEffect } from 'react';

interface Props {
  data: QueueElement;
  idx: number;
}

export default function QueueItemCard({ data, idx }: Props) {
  const checkProgress = useQueue((s) => s.checkProgress);

  useEffect(() => {
    checkProgress(idx, data.id);
    const interval = setInterval(() => checkProgress(idx, data.id), 2000);

    return () => clearInterval(interval);
  }, []);

  return (
    <View
      style={{
        backgroundColor: colors.surface_container_lowest,
        borderRadius: 12,
        boxShadow: colors.boxShadow,
        padding: 16,
        gap: 4,
        borderColor: colors.error,
        borderWidth: data.error ? 1 : 0,
      }}
    >
      <View style={{ flexDirection: 'row', justifyContent: 'space-between', alignItems: 'center' }}>
        <SText style={{ fontSize: 16 }}>{data.title}</SText>
        <DestinationIndicator dest={data.destination} />
      </View>

      {data.error && <ErrorMessage error={data.error} />}
      {!data.error && <LoadingSection data={data} />}
    </View>
  );
}

function LoadingSection({ data }: { data: QueueElement }) {
  return (
    <>
      {!data.error && data.progress === 100 ? (
        <View style={{ flexDirection: 'row', alignItems: 'center', gap: 4 }}>
          <SIcon name="check_circle" color={colors.ok} size={16} />
          <SText style={{ fontSize: 14, fontFamily: 'medium', color: colors.ok }}>Completed</SText>
        </View>
      ) : (
        <View style={{ gap: 8 }}>
          <View style={{ flexDirection: 'row', alignItems: 'center', gap: 4 }}>
            <SIcon name="autorenew" color={colors.primary} size={16} />
            <SText style={{ fontSize: 14, color: colors.primary }}>
              Converting ({data.progress}%)
            </SText>
          </View>
          <PulseProgressBar progress={data.progress / 100} />
        </View>
      )}

      {data.progress === 100 && data.destination === 'local' && (
        <SButton
          onPress={() => {}} //TODO: download
          style={{
            marginTop: 8,
            backgroundColor: colors.primary_container,
            padding: 10,
            alignItems: 'center',
            justifyContent: 'center',
            borderRadius: 12,
            flexDirection: 'row',
            gap: 8,
          }}
        >
          <SIcon name="download" color={colors.on_primary} size={24} type="outlined" />
          <SText style={{ fontFamily: 'semibold', color: colors.on_primary, fontSize: 16 }}>
            Download
          </SText>
        </SButton>
      )}
    </>
  );
}

function ErrorMessage({ error }: { error: string }) {
  return (
    <View style={{ flexDirection: 'row', alignItems: 'center', gap: 4 }}>
      <SIcon name="info" color={colors.error} size={16} type="outlined" />
      <SText style={{ fontSize: 14, fontFamily: 'medium', color: colors.error }}>{error}</SText>
    </View>
  );
}

function DestinationIndicator({ dest }: { dest: Destination }) {
  return (
    <View
      style={{
        flexDirection: 'row',
        alignItems: 'center',
        gap: 4,
        paddingVertical: 4,
        paddingHorizontal: 8,
        borderRadius: 15,
        backgroundColor: colors.surface_container,
        overflow: 'hidden',
      }}
    >
      <SIcon name="cloud" color={colors.outline} size={15} type="outlined" />
      <SText style={{ color: colors.outline, fontSize: 10 }}>
        {dest[0].toUpperCase() + dest.slice(1)}
      </SText>
    </View>
  );
}
