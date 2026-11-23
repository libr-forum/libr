import { SafeAreaProvider } from 'react-native-safe-area-context';
import { DarkTheme, ThemeProvider } from '@react-navigation/native';
import { Stack } from 'expo-router';
import { StatusBar } from 'expo-status-bar';
import { useFonts } from 'expo-font';
import 'react-native-reanimated';
import React from 'react';
import { useKeepAwake } from 'expo-keep-awake';
import { Text, TextInput, ScrollView } from 'react-native';

import { Fonts, getAppColors } from '@/constants/theme';
import { AppStoreProvider, useAppStore } from '@/store/useAppStore';

export const unstable_settings = {
  anchor: '(main)',
};

class ErrorBoundary extends React.Component<
  { children: React.ReactNode },
  { error: Error | null }
> {
  constructor(props: { children: React.ReactNode }) {
    super(props);
    this.state = { error: null };
  }
  static getDerivedStateFromError(error: Error) {
    return { error };
  }
  render() {
    if (this.state.error) {
      const colors = getAppColors(false);
      return (
        <ScrollView style={{ flex: 1, padding: 24, paddingTop: 60, backgroundColor: colors.background }}>
          <Text style={{ color: '#ef4444', fontSize: 18, fontWeight: 'bold', marginBottom: 12 }}>
            App Error
          </Text>
          <Text style={{ color: colors.text, fontSize: 13, fontFamily: Fonts.mono }}>
            {this.state.error.message}\n\n{this.state.error.stack}
          </Text>
        </ScrollView>
      );
    }
    return this.props.children;
  }
}

function AppShell() {
  const { state } = useAppStore();
  const colors = getAppColors(state.isIncognito);

  const theme = {
    ...DarkTheme,
    colors: {
      ...DarkTheme.colors,
      background: colors.background,
      card: colors.primary,
      text: colors.text,
      border: colors.border,
      primary: colors.tint,
    },
  };

  return (
    <ThemeProvider value={theme}>
      <Stack>
        <Stack.Screen name="(main)" options={{ headerShown: false }} />
        <Stack.Screen name="modal" options={{ presentation: 'modal', title: 'Modal' }} />
      </Stack>
      <StatusBar style="light" />
    </ThemeProvider>
  );
}

export default function RootLayout() {
  useKeepAwake();
  const [fontsLoaded] = useFonts({
    'Comfortaa-Regular': require('../assets/fonts/Comfortaa-Regular.ttf'),
    'Comfortaa-SemiBold': require('../assets/fonts/Comfortaa-SemiBold.ttf'),
    'Comfortaa-Bold': require('../assets/fonts/Comfortaa-Bold.ttf'),
  });

  (Text as any).defaultProps = (Text as any).defaultProps ?? {};
  (Text as any).defaultProps.style = [{ fontFamily: Fonts.sans }];
  (TextInput as any).defaultProps = (TextInput as any).defaultProps ?? {};
  (TextInput as any).defaultProps.style = [{ fontFamily: Fonts.sans }];

  if (!fontsLoaded) {
    return null;
  }

  return (
    <SafeAreaProvider>
      <ErrorBoundary>
        <AppStoreProvider>
          <AppShell />
        </AppStoreProvider>
      </ErrorBoundary>
    </SafeAreaProvider>
  );
}
