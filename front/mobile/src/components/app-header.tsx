import React from 'react';
import { Image, View } from 'react-native';
import { colors } from '../theme/colors';
import SText from './shared/SText';
import logo from '../../assets/icon.png';

export default function AppHeader() {
  return (
    <View
      style={{
        alignItems: 'center',
        justifyContent: 'center',
        flexDirection: 'row',
        gap: '8',
        marginBottom: 24,
      }}
    >
      <View style={{}}>
        <Image source={logo} style={{ width: 38, height: 38 }} resizeMode="contain" />
      </View>
      <SText
        style={{
          fontSize: 32,
          color: colors.primary,
          fontFamily: 'bold',
        }}
      >
        Inkomi
      </SText>
    </View>
  );
}
