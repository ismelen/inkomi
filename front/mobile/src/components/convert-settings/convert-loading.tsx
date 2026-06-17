import React from 'react';
import { ActivityIndicator, View } from 'react-native';
import { colors } from '../../theme/colors';
import CloudUploadIcon from '../icons/cloud-upload-icon';
import SText from '../shared/SText';

export default function ConvertLoading() {
  return (
    <View style={{ flex: 1, justifyContent: 'center', alignItems: 'center' }}>
      <View
        style={{
          justifyContent: 'center',
          alignItems: 'center',
          overflow: 'visible',
          marginBottom: 80,
        }}
      >
        <ActivityIndicator
          size={120}
          color={colors.primary}
          style={{ position: 'absolute', marginTop: 10 }}
        />
        <View
          style={{
            backgroundColor: colors.background,
            width: 95,
            height: 95,
            position: 'absolute',
            marginTop: 10,
            borderRadius: 200,
          }}
        />
        <CloudUploadIcon size="60px" color={colors.primary} />
      </View>
      <SText style={{ opacity: 0.6, fontWeight: 600, fontSize: 22 }}>Initializing conversion</SText>
      <SText style={{ opacity: 0.6, fontSize: 16, color: colors.primary, marginTop: 10 }}>
        Preparing your files
      </SText>
    </View>
  );
}
