
import { create } from 'zustand';
import { persist } from 'zustand/middleware';
import { TitleBarTheme } from 'wailsjs/go/main/App';

export interface User {
  id: string;
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
  id: string;
  content: string;
  authorId: string;
  authorAlias: string;
  timestamp: Date;
  communityId: string;
  status: 'pending' | 'approved' | 'rejected';
  moderationNote?: string;
  avatarSvg?:string;
}

export interface ModLogEntry {
  content: string;
  timestamp: number;
  status: string;
}


interface AppState {
  // User state
  user: User | null;
  isAuthenticated: boolean;
  
  // Theme state
  isDarkMode: boolean;
  
  // Community state
  communities: Community[];
  currentCommunity: Community | null;
  
  // Messages state
  messages: Message[];
  isLoading: boolean;
  
  // Actions
  setUser: (user: User) => void;
  logout: () => void;
  toggleTheme: () => void;
  setCommunities: (communities: Community[]) => void;
  setCurrentCommunity: (community: Community) => void;
  setMessages: (messages: Message[]) => void;
  addMessage: (message: Message) => void;
  setLoading: (loading: boolean) => void;
  joinCommunity: (communityId: string) => void;
}

export const useAppStore = create<AppState>()(
  persist(
    (set, get) => ({
      // Initial state
      user: null,
      isAuthenticated: false,
      isDarkMode: true,
      communities: [],
      currentCommunity: null,
      messages: [],
      isLoading: false,

      // Actions
      setUser: (user) => set({ user, isAuthenticated: true }),
      
      logout: () => set({ 
        user: null, 
        isAuthenticated: false,
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
    }),
    {
      name: 'libr-storage',
      partialize: (state) => ({ 
        isDarkMode: state.isDarkMode 
      }),
    }
  )
);
