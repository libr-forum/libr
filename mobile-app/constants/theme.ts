import { Platform } from 'react-native';

const desktopTheme = {
  background: '#0a0f1c',
  text: '#ededed',
  primary: '#020817',
  border: '#1e2b3c',
  tint: '#1fa4a9',
  accent: '#00fcdf',
  muted: '#8b9ab5',
  debugBg: '#1e1e2e',
  green: '#22c55e',
  red: '#ef4444',
  amber: '#f59e0b',
} as const;

const incognitoTheme = {
  background: '#0f0a18',
  text: '#f2eefc',
  primary: '#171025',
  border: '#3b2855',
  tint: '#9b6fd4',
  accent: '#d7b5ff',
  muted: '#a99ac6',
  debugBg: '#1d1530',
  green: desktopTheme.green,
  red: desktopTheme.red,
  amber: desktopTheme.amber,
} as const;

export const Colors = {
  light: {
    text: desktopTheme.text,
    background: desktopTheme.background,
    primary: desktopTheme.primary,
    tint: desktopTheme.tint,
    accent: desktopTheme.accent,
    icon: desktopTheme.muted,
    muted: desktopTheme.muted,
    border: desktopTheme.border,
    debugBg: desktopTheme.debugBg,
    green: desktopTheme.green,
    red: desktopTheme.red,
    amber: desktopTheme.amber,
    tabIconDefault: desktopTheme.muted,
    tabIconSelected: desktopTheme.tint,
  },
  dark: {
    text: desktopTheme.text,
    background: desktopTheme.background,
    primary: desktopTheme.primary,
    tint: desktopTheme.tint,
    accent: desktopTheme.accent,
    icon: desktopTheme.muted,
    muted: desktopTheme.muted,
    border: desktopTheme.border,
    debugBg: desktopTheme.debugBg,
    green: desktopTheme.green,
    red: desktopTheme.red,
    amber: desktopTheme.amber,
    tabIconDefault: desktopTheme.muted,
    tabIconSelected: desktopTheme.tint,
  },
};

export function getAppColors(isIncognito = false) {
  if (!isIncognito) {
    return Colors.dark;
  }

  return {
    ...Colors.dark,
    text: incognitoTheme.text,
    background: incognitoTheme.background,
    primary: incognitoTheme.primary,
    tint: incognitoTheme.tint,
    accent: incognitoTheme.accent,
    icon: incognitoTheme.muted,
    muted: incognitoTheme.muted,
    border: incognitoTheme.border,
    debugBg: incognitoTheme.debugBg,
    tabIconDefault: incognitoTheme.muted,
    tabIconSelected: incognitoTheme.tint,
  };
}

export const Fonts = Platform.select({
  ios: {
    sans: 'Comfortaa-Regular',
    medium: 'Comfortaa-Regular',
    bold: 'Comfortaa-Regular',
    display: 'Comfortaa-Regular',
    rounded: 'Comfortaa-Regular',
    mono: 'ui-monospace',
  },
  default: {
    sans: 'Comfortaa-Regular',
    medium: 'Comfortaa-Regular',
    bold: 'Comfortaa-Regular',
    display: 'Comfortaa-Regular',
    rounded: 'Comfortaa-Regular',
    mono: 'monospace',
  },
  web: {
    sans: "'Comfortaa', system-ui, -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Helvetica, Arial, sans-serif",
    medium: "'Comfortaa', system-ui, -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Helvetica, Arial, sans-serif",
    bold: "'Comfortaa', system-ui, -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Helvetica, Arial, sans-serif",
    display: "'Comfortaa', system-ui, -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Helvetica, Arial, sans-serif",
    rounded: "'Comfortaa', system-ui, -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Helvetica, Arial, sans-serif",
    mono: "SFMono-Regular, Menlo, Monaco, Consolas, 'Liberation Mono', 'Courier New', monospace",
  },
});
