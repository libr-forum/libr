import React, { useState,useEffect } from 'react';
import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import { motion, AnimatePresence } from 'framer-motion';
import { Toaster } from "@/components/ui/toaster";
import { Toaster as Sonner } from "@/components/ui/sonner";
import { TooltipProvider } from "@/components/ui/tooltip";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { toast } from 'sonner';
import logotransparent from "./components/assets/logo_transparent_noname-01.png";

import { useAppStore } from './store/useAppStore';
import { apiService } from './services/api';
import { Sidebar } from './components/layout/Sidebar';
import { ChatRoom } from './pages/ChatRoom';
import { ModLogs } from './pages/ModLogs';
import { ModConfig } from './pages/ModConfig';
import { Communities } from './pages/Communities';
import { logger } from './logger/logger'

import { Connect,GetRelayStatus,FetchPubKey,TitleBarTheme, GetRelayAddr, IsIncognitoEnabled } from '../wailsjs/go/main/App';
import { EventsOn,Quit } from 'wailsjs/runtime/runtime';
import {
  AlertDialog,
  AlertDialogTrigger,
  AlertDialogContent,
  AlertDialogHeader,
  AlertDialogTitle,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogCancel,
  AlertDialogAction
} from "@/components/ui/alert-dialog";
import { MsgReports } from './pages/MessageReports';

const queryClient = new QueryClient();

interface RelayErrorDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onRetry: () => void;
}

export const RelayErrorDialog: React.FC<RelayErrorDialogProps> = ({ open, onOpenChange, onRetry }) => {
  return (
    <AlertDialog open={open} onOpenChange={onOpenChange}>
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>Relay Connection Failed</AlertDialogTitle>
          <AlertDialogDescription>
            We couldn’t connect to the relay after multiple attempts. Please try again later.
          </AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter>
          <AlertDialogCancel onClick={() => Quit()}>Close App</AlertDialogCancel>
          <AlertDialogAction onClick={() => {
            onOpenChange(false);
            onRetry();
          }}>Try Again</AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  );
}

export function useRelayConnection() {
  const [relayFailed, setRelayFailed] = useState(false);

  useEffect(() => {
    const tryConnect = async () => {
      logger.info("🛜 Starting relay connection...");

      const relayAddrs = await fetchRelayAddrs();
      let connected = false;

      for (let i = 0; i < 10; i++) {
        const status = await GetRelayStatus();
        if (status === 'online') {
          useAppStore.getState().setRelayStatus('online');
          connected = true;
          break;
        }

        const error = await Connect(relayAddrs);
        if (error != null) {
          const recheck = await GetRelayStatus();
          if (recheck === 'online') {
            useAppStore.getState().setRelayStatus('online');
            connected = true;
            break;
          }
        }

        await new Promise(res => setTimeout(res, 1000));
      }

        if (!connected) {
          logger.error("Could not connect to relay after multiple attempts.");
          setRelayFailed(true);
        }
    };

    tryConnect();
  }, []);

  return { relayFailed, setRelayFailed };
}

async function fetchRelayAddrs(): Promise<string[]> {
  try {
    const addrs = await GetRelayAddr();
    // Only keep valid multiaddrs
    const validAddrs = addrs.filter(addr => addr.startsWith('/'));
    logger.debug('[RelayAddrs] Valid relay addresses:', validAddrs);
    return validAddrs;
  } catch (error) {
    logger.error("Error loading relay addresses from MongoDB:", error);
    return [];
  }
}


// const Layout: React.FC<{ children: React.ReactNode }> = ({ children }) => {
//   return (
//     <div className="flex h-screen bg-libr-primary relative">
//       <Sidebar />
//       <main className="flex-1 overflow-hidden transition-all duration-300">
//         {children}
//       </main>
//     </div>
//   );
// };

const App: React.FC = () => {
  const {
    isAuthenticated,
    setUser,
    setIncognito,
    setCommunities,
    isDarkMode,
    isIncognito,
    setCurrentCommunity,
    communities,
    setRelayStatus
  } = useAppStore();

  const [relayFailed, setRelayFailed] = useState(false);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    document.documentElement.classList.toggle("dark", isDarkMode);
  }, [isDarkMode]);

  useEffect(() => {
    document.documentElement.classList.toggle("incognito-theme", isIncognito);
  }, [isIncognito]);
  
  useEffect(() => {
    TitleBarTheme(isDarkMode);
  }, []);

  useEffect(() => {
    EventsOn("navigate-to-root", () => {
      window.location.href = "/";
    });

    EventsOn("relay_status_changed", (status: string) => {
      logger.info(`🌐 Relay status changed: ${status}`);
      setRelayStatus(status);
      if (status === 'offline') {
        toast.error("Relay Disconnected", {
          description: "Attempting to reconnect...",
        });
      } else if (status === 'online') {
        toast.success("Relay Reconnected", {
          description: "Secure connection established.",
        });
      }
    });

    // Initial status sync
    GetRelayStatus().then(setRelayStatus);
  }, [setRelayStatus]);

  const initializeApp = React.useCallback(async () => {
    setLoading(true);
    setRelayFailed(false);
    try {
      logger.debug("🔄 Fetching relay addresses...");
      const relayAddrs = await fetchRelayAddrs();
      
      let status = await GetRelayStatus();
      let connected = false;
      if (status === "online") {
        setRelayStatus("online");
        connected = true;
      } else {
        for (let i = 0; i < 10; i++) {
          const error = await Connect(relayAddrs);
          const recheck = await GetRelayStatus();
          if (recheck === "online") {
            setRelayStatus("online");
            connected = true;
            break;
          }

          await new Promise(res => setTimeout(res, 1000));
        }
      }

      if (!connected) {
        logger.error("❌ Could not connect to relay.");
        setRelayFailed(true);
        return;
      }

      logger.info("✅ Relay connected. Authenticating...");
      
      const publicKey = await FetchPubKey();
      const user = await apiService.authenticate(publicKey);
      setUser(user);
      setIncognito(await IsIncognitoEnabled());

      const fetchedCommunities = await apiService.getCommunities();
      setCommunities(fetchedCommunities);

      const firstJoined = fetchedCommunities.find(c => c.isJoined);
      if (firstJoined) {
        setCurrentCommunity(firstJoined);
      }
    } catch (err) {
      logger.error("🔥 App initialization failed:", err);
      setRelayFailed(true);
    } finally {
      setLoading(false);
    }
  }, [setUser, setIncognito, setCommunities, setCurrentCommunity]);

  useEffect(() => {
    initializeApp();
  }, [initializeApp]);

  if (loading) {
    return (
      <>
        <div className="min-h-screen bg-libr-primary flex items-center justify-center">
          <motion.div
            initial={{ opacity: 0, scale: 0.9 }}
            animate={{ opacity: 1, scale: 1 }}
            className="text-center"
          >
            <div className="w-40 h-40 rounded-2xl flex items-center justify-center mx-auto mb-2 libr-glow bg-cover bg-center" style={{ backgroundImage: `url(${logotransparent})` }}>
              <motion.div
                animate={{ rotate: 360 }}
                transition={{ duration: 2, repeat: Infinity, ease: "linear" }}
                className="w-10 h-10 border-3 border-white border-t-transparent rounded-full"
              />
            </div>
            <h1 className="text-3xl font-bold text-foreground mb-2">
              Connecting to libr
            </h1>
            <p className="text-muted-foreground">
              Establishing secure connection...
            </p>
          </motion.div>
        </div>
        <RelayErrorDialog open={relayFailed} onOpenChange={setRelayFailed} onRetry={initializeApp} />
      </>
    );
  }

  return (
    <>
      <QueryClientProvider client={queryClient}>
        <TooltipProvider>
          <Toaster />
          <Sonner />
          <BrowserRouter>
            {/* <Layout> */}
              <AnimatePresence mode="wait">
                <Routes>
                  <Route path="/" element={<Navigate to="/chat" replace />} />
                  <Route path="/chat" element={<ChatRoom />} />
                  <Route path="/communities" element={<Communities />} />
                  <Route path="/modlogs" element={<ModLogs />} />
                  <Route path="/modconfig" element={<ModConfig />} />
                  <Route path="/msgreports" element={<MsgReports />} />
                  <Route path="*" element={<Navigate to="/chat" replace />} />
                </Routes>
              </AnimatePresence>
            {/* </Layout> */}
          </BrowserRouter>
        </TooltipProvider>
      </QueryClientProvider>
      <RelayErrorDialog open={relayFailed} onOpenChange={setRelayFailed} onRetry={initializeApp} />
    </>
  );
};


export default App;
