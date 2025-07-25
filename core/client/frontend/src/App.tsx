
import React, { useEffect } from 'react';
import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import { motion, AnimatePresence } from 'framer-motion';
import { Toaster } from "@/components/ui/toaster";
import { Toaster as Sonner } from "@/components/ui/sonner";
import { TooltipProvider } from "@/components/ui/tooltip";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";

import { useAppStore } from './store/useAppStore';
import { apiService } from './services/api';
import { Sidebar } from './components/layout/Sidebar';
import { ChatRoom } from './pages/ChatRoom';
import { ModLogs } from './pages/ModLogs';
import { Communities } from './pages/Communities';

import { Connect,GetRelayStatus,FetchPubKey } from '../wailsjs/go/main/App';

const queryClient = new QueryClient();

const csvURL = 'https://raw.githubusercontent.com/cherry-aggarwal/LIBR/refs/heads/integration/docs/network.csv';

async function fetchRelayAddr():Promise<string> {
  try {
    const response = await fetch(csvURL);
    if (!response.ok) throw new Error('Failed to fetch CSV');
    const text = await response.text();
    const lines = text.trim().split('\n');
    if (lines.length < 2) throw new Error('CSV does not have enough rows');
    const firstDataRow = lines[1].split(',')[0]; // adjust index if needed
    return firstDataRow;
  } catch (error) {
      console.error('Error loading relay address:', error);
      return '';
    }
};

const Layout: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  return (
    <div className="flex h-screen bg-libr-primary relative">
      <Sidebar />
      <main className="flex-1 overflow-hidden transition-all duration-300">
        {children}
      </main>
    </div>
  );
};

const App: React.FC = () => {
  const { 
    isAuthenticated, 
    setUser, 
    setCommunities, 
    isDarkMode,
    setCurrentCommunity,
    communities
  } = useAppStore();

  useEffect(() => {
    // Initialize theme
    document.documentElement.classList.toggle('dark', isDarkMode);
  }, [isDarkMode]);

  useEffect(() => {
    // Initialize app
    const initializeApp = async () => {
      try {
        //Relay connection
        const relayAddr = await fetchRelayAddr();
        const status = await GetRelayStatus();
        if (status !== "online") {
          const result = await Connect(relayAddr);
          console.log("Relay status:", result);
        }

        // Mock authentication with a demo public key
        const PublicKey=await FetchPubKey()
        const user = await apiService.authenticate(PublicKey);
        setUser(user);
        
        // Load communities
        const fetchedCommunities = await apiService.getCommunities();
        setCommunities(fetchedCommunities);

        // Set initial community (first joined community)
        const firstJoinedCommunity = fetchedCommunities.find(c => c.isJoined);
        if (firstJoinedCommunity) {
          setCurrentCommunity(firstJoinedCommunity);
        }
      } catch (error) {
        console.error('Failed to initialize app:', error);
      }
    };

    initializeApp();
  }, []);

  if (!isAuthenticated) {
    return (
      <div className="min-h-screen bg-libr-primary flex items-center justify-center">
        <motion.div
          initial={{ opacity: 0, scale: 0.9 }}
          animate={{ opacity: 1, scale: 1 }}
          className="text-center"
        >
          <div className="w-20 h-20 bg-gradient-to-br from-libr-accent1 to-libr-accent2 rounded-2xl flex items-center justify-center mx-auto mb-6 libr-glow">
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
    );
  }

  return (
    <QueryClientProvider client={queryClient}>
      <TooltipProvider>
        <Toaster />
        <Sonner />
        <BrowserRouter>
          <Layout>
            <AnimatePresence mode="wait">
              <Routes>
                <Route path="/" element={<Navigate to="/chat" replace />} />
                <Route path="/chat" element={<ChatRoom />} />
                <Route path="/communities" element={<Communities />} />
                <Route path="/modlogs" element={<ModLogs />} />
                <Route path="*" element={<Navigate to="/chat" replace />} />
              </Routes>
            </AnimatePresence>
          </Layout>
        </BrowserRouter>
      </TooltipProvider>
    </QueryClientProvider>
  );
};

export default App;
