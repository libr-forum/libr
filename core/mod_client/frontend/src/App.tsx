import React, { useState,useEffect } from 'react';
import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import { motion, AnimatePresence } from 'framer-motion';
import { Toaster } from "@/components/ui/toaster";
import { Toaster as Sonner } from "@/components/ui/sonner";
import { TooltipProvider } from "@/components/ui/tooltip";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import logotransparent from "./components/assets/logo_transparent_noname-01.png";

import { useAppStore } from './store/useAppStore';
import { apiService } from './services/api';
import { Sidebar } from './components/layout/Sidebar';
import { ChatRoom } from './pages/ChatRoom';
import { ModLogs } from './pages/ModLogs';
import { ModConfig } from './pages/ModConfig';
import { Communities } from './pages/Communities';

import { Connect,GetRelayStatus,FetchPubKey,TitleBarTheme } from '../wailsjs/go/main/App';
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

const baseURL = 'https://docs.google.com/spreadsheets/d/e/2PACX-1vRDDE0x6LttdW13zLUwodMcVBsqk8fpnUsv-5SIJifZKWRehFpSKuJZawhswGMHSI2fZJDuENQ8SX1v/pub?output=csv';

interface RelayErrorDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

export const RelayErrorDialog: React.FC<RelayErrorDialogProps> = ({ open, onOpenChange }) => {
  const closeApp = () => {
    window.close(); // Wails window close
  };
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
          {/* <AlertDialogCancel onClick={() => onOpenChange(false)}>Cancel</AlertDialogCancel> */}
          <AlertDialogAction onClick={() => Quit()}>Close App</AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  );
}

export function useRelayConnection() {
  const [relayFailed, setRelayFailed] = useState(false);

  useEffect(() => {
    const tryConnect = async () => {
      console.log("🛜 Starting relay connection...");

      const relayAddrs = await fetchRelayAddrs();
      let connected = false;

      for (let i = 0; i < 10; i++) {
        const status = await GetRelayStatus();
        if (status === 'online') {
          connected = true;
          break;
        }

        const error = await Connect(relayAddrs);
        if (error != null) {
          const recheck = await GetRelayStatus();
          if (recheck === 'online') {
            connected = true;
            break;
          }
        }

        await new Promise(res => setTimeout(res, 1000));
      }

      if (!connected) {
        console.error("❌ Failed to connect to relay after 10 attempts.");
        setRelayFailed(true);
      }
    };

    tryConnect();
  }, []);

  return { relayFailed, setRelayFailed };
}

async function fetchRelayAddrs(): Promise<string[]> {
  try {
    const url = "https://docs.google.com/spreadsheets/d/e/2PACX-1vRDDE0x6LttdW13zLUwodMcVBsqk8fpnUsv-5SIJifZKWRehFpSKuJZawhswGMHSI2fZJDuENQ8SX1v/pub?output=csv&gid=1789680527";
    console.log("▶ fetching relay addrs from CSV:", url);

    const response = await fetch(url);
    if (!response.ok) throw new Error(`Failed to fetch relay CSV (HTTP ${response.status})`);

    const csvText = await response.text();
    // Split by lines, always skip the first line (header)
    const lines = csvText.split('\n').map(line => line.trim()).filter(line => line.length > 0);
    const dataLines = lines.slice(1); // Always skip header

    // Each line is an address (or first column if comma separated), only keep valid multiaddrs
    const addrs = dataLines
      .map(line => line.split(',')[0].trim())
      .filter(addr => addr.startsWith('/'));

    console.log('[RelayAddrs] Valid relay addresses:', addrs);
    return addrs;
  } catch (error) {
    console.error("Error loading relay addresses:", error);
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
    setCommunities,
    isDarkMode,
    setCurrentCommunity,
    communities
  } = useAppStore();

  const [relayFailed, setRelayFailed] = useState(false);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    document.documentElement.classList.toggle("dark", isDarkMode);
  }, [isDarkMode]);
  
  useEffect(() => {
    TitleBarTheme(isDarkMode);
  }, []);

  useEffect(() => {
    EventsOn("navigate-to-root", () => {
      window.location.href = "/";
    });
  }, []);

  useEffect(() => {
    const initializeApp = async () => {
      try {
        console.log("🔄 Fetching relay addresses...");
        const relayAddrs = await fetchRelayAddrs();
        // const relayAddrs = ["/dns4/libr-relay007.onrender.com/tcp/443/wss/p2p/12D3KooWCG3Jp3Jm3AeD9WgUVAVeze71X3mPaCz2jfQAAGnCDgku"]
        const status = await GetRelayStatus();
        let connected = false;
        for (let i = 0; i < 10; i++) {
          if (status === "online") {
            connected = true;
            break;
          }

        const error = await Connect(relayAddrs);
          const recheck = await GetRelayStatus();
          if (recheck === "online") {
            connected = true;
            break;
          }

          await new Promise(res => setTimeout(res, 1000));
        }

        if (!connected) {
          console.error("❌ Could not connect to relay.");
          setRelayFailed(true);
          return;
        }

        console.log("✅ Relay connected. Authenticating...");
        
        const publicKey = await FetchPubKey();
        const user = await apiService.authenticate(publicKey);
        setUser(user);

        const fetchedCommunities = await apiService.getCommunities();
        setCommunities(fetchedCommunities);

        const firstJoined = fetchedCommunities.find(c => c.isJoined);
        if (firstJoined) {
          setCurrentCommunity(firstJoined);
        }
      } catch (err) {
        console.error("🔥 App initialization failed:", err);
        setRelayFailed(true);
      } finally {
        setLoading(false);
      }
    };

    initializeApp();
  }, []);

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
        <RelayErrorDialog open={relayFailed} onOpenChange={setRelayFailed} />
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
      <RelayErrorDialog open={relayFailed} onOpenChange={setRelayFailed} />
    </>
  );
};


export default App;
