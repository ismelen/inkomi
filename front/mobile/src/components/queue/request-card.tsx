import React, { useState } from 'react';
import { Pressable, View } from 'react-native';
import { QueueTime } from '../../models/queue';
import { colors } from '../../theme/colors';
import DeleteForeverIcon from '../icons/delete-forever-icon';
import DocsIcon from '../icons/docs-icon';
import DownloadIcon from '../icons/download-icon';
import SProgressTimeBar from '../shared/SProgressTimeBar';
import SText from '../shared/SText';
interface Props {
  data: QueueTime;
  onDelete?(): void;
  onDownload?(): void;
  onComplete?(): void;
}

export function RequestCard({ data, onDelete, onDownload, onComplete }: Props) {
  const [showProgressBar, setShowProgressBar] = useState(
    data.endTime && data.endTime?.getTime() > Date.now()
  );

  return (
    <Pressable
      onPress={() => {
        if (!showProgressBar) {
          if (data.endTime) return onDownload?.();
          return onDelete?.();
        }
      }}
    >
      <View style={{ flexDirection: 'row', gap: 8, alignItems: 'center' }}>
        <DocsIcon size="24px" color={colors.onCard} />
        <SText numberOfLines={2} style={{ fontWeight: 500, fontSize: 16, flex: 1 }}>
          {data.path.split('/').pop()}
        </SText>
        {showProgressBar && (
          <SText style={{ color: colors.primary, fontSize: 14, opacity: 0.8 }}>
            Ends at ~
            {data.endTime!.toLocaleTimeString(undefined, {
              hour: '2-digit',
              minute: '2-digit',
            })}
          </SText>
        )}
        {!showProgressBar && data.endTime && <DownloadIcon size="24px" color={colors.primary} />}
        {!showProgressBar && !data.endTime && (
          <DeleteForeverIcon size="24px" color={colors.primary} />
        )}
      </View>
      {showProgressBar && (
        <SProgressTimeBar
          endTime={data.endTime!}
          style={{ marginTop: 14 }}
          onFinish={() => {
            setShowProgressBar(false);
            onComplete?.();
          }}
        />
      )}
    </Pressable>
  );
}

function formatTime(date: Date): string {
  const minutes = date.getMinutes().toString().padStart(2, '0');
  const seconds = date.getSeconds().toString().padStart(2, '0');

  return `${minutes}m ${seconds}s`;
}
