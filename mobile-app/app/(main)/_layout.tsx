import { Stack } from 'expo-router';
import React, { createContext, useContext, useState } from 'react';
import { View, StyleSheet } from 'react-native';
import { Sidebar } from '@/components/Sidebar';
import { getAppColors } from '@/constants/theme';
import { useAppStore } from '@/store/useAppStore';

interface SidebarContextType {
  openSidebar: () => void;
  closeSidebar: () => void;
  toggleSidebar: () => void;
  isOpen: boolean;
}

const SidebarContext = createContext<SidebarContextType>({
  openSidebar: () => { },
  closeSidebar: () => { },
  toggleSidebar: () => { },
  isOpen: false,
});

export const useSidebar = () => useContext(SidebarContext);

export default function SidebarLayout() {
  const [isOpen, setIsOpen] = useState(false);
  const { state } = useAppStore();
  const colors = getAppColors(state.isIncognito);

  const openSidebar = () => setIsOpen(true);
  const closeSidebar = () => setIsOpen(false);
  const toggleSidebar = () => setIsOpen(p => !p);

  return (
    <SidebarContext.Provider value={{ openSidebar, closeSidebar, toggleSidebar, isOpen }}>
      <View style={[styles.container, { backgroundColor: colors.background }]}>
        <Stack
          screenOptions={{
            headerShown: false,
            contentStyle: { backgroundColor: colors.background },
          }}
        >
          <Stack.Screen name="index" />
          <Stack.Screen name="moderation" />
          <Stack.Screen name="queue" />
          <Stack.Screen name="profile" />
        </Stack>
        <Sidebar isOpen={isOpen} onClose={closeSidebar} onOpen={openSidebar} />
      </View>
    </SidebarContext.Provider>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: 'transparent',
  },
});
