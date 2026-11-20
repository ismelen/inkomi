import React, { useEffect, useRef, useState } from 'react';
import { ActivityIndicator, FlatList, Pressable, View } from 'react-native';
import { useTranslation } from 'react-i18next';
import { wishlistApi } from '../../api/wishlist-api';
import { AddWishRequest, WishKind, WishSuggestion } from '../../api/types/wishlist';
import { colors, radii, spacing } from '../../theme';
import { BookCover } from '../books/book-cover';
import { SModal } from '../ui/s-modal';
import { SText } from '../ui/s-text';
import { STextInput } from '../ui/s-text-input';

const debounceMs = 400;
const minQuery = 2;

interface Props {
  kind: WishKind | null;
  onClose: () => void;
  onAdd: (item: AddWishRequest) => void;
}

/** Adds a wish: suggestions from the catalog while typing, or the typed text as is. */
export function AddWishModal({ kind, onClose, onAdd }: Props) {
  const { t } = useTranslation();
  const [query, setQuery] = useState('');
  const [suggestions, setSuggestions] = useState<WishSuggestion[]>([]);
  const [loading, setLoading] = useState(false);
  const timer = useRef<ReturnType<typeof setTimeout> | null>(null);
  const controller = useRef<AbortController | null>(null);

  useEffect(() => {
    setQuery('');
    setSuggestions([]);
  }, [kind]);

  useEffect(
    () => () => {
      if (timer.current) clearTimeout(timer.current);
      controller.current?.abort();
    },
    []
  );

  const onChange = (text: string) => {
    setQuery(text);
    if (timer.current) clearTimeout(timer.current);
    controller.current?.abort();
    if (!kind || text.trim().length < minQuery) {
      setSuggestions([]);
      setLoading(false);
      return;
    }
    timer.current = setTimeout(() => {
      const ctrl = new AbortController();
      controller.current = ctrl;
      setLoading(true);
      wishlistApi
        .lookup(kind, text.trim(), ctrl.signal)
        .then((found) => !ctrl.signal.aborted && setSuggestions(found))
        .catch(() => !ctrl.signal.aborted && setSuggestions([]))
        .finally(() => !ctrl.signal.aborted && setLoading(false));
    }, debounceMs);
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
      visible={kind !== null}
      title={t(kind === 'author' ? 'wishlist.addAuthor' : 'wishlist.addBook')}
      onClose={onClose}
      actions={[
        { label: t('common.close'), onPress: onClose, variant: 'secondary' },
        {
          label: t('wishlist.add'),
          onPress: () => kind && onAdd({ kind, title: text }),
          disabled: text.length === 0,
        },
      ]}
    >
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
