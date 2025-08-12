
import React from 'react';
import { motion } from 'framer-motion';
import { useAppStore } from '../../store/useAppStore';
import { Hash, Users, Pin,Send,PencilLine,MessageCircle,MessageSquare,Sun,Moon } from 'lucide-react';
import { logger } from '../../logger/logger';

export const TopBar: React.FC = () => {
  
  const { currentCommunity, user,isDarkMode,toggleTheme } = useAppStore();
  logger.debug("TopBar: No current community, skipping render");
  if (!currentCommunity) return null;

  return (
    <motion.div
      initial={{ y: -50, opacity: 0 }}
      animate={{ y: 0, opacity: 1 }}
      className="bg-card border-b border-border/50 flex items-center rounded-3xl shadow-md justify-between px-5 h-full z-50"
    >
      <div className="flex items-center h-[100%] justify-between space-x-3 rounded-xl">
        <div className="aspect-square h-[40%] bg-libr-accent1/20 rounded-lg flex items-center justify-center">
          <Hash className="w-[62%] h-[62%] text-libr-accent1" />
        </div>
        <div>
          <h2 className="text-lg mt-1 font-semibold text-foreground">
            {currentCommunity.name}
          </h2>
          {/* <div className="flex items-center space-x-2 text-sm text-muted-foreground">
            <Users className="w-4 h-4" />
            <span>{currentCommunity.memberCount} members</span>
          </div> */}
        </div>
      </div>
        {/* <button
          onClick={() => BrowserOpenURL("https://forms.gle/Uchqc6Z49aoJwjvZ9")}
          className="libr-button bg-muted/30 hover:bg-muted/80 flex items-center space-x-3 translate-x-2"
        >
          <PencilLine className="w-5 h-5" />
          <span className='mt-0.5'>Feedback</span>
        </button> */}
        <button
          onClick={() => {
            logger.debug("Toggling theme", { from: isDarkMode ? "Dark" : "Light" });
            toggleTheme();
          }}
          className="libr-button rounded-3xl hover:bg-muted/80 flex items-center space-x-2 translate-x-2"
          title={isDarkMode ? 'Switch to Light Mode' : 'Switch to Dark Mode'}
        >
          {isDarkMode ? <Sun className="w-4 h-4" /> : <Moon className="w-4 h-4" />}
          <span className='mt-0.5'>{isDarkMode ? 'Light' : 'Dark'}</span>
        </button>
      {/* <div className="flex items-center space-x-3">
        {user && (
          <div className="flex items-center space-x-2">
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

            <div className="hidden sm:block">
              <p className="text-sm font-medium text-foreground">{user.alias}</p>
              <div className="flex items-center space-x-1">
                {user.role === 'moderator' && (
                  <span className="text-xs text-libr-accent2">🛡 Mod</span>
                )}
              </div>
            </div>
          </div>
        )}
      </div> */}
    </motion.div>
  );
};
