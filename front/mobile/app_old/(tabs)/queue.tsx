import React, { useEffect, useMemo, useState } from 'react';
import { StyleSheet, View } from 'react-native';
import { RequestCard } from '../../src/components/queue/request-card';
import SColumn from '../../src/components/shared/SColumn';
import SText from '../../src/components/shared/SText';
import { useQueue } from '../../src/hooks/use-queue';
import { QueueTime } from '../../src/models/queue';
import { colors } from '../../src/theme/colors';

interface RequestOroder {
  rIdx: number;
  tIdx: number;
  request: QueueTime;
}

export default function queue() {
  const requests = useQueue((s) => s.requests);
  const storedRequests: RequestOroder[] = useMemo(() => {
    return requests
      .map((r, rIdx) =>
        r.times.map((t, tIdx) => ({
          rIdx: rIdx,
          tIdx: tIdx,
          request: t,
        }))
      )
      .flat();
  }, [requests]);

  const deleteRequest = useQueue((s) => s.delete);
  const download = useQueue((s) => s.download);

  const [ongoing, setOngoing] = useState<RequestOroder[]>([]);
  const [completed, setCompleted] = useState<RequestOroder[]>([]);
  const [requestsCompleted, setRequestsCompleted] = useState(0);

  useEffect(() => {
    useQueue.getState().init();
  }, []);

  useEffect(() => {
    const now = Date.now();

    const ongonigReqs: RequestOroder[] = [];
    const completedReqs: RequestOroder[] = [];

    for (let req of storedRequests) {
      if ((req.request.endTime?.getTime() ?? 0) > now) {
        ongonigReqs.push(req);
      } else {
        completedReqs.push(req);
      }
    }

    setOngoing(ongonigReqs);
    setCompleted(completedReqs);
  }, [storedRequests, requestsCompleted]);

  return (
    <View style={[styles.padding]}>
      <SText style={[styles.title, { fontSize: 20 }]}>Uploads</SText>
      {requests.length === 0 && (
        <View>
          <SText>No pending uploads</SText>
        </View>
      )}

      {ongoing.length !== 0 && <SText style={styles.sectionTitle}>ONGOING</SText>}
      <SColumn>
        {ongoing.map((e) => (
          <RequestCard
            key={e.request.path}
            data={e.request}
            onComplete={() => {
              setRequestsCompleted((v) => v + 1);
            }}
          />
        ))}
      </SColumn>
      {completed.length !== 0 && <SText style={styles.sectionTitle}>COMPLETED</SText>}
      <SColumn>
        {completed.map((e) => (
          <RequestCard
            key={e.request.path}
            data={e.request}
            onDelete={() => deleteRequest(e.rIdx, e.tIdx)}
            onDownload={() => download(e.rIdx, e.tIdx)}
          />
        ))}
      </SColumn>
    </View>
  );
}

const styles = StyleSheet.create({
  padding: {
    paddingHorizontal: 20,
    paddingVertical: 10,
  },
  title: {
    fontWeight: '500',
  },
  sectionTitle: { color: colors.onCard, marginBottom: 10, marginTop: 20 },
});
