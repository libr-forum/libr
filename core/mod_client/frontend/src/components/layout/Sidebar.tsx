import React from 'react';
import { motion } from 'framer-motion';
import { useAppStore } from '../../store/useAppStore';
import { useSidebarStore } from '../../store/useSidebarStore';
import { Shield, Cog, Hash, Plus, Settings, Moon, Sun, RefreshCcw , ChevronLeft, Eye, Menu, Wrench, User, AlertTriangle } from 'lucide-react';
import { useNavigate, useLocation } from 'react-router-dom';
import { RegenKeys } from 'wailsjs/go/main/App';
import logoSVG from '../assets/icon_transparent.svg';
import { apiService } from '@/services/api';
import { logger } from '../../logger/logger';
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
    logger.debug("Selected community", community);
    setCurrentCommunity(community);
    navigate('/chat');
  };

  const handleJoinCommunity = (communityId: string) => {
    logger.debug("Joining community", communityId);
    joinCommunity(communityId);
  };

  const reloadCommunities = async () => {
    logger.debug("Reloading communities...");
    setIsReloadingCommunities(true);
    try {
      // Simulate API call to reload communities
      await new Promise(resolve => setTimeout(resolve, 1000));
      // You would replace this with actual API call
      // const communities = await apiService.getCommunities();
      // setCommunities(communities);
    } catch (error) {
      logger.error("Failed to reload communities", { error });
    } finally {
      setIsReloadingCommunities(false);
    }
  };

  const availableCommunities = communities.filter(c => !c.isJoined);
  const showViewAll = availableCommunities.length > 3;
  const [showResetConfirm, setShowResetConfirm] = React.useState(false);
  const [confirmAlias, setConfirmAlias] = React.useState('');

  return (
  <div className='flex p-4 h-full w-full'>  
    <div className="flex flex-col shadow-md rounded-3xl z-50 w-full bg-card">
      {/* Header */}
      <div className="flex-col pt-5 pl-2 align-center items-center">
        <div className="flex items-center">
          <div className="w-16 h-16 rounded-lg flex items-center justify-center libr-glow">
            <img src={logoSVG} alt="Libr Logo" className="rounded-lg" />
          </div>
          <div className="flex translate-y-2 h-16">
            <h1 className="text-6xl text-libr-secondary">libr</h1>
          </div>
        </div>
        <div className='flex items-center pl-4 mt-2 h-14'>
            <span className="text-sm text-muted-foreground">Your Space.<br/>Your Quorum.<br/>Your Rules.</span>
        </div>
      </div>

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
            <div className="flex flex-col gap-3"> {/* Consistent spacing between buttons */}
              <motion.button
                whileHover={{ scale: 1.02 }}
                whileTap={{ scale: 0.98 }}
                onClick={() => navigate('/modlogs')}
                className={`w-full text-left p-3 rounded-lg transition-all duration-200 ${
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
                  <span className="font-medium text-foreground">
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
                  <span className="font-medium text-foreground">
                    Moderation Config
                  </span>
                </div>
              </motion.button>
              <motion.button
                whileHover={{ scale: 1.02 }}
                whileTap={{ scale: 0.98 }}
                onClick={() => navigate('/msgreports')}
                className={`w-full text-left p-3 rounded-lg transition-all duration-200 ${
                  location.pathname === '/msgreports'
                    ? 'bg-libr-accent2/20 border border-libr-accent2/30'
                    : 'hover:bg-muted/50'
                }`}
              >
                <div className="flex items-center space-x-3">
                  <AlertTriangle className={`w-5 h-5 ${
                    location.pathname === '/msgreports'
                      ? 'text-libr-accent2'
                      : 'text-muted-foreground'
                  }`} />
                  <span className="font-medium text-foreground">
                    Message Reports
                  </span>
                </div>
              </motion.button>
            </div>
          </div>
        )}
      </div>

      {/* User Info */}
      {/* {user && (
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
      )} */}

      {/* Footer */}
      {/* <div className="p-4">
        <div className="flex items-center justify-between">
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
      </div> */}
      {user && (
        <AlertDialog>
          <AlertDialogTrigger asChild>
            <button className="w-full text-left p-4 hover:bg-muted/50 transition-colors rounded-3xl">
              <div className="flex items-center space-x-3">
                {user.avatarSvg && user.avatarSvg !== "unknown" ? (
                  <img
                    src={`data:image/svg+xml;base64,${user.avatarSvg}`}
                    alt="avatar"
                    className="w-10 h-10 rounded-xl"
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
                          <span className="inline-flex items-center px-2 py-1 rounded-lg text-xs bg-libr-accent2/20 text-libr-accent2">
                            🛡 Moderator
                          </span>
                        );
                      }

                      if (currentRole === 'admin') {
                        return (
                          <span className="inline-flex items-center px-2 py-1 rounded-lg text-xs bg-red-500/20 text-red-500">
                            👑 Admin
                          </span>
                        );
                      }

                      return (
                        <span className="inline-flex items-center px-2 py-1 rounded-lg text-xs bg-libr-accent1/20 text-libr-accent1">
                          🗣 Member
                        </span>
                      );
                    })()}
                  </div>
                </div>
              </div>
            </button>
          </AlertDialogTrigger>

          {/* First Popup: User Info */}
          <AlertDialogContent className="bg-card border border-border/50 rounded-2xl shadow-xl text-foreground p-4 w-[90%] max-w-md">
            {/* <AlertDialogHeader>
              <AlertDialogTitle>Your Identity</AlertDialogTitle>
              <AlertDialogDescription className="mt-2">
                This is your current alias and role.
              </AlertDialogDescription>
            </AlertDialogHeader> */}

            <div className="flex flex-row space-x-4 items-center">
                {user.avatarSvg && user.avatarSvg !== "unknown" ? (
                  <img
                    src={`data:image/svg+xml;base64,${user.avatarSvg}`}
                    alt="avatar"
                    className="w-24 h-24 rounded-xl"
                  />
                ) : (
                  <div className="w-16 h-16 bg-libr-accent1 rounded-full flex items-center justify-center">
                    <span className="text-white text-lg font-medium">
                      {user.alias.charAt(0).toUpperCase()}
                    </span>
                  </div>
                )}
              <div className='flex flex-col space-y-1'>
                <p className="font-medium text-2xl">{user.alias}</p>
                <div className='text-lg'>
                  {(() => {
                    const currentRole = currentCommunity?.id
                      ? user.communityRoles?.[currentCommunity.id] || user.role
                      : user.role;
                    if (currentRole === 'moderator') {
                      return <span className="inline-flex items-center px-2 rounded-lg bg-libr-accent2/20 text-libr-accent2">🛡 Moderator</span>;
                    }
                    if (currentRole === 'admin') {
                      return <span className="inline-flex items-center px-2 rounded-lg bg-red-500/20 text-red-500">👑 Admin</span>;
                    }
                    return <span className="inline-flex items-center px-2 rounded-lg bg-libr-accent1/20 text-libr-accent1">🗣 Member</span>;
                  })()}
                </div>
              </div>
            </div>
            
            <AlertDialogFooter>
              <div className='flex justify-between w-full'>
                <button
                  onClick={() => {
                    setShowResetConfirm(true);
                    logger.debug("Reset identity initiated", { alias: user.alias });

                  }}
                  className="libr-button bg-red-500 hover:bg-red-600 text-white"
                >
                  Reset Identity
                </button>
                <AlertDialogCancel className="libr-button bg-muted hover:bg-muted/70">
                  Close
                </AlertDialogCancel>
              </div>
            </AlertDialogFooter>
          </AlertDialogContent>
        </AlertDialog>
      )}
      <AlertDialog open={showResetConfirm} onOpenChange={setShowResetConfirm}>
        <AlertDialogContent className="bg-card border border-border/50 rounded-xl shadow-xl text-red-600 p-6 w-[90%] max-w-md">
          <AlertDialogHeader>
            <AlertDialogTitle>Confirm Identity Reset</AlertDialogTitle>
            <AlertDialogDescription className='text-red-600'>
              This action is <b>irreversible</b>. Type your full username to confirm reset.
            </AlertDialogDescription>
          </AlertDialogHeader>

          <input
            type="text"
            className="mt-4 w-full px-3 py-2 border border-border rounded-md bg-background text-foreground"
            placeholder="Enter your username"
            value={confirmAlias}
            onChange={(e) => setConfirmAlias(e.target.value)}
          />

          <AlertDialogFooter className="mt-4">
            <AlertDialogCancel
              className="libr-button bg-muted hover:bg-muted/70"
              onClick={() => {
                logger.error("Identity reset failed", { error });
                setShowResetConfirm(false);
                setConfirmAlias('');
              }}
            >
              Cancel
            </AlertDialogCancel>
            <AlertDialogAction
              disabled={confirmAlias !== user?.alias}
              className={`libr-button ${
                confirmAlias === user?.alias
                  ? 'bg-red-500 hover:bg-red-600 text-white'
                  : 'bg-muted text-muted-foreground cursor-not-allowed'
              }`}
              onClick={async () => {
                try {
                  const newPubKey = await RegenKeys();
                  const newUser = await apiService.authenticate(newPubKey);
                  setUser(newUser);
                } catch (error) {
                  console.error("Reset failed:", error);
                } finally {
                  setShowResetConfirm(false);
                  setConfirmAlias('');
                }
              }}
            >
              Confirm Reset
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  </div>
);
};
