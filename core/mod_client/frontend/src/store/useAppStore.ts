
import { create } from 'zustand';
import { persist } from 'zustand/middleware';
import { TitleBarTheme } from 'wailsjs/go/main/App';
import { types } from 'wailsjs/go/models';

export interface User {
  publicKey: string;
  alias: string;
  role: 'member' | 'moderator' | 'admin';
  communityRoles?: Record<string, 'member' | 'moderator' | 'admin'>;
  avatarSvg?: string;
}

export interface Community {
  id: string;
  name: string;
  topic: string;
  memberCount: number;
  isJoined: boolean;
  requiresApproval: boolean;
}

export interface Message {
  content: string;
  authorPublicKey: string;
  authorAlias: string;
  timestamp: bigint;
  communityId: string;
  status: 'pending' | 'approved' | 'rejected';
  moderationNote?: types.ModCert[];
  avatarSvg?:string;
  sign:string;
}

export interface ReportedMessage {
  content: string;
  communityId: string;
  sign:string;
  note:string;
  type?: string;
}

export interface ModLogEntry {
  content: string;
  timestamp: number;
  status: string;
}

export interface PendingModeration {
  id: string;
  ts: number;
  content: string;
  reason: string;
  approved: number;
  rejected: number;
  totalMods: number;
  ackCount: number;
  awaitingMods: number;
}

interface AppState {
  // User state
  user: User | null;
  isAuthenticated: boolean;
  isIncognito: boolean;
  
  // Theme state
  isDarkMode: boolean;
  
  // Community state
  communities: Community[];
  currentCommunity: Community | null;
  
  // Messages state
  messages: Message[];
  isLoading: boolean;

  // Pending moderation queue (persists across tab switches)
  pendingQueue: PendingModeration[];

  // Network state
  relayStatus: string;

  // Actions
  setUser: (user: User) => void;
  setIncognito: (value: boolean) => void;
  logout: () => void;
  toggleTheme: () => void;
  setCommunities: (communities: Community[]) => void;
  setCurrentCommunity: (community: Community) => void;
  setMessages: (messages: Message[]) => void;
  addMessage: (message: Message) => void;
  setLoading: (loading: boolean) => void;
  joinCommunity: (communityId: string) => void;
  setPendingQueue: (queue: PendingModeration[]) => void;
  setRelayStatus: (status: string) => void;
}

export const useAppStore = create<AppState>()(
  persist(
    (set, get) => ({
      // Initial state
      user: null,
      isAuthenticated: false,
      isIncognito: false,
      isDarkMode: true,
      communities: [],
      currentCommunity: null,
      messages: [],
      isLoading: false,
      pendingQueue: [],
      relayStatus: 'offline',

      // Actions
      setUser: (user) => set({ user, isAuthenticated: true }),
      setIncognito: (value) => set({ isIncognito: value }),
      
      logout: () => set({ 
        user: null, 
        isAuthenticated: false,
        isIncognito: false,
        currentCommunity: null,
        messages: [] 
      }),
      
      toggleTheme: () => {
        const newTheme = !get().isDarkMode;
        set({ isDarkMode: newTheme });
        document.documentElement.classList.toggle('dark', newTheme);
        TitleBarTheme(newTheme);
      },
      
      setCommunities: (communities) => set({ communities }),
      
      setCurrentCommunity: (community) => set({ 
        currentCommunity: community,
        messages: [] // Clear messages when switching communities
      }),
      
      setMessages: (messages) => set({ messages }),
      
      addMessage: (message) => set((state) => ({ 
        messages: [message, ...state.messages] 
      })),
      
      setLoading: (loading) => set({ isLoading: loading }),
      
      joinCommunity: (communityId) => set((state) => ({
        communities: state.communities.map(c => 
          c.id === communityId ? { ...c, isJoined: true } : c
        )
      })),

      setPendingQueue: (queue) => set({ pendingQueue: queue }),
      setRelayStatus: (status) => set({ relayStatus: status }),
    }),
    {
      name: 'libr-storage',
      partialize: (state) => ({ 
        isDarkMode: state.isDarkMode 
      }),
    }
  )
);
