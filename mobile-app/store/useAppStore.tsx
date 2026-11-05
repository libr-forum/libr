import React, { createContext, useContext, useReducer, useCallback, ReactNode, useEffect } from 'react';
import { AppState as ReactNativeAppState, AppStateStatus, Platform, PermissionsAndroid } from 'react-native';
import AsyncStorage from '@react-native-async-storage/async-storage';
import * as BackgroundFetch from 'expo-background-fetch';
import * as TaskManager from 'expo-task-manager';
import LibrCore, { RetMsgCert } from '@/modules/LibrCore';

// ── Types ─────────────────────────────────────────────────────────────────────

export type ConnectionStatus = 'idle' | 'initializing' | 'connected' | 'error';

interface AppState {
  /** Ed25519 public key (base64) of this device. */
  publicKey: string;
  /** Whether the active identity is a temporary incognito one. */
  isIncognito: boolean;
  /** libp2p Peer ID for this device. */
  peerId: string;
  /** Current network connection/init status. */
  connectionStatus: ConnectionStatus;
  /** Last error message, if any. */
  lastError: string | null;
  /** All fetched messages (most-recent first). */
  messages: RetMsgCert[];
  /** Whether a message fetch is in progress. */
  isFetching: boolean;
  /** Set of message signatures reported by the current user. */
  reportedSigns: Set<string>;
  /** Whether the current user is a moderator. */
  isModerator: boolean;
  /** Global custom alert / notification modal config. */
  noticeModal: NoticeModalConfig | null;
}

export type NoticeModalConfig = {
  visible: boolean;
  title: string;
  message: string;
  icon?: 'check' | 'clock' | 'info';
};

type Action =
  | { type: 'SET_PUBLIC_KEY'; payload: string }
  | { type: 'SET_INCOGNITO'; payload: boolean }
  | { type: 'SET_PEER_ID'; payload: string }
  | { type: 'SET_CONNECTION_STATUS'; payload: ConnectionStatus }
  | { type: 'SET_ERROR'; payload: string | null }
  | { type: 'SET_MESSAGES'; payload: RetMsgCert[] }
  | { type: 'SET_FETCHING'; payload: boolean }
  | { type: 'REMOVE_MESSAGE'; payload: string }
  | { type: 'ADD_MESSAGE'; payload: RetMsgCert }
  | { type: 'SET_MODERATOR'; payload: boolean }
  | { type: 'ADD_REPORTED_SIGN'; payload: string }
  | { type: 'REMOVE_REPORTED_SIGN'; payload: string }
  | { type: 'SET_REPORTED_SIGNS'; payload: Set<string> }
  | { type: 'SET_NOTICE_MODAL'; payload: NoticeModalConfig | null };

// ── Reducer ───────────────────────────────────────────────────────────────────

const initialState: AppState = {
  publicKey: '',
  isIncognito: false,
  peerId: '',
  connectionStatus: 'idle',
  lastError: null,
  messages: [],
  isFetching: false,
  reportedSigns: new Set(),
  isModerator: false,
  noticeModal: null,
};

function reducer(state: AppState, action: Action): AppState {
  switch (action.type) {
    case 'SET_PUBLIC_KEY': return { ...state, publicKey: action.payload };
    case 'SET_INCOGNITO': return { ...state, isIncognito: action.payload };
    case 'SET_PEER_ID': return { ...state, peerId: action.payload };
    case 'SET_CONNECTION_STATUS': return { ...state, connectionStatus: action.payload };
    case 'SET_ERROR': return { ...state, lastError: action.payload };
    case 'SET_MESSAGES': {
      // Keep optimistic messages (temp-prefix) that haven't been "confirmed" yet
      const optimistic = state.messages.filter(m => m.sign.startsWith('temp-'));
      const incoming = action.payload;

      // Rough matching to remove optimistic once real one arrives: match by content + sender
      const remainingOptimistic = optimistic.filter(opt =>
        !incoming.some(inc => inc.public_key === opt.public_key && inc.msg.content === opt.msg.content)
      );

      // Merge and sort by timestamp descending
      const merged = [...remainingOptimistic, ...incoming].sort((a, b) => b.msg.ts - a.msg.ts);
      return { ...state, messages: merged };
    }
    case 'SET_FETCHING': return { ...state, isFetching: action.payload };
    case 'REMOVE_MESSAGE': return { ...state, messages: state.messages.filter(m => m.sign !== action.payload) };
    case 'ADD_MESSAGE': {
      if (state.messages.find(m => m.sign === action.payload.sign)) return state;
      // Add to beginning and sort just in case
      const newMessages = [action.payload, ...state.messages].sort((a, b) => b.msg.ts - a.msg.ts);
      return { ...state, messages: newMessages };
    }
    case 'ADD_REPORTED_SIGN': {
      const newSet = new Set(state.reportedSigns);
      newSet.add(action.payload);
      return { ...state, reportedSigns: newSet };
    }
    case 'REMOVE_REPORTED_SIGN': {
      const newSet = new Set(state.reportedSigns);
      newSet.delete(action.payload);
      return { ...state, reportedSigns: newSet };
    }
    case 'SET_REPORTED_SIGNS': return { ...state, reportedSigns: action.payload };
    case 'SET_MODERATOR': return { ...state, isModerator: action.payload };
    case 'SET_NOTICE_MODAL': return { ...state, noticeModal: action.payload };
    default: return state;
  }
}

// ── Context ───────────────────────────────────────────────────────────────────

interface StoreContextValue {
  state: AppState;
  setPublicKey: (key: string) => void;
  setIncognito: (value: boolean) => void;
  setPeerId: (id: string) => void;
  setConnectionStatus: (status: ConnectionStatus) => void;
  setError: (err: string | null) => void;
  setMessages: (msgs: RetMsgCert[]) => void;
  setFetching: (v: boolean) => void;
  removeMessage: (sign: string) => void;
  addMessage: (msg: RetMsgCert) => void;
  setModerator: (isMod: boolean) => void;
  addReportedSign: (sign: string, cert?: RetMsgCert) => void;
  removeReportedSign: (sign: string) => void;
  showNotice: (title: string, message: string, icon?: 'check' | 'clock' | 'info') => void;
  hideNotice: () => void;
}

const StoreContext = createContext<StoreContextValue | undefined>(undefined);
const REPORTED_SIGNS_KEY = '@libr_reported_signs';
const REPORTED_MESSAGES_KEY = '@libr_reported_messages';
const NOTIFIED_SIGNS_KEY = '@libr_notified_signs';
const DISMISSED_SIGNS_KEY = '@libr_dismissed_signs';

/**
 * Checks for newly finalized moderation results and displays an OS notification
 * for any items that have not yet been notified.
 */
async function checkAndSendModerationNotifications(): Promise<boolean> {
  // Go core's cronjob now triggers native system notifications directly via NotificationHandler
  // immediately upon reaching quorum, both in the background and foreground.
  return false;
}

// ── Background Task Definition ────────────────────────────────────────────────
const BACKGROUND_MODERATION_TASK = 'BACKGROUND_MODERATION_TASK';

TaskManager.defineTask(BACKGROUND_MODERATION_TASK, async () => {
  try {
    const tickResult = await LibrCore.tickCron();
    console.log('[BackgroundFetch] Moderation tick result:', tickResult);

    const hasNewNotification = await checkAndSendModerationNotifications();

    return hasNewNotification || tickResult === 'ok'
      ? BackgroundFetch.BackgroundFetchResult.NewData
      : BackgroundFetch.BackgroundFetchResult.NoData;
  } catch (error) {
    console.warn('[BackgroundFetch] Moderation tick failed:', error);
    return BackgroundFetch.BackgroundFetchResult.Failed;
  }
});

export function AppStoreProvider({ children }: { children: ReactNode }) {
  const [state, dispatch] = useReducer(reducer, initialState);

  // Request notification permissions on Android 13+
  useEffect(() => {
    if (Platform.OS === 'android' && Platform.Version >= 33) {
      PermissionsAndroid.request(PermissionsAndroid.PERMISSIONS.POST_NOTIFICATIONS)
        .catch(err => console.warn('[AppStore] Notification permission request error:', err));
    }
  }, []);

  // Periodic foreground check for completed moderations to show status bar alerts
  useEffect(() => {
    // Initial check on mount
    checkAndSendModerationNotifications();

    const interval = setInterval(() => {
      if (ReactNativeAppState.currentState === 'active') {
        checkAndSendModerationNotifications();
      }
    }, 10000); // Check every 10 seconds while active

    return () => clearInterval(interval);
  }, []);

  // Register Background Task
  useEffect(() => {
    (async () => {
      try {
        const isRegistered = await TaskManager.isTaskRegisteredAsync(BACKGROUND_MODERATION_TASK);
        if (!isRegistered) {
          await BackgroundFetch.registerTaskAsync(BACKGROUND_MODERATION_TASK, {
            minimumInterval: 15 * 60, // 15 minutes
            stopOnTerminate: false,
            startOnBoot: true,
          });
          console.log('[BackgroundFetch] Task registered');
        }
      } catch (err) {
        console.warn('[BackgroundFetch] Registration failed:', err);
      }
    })();
  }, []);

  // Load reported signs on mount
  useEffect(() => {
    AsyncStorage.getItem(REPORTED_SIGNS_KEY).then((data: string | null) => {
      if (data) {
        try {
          const arr = JSON.parse(data);
          if (Array.isArray(arr)) {
            dispatch({ type: 'SET_REPORTED_SIGNS', payload: new Set(arr) });
          }
        } catch { /* ignore parse error */ }
      }
    });
  }, []);

  // Manage Go Cron Job Lifecycle
  useEffect(() => {
    const subscription = ReactNativeAppState.addEventListener('change', nextAppState => {
      if (nextAppState === 'active' && state.connectionStatus === 'connected') {
        LibrCore.startCron().catch(console.warn);
      } else if (nextAppState === 'background') {
        // We no longer explicitly stop the cron, but it likely will be suspended by OS.
        // The BackgroundFetch task will wake it up for a "tick" periodically.
        console.log('[LibrCore] App backgrounded, letting periodic fetch handle cron.');
      }
    });

    // Start immediately if already connected
    if (ReactNativeAppState.currentState === 'active' && state.connectionStatus === 'connected') {
      LibrCore.startCron().catch(console.warn);
    }

    return () => {
      subscription.remove();
      LibrCore.stopCron().catch(console.warn);
    };
  }, [state.connectionStatus]);

  const setPublicKey = useCallback((key: string) => dispatch({ type: 'SET_PUBLIC_KEY', payload: key }), []);
  const setIncognito = useCallback((value: boolean) => dispatch({ type: 'SET_INCOGNITO', payload: value }), []);
  const setPeerId = useCallback((id: string) => dispatch({ type: 'SET_PEER_ID', payload: id }), []);
  const setConnectionStatus = useCallback((s: ConnectionStatus) => dispatch({ type: 'SET_CONNECTION_STATUS', payload: s }), []);
  const setError = useCallback((e: string | null) => dispatch({ type: 'SET_ERROR', payload: e }), []);
  const setMessages = useCallback((m: RetMsgCert[]) => dispatch({ type: 'SET_MESSAGES', payload: m }), []);
  const setFetching = useCallback((v: boolean) => dispatch({ type: 'SET_FETCHING', payload: v }), []);
  const removeMessage = useCallback((sign: string) => dispatch({ type: 'REMOVE_MESSAGE', payload: sign }), []);
  const addMessage = useCallback((msg: RetMsgCert) => dispatch({ type: 'ADD_MESSAGE', payload: msg }), []);
  const setModerator = useCallback((isMod: boolean) => dispatch({ type: 'SET_MODERATOR', payload: isMod }), []);
  const addReportedSign = useCallback((sign: string, cert?: RetMsgCert) => {
    dispatch({ type: 'ADD_REPORTED_SIGN', payload: sign });
    // Persist sign to AsyncStorage
    AsyncStorage.getItem(REPORTED_SIGNS_KEY).then((data: string | null) => {
      const arr = data ? JSON.parse(data) : [];
      if (!arr.includes(sign)) {
        arr.push(sign);
        AsyncStorage.setItem(REPORTED_SIGNS_KEY, JSON.stringify(arr));
      }
    }).catch(() => { });
    // Also persist the full message cert for offline display
    if (cert) {
      AsyncStorage.getItem(REPORTED_MESSAGES_KEY).then((data: string | null) => {
        const cache: Record<string, RetMsgCert> = data ? JSON.parse(data) : {};
        cache[sign] = cert;
        AsyncStorage.setItem(REPORTED_MESSAGES_KEY, JSON.stringify(cache));
      }).catch(() => { });
    }
  }, []);

  const removeReportedSign = useCallback((sign: string) => {
    dispatch({ type: 'REMOVE_REPORTED_SIGN', payload: sign });
    AsyncStorage.getItem(REPORTED_SIGNS_KEY).then((data: string | null) => {
      if (data) {
        const arr: string[] = JSON.parse(data).filter((s: string) => s !== sign);
        AsyncStorage.setItem(REPORTED_SIGNS_KEY, JSON.stringify(arr));
      }
    }).catch(() => { });
    AsyncStorage.getItem(REPORTED_MESSAGES_KEY).then((data: string | null) => {
      if (data) {
        const cache: Record<string, RetMsgCert> = JSON.parse(data);
        delete cache[sign];
        AsyncStorage.setItem(REPORTED_MESSAGES_KEY, JSON.stringify(cache));
      }
    }).catch(() => { });
  }, []);

  const showNotice = useCallback((title: string, message: string, icon: 'check' | 'clock' | 'info' = 'check') => {
    dispatch({ type: 'SET_NOTICE_MODAL', payload: { visible: true, title, message, icon } });
  }, []);

  const hideNotice = useCallback(() => {
    dispatch({ type: 'SET_NOTICE_MODAL', payload: null });
  }, []);

  return (
    <StoreContext.Provider value={{ state, setPublicKey, setIncognito, setPeerId, setConnectionStatus, setError, setMessages, setFetching, removeMessage, addMessage, setModerator, addReportedSign, removeReportedSign, showNotice, hideNotice }}>
      {children}
    </StoreContext.Provider>
  );
}

export function useAppStore() {
  const ctx = useContext(StoreContext);
  if (!ctx) throw new Error('useAppStore must be used inside <AppStoreProvider>');
  return ctx;
}
