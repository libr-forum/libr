
import React from 'react';
import { motion } from 'framer-motion';
import { useAppStore } from '../../store/useAppStore';
import { useSidebarStore } from '../../store/useSidebarStore';
import { Shield, UserCog, Hash, Plus, Settings, Moon, Sun, RefreshCcw , ChevronLeft, Eye, Menu, Wrench, User } from 'lucide-react';
import { useNavigate, useLocation } from 'react-router-dom';
import { RegenKeys } from 'wailsjs/go/main/App';
import logoSVG from '../assets/icon_transparent-01.svg';
import { apiService } from '@/services/api';
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

export const Sidebar: React.FC = () => {
  const { 
    communities, 
    currentCommunity, 
    setCurrentCommunity, 
    setUser,
    user, 
    isDarkMode, 
    toggleTheme,
    joinCommunity,
    setCommunities
  } = useAppStore();
  
  const navigate = useNavigate();
  const location = useLocation();
  const [isReloadingCommunities, setIsReloadingCommunities] = React.useState(false);

  const handleCommunitySelect = (community: any) => {
    setCurrentCommunity(community);
    navigate('/chat');
  };

  const handleJoinCommunity = (communityId: string) => {
    joinCommunity(communityId);
  };

  const reloadCommunities = async () => {
    setIsReloadingCommunities(true);
    try {
      // Simulate API call to reload communities
      await new Promise(resolve => setTimeout(resolve, 1000));
      // You would replace this with actual API call
      // const communities = await apiService.getCommunities();
      // setCommunities(communities);
    } catch (error) {
      console.error('Failed to reload communities:', error);
    } finally {
      setIsReloadingCommunities(false);
    }
  };

  const availableCommunities = communities.filter(c => !c.isJoined);
  const showViewAll = availableCommunities.length > 3;

  return (
  <div className="flex flex-col h-full w-80 border-r border-border/50 bg-card">
    {/* Header */}
    <div className="flex-col pt-5  pl-2 align-center items-center h-32 border-b border-border/50">
      <div className="flex items-center w-76">
        <div className="w-16 h-16 rounded-lg flex items-center justify-center libr-glow">
          <img src={logoSVG} alt="Libr Logo" className="rounded-lg" />
        </div>
        <div className="flex translate-y-2 h-16">
          <h1 className="text-6xl text-libr-secondary">libr</h1>
        </div>
      </div>
      <div className='flex items-center mt-2 pl-4 w-60 h-6'>
          <p className="text-sm text-muted-foreground">Censorship-resilient</p>
      </div>
    </div>

    {/* User Info */}
    {user && (
      <div className="p-4">
        <div className="flex items-center space-x-3">
          {user.avatarSvg && user.avatarSvg !== "unknown" ? (
            <img
              src={`data:image/svg+xml;base64,${user.avatarSvg}`}
              alt="avatar"
              className="w-10 h-10 rounded-full"
            />
          ) : (
            <div className="w-10 h-10 bg-libr-accent1 rounded-full flex items-center justify-center">
              <span className="text-white text-sm font-medium">
                {user.alias.charAt(0).toUpperCase()}
              </span>
            </div>
          )}
          <div className="flex-1 min-w-0">
            <p className="text-sm font-medium text-foreground truncate">
              {user.alias}
            </p>
            <div className="flex items-center mt-1 space-x-1">
              {(() => {
                const currentRole = currentCommunity?.id
                  ? user.communityRoles?.[currentCommunity.id] || user.role
                  : user.role;

                if (currentRole === 'moderator') {
                  return (
                    <span className="inline-flex items-center px-2 py-1 rounded-full text-xs bg-libr-accent2/20 text-libr-accent2">
                      🛡 Moderator
                    </span>
                  );
                }

                if (currentRole === 'admin') {
                  return (
                    <span className="inline-flex items-center px-2 py-1 rounded-full text-xs bg-red-500/20 text-red-500">
                      👑 Admin
                    </span>
                  );
                }

                return (
                  <span className="inline-flex items-center px-2 py-1 rounded-full text-xs bg-libr-accent1/20 text-libr-accent1">
                    🗣 Member
                  </span>
                );
              })()}
            </div>
          </div>
        </div>
      </div>
    )}

    {/* Navigation */}
    <div className="flex-1 overflow-y-auto p-4 space-y-6">
      {/* Joined Communities */}
      <div className="space-y-2">
        {communities.filter(c => c.isJoined).map((community) => (
          <motion.button
            key={community.id}
            whileHover={{ scale: 1.02 }}
            whileTap={{ scale: 0.98 }}
            onClick={() => handleCommunitySelect(community)}
            className={`w-full text-left p-3 rounded-lg transition-all duration-200 ${
              currentCommunity?.id === community.id
                ? 'bg-libr-accent1/20 border border-libr-accent1/30'
                : 'hover:bg-muted/50'
            }`}
          >
            <div className="flex items-center space-x-3">
              <Hash className={`w-5 h-5 ${
                currentCommunity?.id === community.id
                  ? 'text-libr-accent1'
                  : 'text-muted-foreground'
              }`} />
              <div className="flex-1 min-w-0">
                <p className="font-medium text-foreground mt-1 truncate">
                  {community.name}
                </p>
              </div>
            </div>
          </motion.button>
        ))}
      </div>

      {/* Moderator Section */}
      {user?.role === 'moderator' && (
        <div>
          <h3 className="text-sm font-semibold text-muted-foreground mb-3">
            Moderation
          </h3>
          <motion.button
            whileHover={{ scale: 1.02 }}
            whileTap={{ scale: 0.98 }}
            onClick={() => navigate('/modlogs')}
            className={`w-full text-left p-3 rounded-lg transition-all duration-200 mb-2 ${
              location.pathname === '/modlogs'
                ? 'bg-libr-accent2/20 border border-libr-accent2/30'
                : 'hover:bg-muted/50'
            }`}
          >
            <div className="flex items-center space-x-3">
              <Shield className={`w-5 h-5 ${
                location.pathname === '/modlogs'
                  ? 'text-libr-accent2'
                  : 'text-muted-foreground'
              }`} />
              <span className="font-medium mt-1 text-foreground">
                Moderation Logs
              </span>
            </div>
          </motion.button>
          <motion.button
            whileHover={{ scale: 1.02 }}
            whileTap={{ scale: 0.98 }}
            onClick={() => navigate('/modconfig')}
            className={`w-full text-left p-3 rounded-lg transition-all duration-200 ${
              location.pathname === '/modconfig'
                ? 'bg-libr-accent2/20 border border-libr-accent2/30'
                : 'hover:bg-muted/50'
            }`}
          >
            <div className="flex items-center space-x-3">
              <Wrench className={`w-5 h-5 ${
                location.pathname === '/modconfig'
                  ? 'text-libr-accent2'
                  : 'text-muted-foreground'
              }`} />
              <span className="font-medium mt-1 text-foreground">
                Moderation Config
              </span>
            </div>
          </motion.button>
        </div>
      )}
    </div>

    {/* Footer */}
    <div className="p-4">
      <div className="flex items-center justify-between">
        {/* <button
          onClick={async () => {
            try {
              const newPubKey = await RegenKeys();
              const user = await apiService.authenticate(newPubKey);
              setUser(user);
            } catch (error) {
              console.error("Authentication failed:", error);
            }
          }}
          className="libr-button bg-muted hover:bg-muted/80 flex items-center space-x-2"
        >
          <UserCog className="w-4 h-4"/>
          <span className='mt-0.5'>Reset Identity</span>
        </button> */}
        <AlertDialog>
          <AlertDialogTrigger asChild>
            <button
              className="libr-button bg-muted hover:bg-muted/80 flex items-center space-x-2"
            >
              <UserCog className="w-4 h-4" />
              <span className='mt-0.5'>Reset Identity</span>
            </button>
          </AlertDialogTrigger>

          <AlertDialogContent className="bg-card border border-border/50 rounded-xl shadow-xl text-foreground p-6 w-[90%] max-w-md">
            <AlertDialogHeader>
              <AlertDialogTitle>Are you sure?</AlertDialogTitle>
              <AlertDialogDescription>
                Resetting your identity will generate a new key pair. You won't be able to restore the previous identity.
              </AlertDialogDescription>
            </AlertDialogHeader>
            <AlertDialogFooter>
              <AlertDialogCancel className="libr-button bg-muted hover:bg-muted/70">Cancel</AlertDialogCancel>
              <AlertDialogAction className="libr-button bg-libr-accent1 hover:bg-libr-accent1/90 text-white"
                onClick={async () => {
                  try {
                    const newPubKey = await RegenKeys();
                    const user = await apiService.authenticate(newPubKey);
                    setUser(user);
                  } catch (error) {
                    console.error("Authentication failed:", error);
                  }
                }}
              >
                Confirm
              </AlertDialogAction>
            </AlertDialogFooter>
          </AlertDialogContent>
        </AlertDialog>


        <button
          onClick={toggleTheme}
          className="libr-button bg-muted hover:bg-muted/80 flex items-center space-x-2"
          title={isDarkMode ? 'Switch to Light Mode' : 'Switch to Dark Mode'}
        >
          {isDarkMode ? <Sun className="w-4 h-4" /> : <Moon className="w-4 h-4" />}
          <span className='mt-0.5'>{isDarkMode ? 'Light' : 'Dark'}</span>
        </button>
        
      </div>
    </div>
  </div>
);

};
