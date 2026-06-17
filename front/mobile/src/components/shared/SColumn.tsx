import React, { Children, ReactNode } from 'react';
import { View } from 'react-native';
import { colors } from '../../theme/colors';
import SDivider from './SDivider';

interface Props {
  children?: ReactNode;
  footer?: ReactNode;
  removeCellPadding?: boolean;
}

export default function SColumn<T>({ children, footer, removeCellPadding }: Props) {
  const childrens = Children.toArray(children);

  return (
    <View style={{ backgroundColor: colors.card, borderRadius: 14, overflow: 'hidden' }}>
      {childrens.map((e, i) => (
        <View key={i}>
          {i !== 0 && <SDivider />}
          <Wrapper removeCellPadding={removeCellPadding}>{e}</Wrapper>
        </View>
      ))}
      {footer && childrens.length !== 0 && <SDivider />}
      {footer && <Wrapper>{footer}</Wrapper>}
    </View>
  );
}

interface WrapperProps {
  children?: ReactNode;
  removeCellPadding?: boolean;
}

function Wrapper({ children, removeCellPadding }: WrapperProps) {
  return (
    <View
      style={
        !removeCellPadding && {
          paddingHorizontal: 16,
          paddingVertical: 14,
        }
      }
    >
      {children}
    </View>
  );
}
