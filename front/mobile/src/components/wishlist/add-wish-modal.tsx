import React, { useEffect, useState } from 'react';
import { ActivityIndicator, FlatList, Pressable, View } from 'react-native';
import { useTranslation } from 'react-i18next';
import { AddWishRequest, WishKind, WishSuggestion } from '../../api/types/wishlist';
import { useCatalogSuggestions } from '../../hooks/use-catalog-suggestions';
import { colors, radii, spacing } from '../../theme';
import { BookCover } from '../books/book-cover';
import { SModal } from '../ui/s-modal';
import { SSegmented } from '../ui/s-segmented';
import { SText } from '../ui/s-text';
import { STextInput } from '../ui/s-text-input';

interface Props {
  visible: boolean;
  /** Kind selected when the modal opens; the user can switch it. */
  initialKind: WishKind;
  onClose: () => void;
  onAdd: (item: AddWishRequest) => void;
}

/** Adds a wish: suggestions from the catalog while typing, or the typed text as is. */
export function AddWishModal({ visible, initialKind, onClose, onAdd }: Props) {
  const { t } = useTranslation();
  const [kind, setKind] = useState<WishKind>(initialKind);
  const [query, setQuery] = useState('');
  const { suggestions, loading, lookup, clear } = useCatalogSuggestions();

  useEffect(() => {
    if (!visible) return;
    setKind(initialKind);
    setQuery('');
    clear();
  }, [visible, initialKind, clear]);

  const onChange = (text: string) => {
    setQuery(text);
    lookup(kind, text);
  };

  const onKind = (next: WishKind) => {
    setKind(next);
    clear();
    lookup(next, query);
  };

  const pick = (s: WishSuggestion) =>
    onAdd(
      s.kind === 'author'
        ? { kind: 'author', title: s.title, authorKey: s.authorKey }
        : { kind: 'book', title: s.title, author: s.author, workKey: s.workKey }
    );

  const text = query.trim();

  return (
    <SModal
      visible={visible}
      title={t('wishlist.addTitle')}
      onClose={onClose}
      actions={[
        { label: t('common.close'), onPress: onClose, variant: 'secondary' },
        {
          label: t('wishlist.add'),
          onPress: () => onAdd({ kind, title: text }),
          disabled: text.length === 0,
        },
      ]}
    >
      <SSegmented
        value={kind}
        onChange={onKind}
        options={[
          { value: 'book', label: t('wishlist.book') },
          { value: 'author', label: t('wishlist.author') },
        ]}
      />
      <STextInput
        autoFocus
        value={query}
        onChangeText={onChange}
        placeholder={t(
          kind === 'author' ? 'wishlist.placeholderAuthor' : 'wishlist.placeholderBook'
        )}
        right={loading ? <ActivityIndicator color={colors.primary} /> : undefined}
      />
      <FlatList
        data={suggestions}
        style={{ maxHeight: 320 }}
        keyboardShouldPersistTaps="handled"
        keyExtractor={(s, i) => `${s.workKey || s.authorKey}-${i}`}
        renderItem={({ item }) => (
          <Pressable
            onPress={() => pick(item)}
            android_ripple={{ color: colors.ripple }}
            style={{
              flexDirection: 'row',
              alignItems: 'center',
              gap: spacing.sm,
              paddingVertical: spacing.xs,
              borderRadius: radii.md,
            }}
          >
            {item.kind === 'book' && <BookCover url={item.coverUrl} width={32} height={46} />}
            <View style={{ flex: 1 }}>
              <SText variant="label" numberOfLines={2}>
                {item.title}
              </SText>
              {!!(item.author || item.year) && (
                <SText variant="small" tone="textSecondary" numberOfLines={1}>
                  {[item.author, item.year || ''].filter(Boolean).join(' · ')}
                </SText>
              )}
            </View>
          </Pressable>
        )}
      />
    </SModal>
  );
}
