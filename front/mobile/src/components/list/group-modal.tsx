import React, { useEffect, useState } from 'react';
import { router } from 'expo-router';
import { useTranslation } from 'react-i18next';
import { Group } from '../../models/group';
import { useList } from '../../stores/use-list';
import { SettingRow } from '../ui/setting-row';
import { SModal } from '../ui/s-modal';
import { SSegmented } from '../ui/s-segmented';
import { SSwitch } from '../ui/s-switch';
import { SText } from '../ui/s-text';
import { STextInput } from '../ui/s-text-input';

interface Props {
  group: Group | null;
  onClose: () => void;
}

export function GroupModal({ group, onClose }: Props) {
  const { t } = useTranslation();
  const updateGroup = useList((s) => s.updateGroup);
  const [title, setTitle] = useState('');
  const [join, setJoin] = useState(false);
  const [rtl, setRtl] = useState(false);

  useEffect(() => {
    if (!group) return;
    setTitle(group.title);
    setJoin(group.join);
    setRtl(group.rtl);
  }, [group]);

  const save = () => {
    if (!group) return;
    updateGroup(group.id, { title: title.trim() || group.title, join, rtl });
    onClose();
  };

  return (
    <SModal
      visible={!!group}
      title={t('group.title')}
      onClose={onClose}
      actions={[
        { label: t('common.cancel'), onPress: onClose, variant: 'secondary' },
        { label: t('common.save'), onPress: save },
      ]}
    >
      <STextInput label={t('common.title')} value={title} onChangeText={setTitle} />
      <SettingRow
        label={t('group.join')}
        right={<SSwitch value={join} onValueChange={setJoin} />}
      />
      <SText variant="section">{t('common.readingDirection')}</SText>
      <SSegmented
        value={rtl}
        onChange={setRtl}
        options={[
          { value: false, label: t('common.ltr') },
          { value: true, label: t('common.rtl') },
        ]}
      />
      <SettingRow
        label={t('group.mangaOptions')}
        onPress={() => {
          if (!group) return;
          save();
          router.push({ pathname: '/options/[id]', params: { id: group.id } });
        }}
      />
    </SModal>
  );
}
