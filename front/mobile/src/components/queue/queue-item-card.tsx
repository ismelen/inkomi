import { useEffect, useRef } from 'react';
import { View } from 'react-native';
import { useQueue } from '../../hooks/useQueue';
import { QueueElement } from '../../models/queue-element';
import { Destination } from '../../models/transaction-request';
import { colors } from '../../theme/colors';
import SIcon from '../icons/SIcon';
import SButton from '../shared/SButton';
import SText from '../shared/SText';
import PulseProgressBar from './pulse-progress-bar';

interface Props {
  data: QueueElement;
  autoCheck?: boolean;
  idx: number;
}

export default function QueueItemCard({ data, idx, autoCheck = false }: Props) {
  const checkProgress = useQueue((s) => s.checkProgress);
  const intervalRef = useRef<number | undefined>(undefined);

  useEffect(() => {
    if (!autoCheck) return;

    const run = async () => {
      const done = await checkProgress(idx, data.id);
      if (done) clearInterval(intervalRef.current);
    };

    run();
    intervalRef.current = setInterval(run, 2000);

    return () => clearInterval(intervalRef.current);
  }, [autoCheck, data.id]);

  return (
    <View
      style={{
        backgroundColor: colors.surface_container_lowest,
        borderRadius: 12,
        boxShadow: colors.boxShadow,
        padding: 10,
        gap: 4,
        borderColor: colors.error,
        borderWidth: data.error ? 1 : 0,
      }}
    >
      <View
        style={{ flexDirection: 'row', justifyContent: 'space-between', alignItems: 'flex-start' }}
      >
        <SText style={{ fontSize: 16, flex: 1, flexShrink: 1 }}>{data.title}</SText>
        <DestinationIndicator dest={data.destination} />
      </View>

      {data.error && <ErrorMessage error={data.error} />}
      {!data.error && <LoadingSection data={data} idx={idx} />}
    </View>
  );
}

function LoadingSection({ data, idx }: { data: QueueElement; idx: number }) {
  const download = useQueue((s) => s.download);

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
          onPress={() => download(idx, data.id)}
          style={{
            marginTop: 8,
            backgroundColor: colors.primary_container,
            padding: 10,
            alignItems: 'center',
            justifyContent: 'center',
            borderRadius: 6,
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
