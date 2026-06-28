import { View, ActivityIndicator } from 'react-native';
import * as WebBrowser from 'expo-web-browser';
import { useEffect } from 'react';
import { router } from 'expo-router';

WebBrowser.maybeCompleteAuthSession();

export default function OAuthCallbackScreen() {
  useEffect(() => {
    if (router.canGoBack()) {
      setTimeout(() => router.back(), 100);
    } else {
      router.replace('/');
    }
  }, []);

  return (
    <View style={{ flex: 1, justifyContent: 'center', alignItems: 'center' }}>
      <ActivityIndicator />
    </View>
  );
}
