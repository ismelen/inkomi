import React, { useState } from 'react';
import { Pressable, ScrollView, View } from 'react-native';
import { useTranslation } from 'react-i18next';
import { colors, radii, spacing } from '../../theme';
import { useList } from '../../stores/use-list';
import { SModal } from '../ui/s-modal';
import { SText } from '../ui/s-text';
import { STextInput } from '../ui/s-text-input';

interface Props {
  visible: boolean;
  count: number;
  onClose: () => void;
}

const ungrouped = '__ungrouped__';
const newGroup = '__new__';

function Choice({
  label,
  detail,
  selected,
  onPress,
}: {
  label: string;
  detail?: string;
  selected: boolean;
  onPress: () => void;
}) {
  return (
    <Pressable
      onPress={onPress}
      style={{
        flexDirection: 'row',
        alignItems: 'center',
        padding: spacing.md,
        borderRadius: radii.md,
        borderWidth: 1,
        borderColor: selected ? colors.text : colors.border,
        backgroundColor: selected ? colors.surface : colors.background,
      }}
    >
      <SText variant="label" style={{ flex: 1 }} numberOfLines={1}>
        {label}
      </SText>
      {!!detail && (
        <SText variant="small" tone="textSecondary">
          {detail}
        </SText>
      )}
    </Pressable>
  );
}

export function MoveModal({ visible, count, onClose }: Props) {
  const { t } = useTranslation();
  const groups = useList((s) => s.groups);
  const rootOrder = useList((s) => s.rootOrder);
  const selection = useList((s) => s.selection);
  const { move, createGroup } = useList.getState();
  const [target, setTarget] = useState<string>(ungrouped);
  const [name, setName] = useState('');

  const close = () => {
    setTarget(ungrouped);
    setName('');
    onClose();
  };

  const confirm = () => {
    let groupId: string | null = target === ungrouped ? null : target;
    if (target === newGroup) groupId = createGroup(name.trim() || t('group.defaultTitle'));
    move(selection, groupId);
    close();
  };

  const groupIds = rootOrder.filter((id) => groups[id] && !selection.includes(id));

  return (
    <SModal
      visible={visible}
      title={t('move.title', { count })}
      onClose={close}
      actions={[
        { label: t('common.cancel'), onPress: close, variant: 'secondary' },
        { label: t('common.move'), onPress: confirm },
      ]}
    >
      <ScrollView style={{ maxHeight: 320 }} contentContainerStyle={{ gap: spacing.sm }}>
        <Choice
          label={t('move.ungrouped')}
          selected={target === ungrouped}
          onPress={() => setTarget(ungrouped)}
        />
        {groupIds.map((id) => (
          <Choice
            key={id}
            label={groups[id].title}
            detail={t('common.files', { count: groups[id].itemIds.length })}
            selected={target === id}
            onPress={() => setTarget(id)}
          />
        ))}
        <Choice
          label={t('move.newGroup')}
          selected={target === newGroup}
          onPress={() => setTarget(newGroup)}
        />
        {target === newGroup && (
          <View>
            <STextInput
              autoFocus
              value={name}
              onChangeText={setName}
              placeholder={t('move.newGroupPlaceholder')}
            />
          </View>
        )}
      </ScrollView>
    </SModal>
  );
}
