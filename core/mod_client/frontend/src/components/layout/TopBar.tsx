
import React from 'react';
import { motion } from 'framer-motion';
import { useAppStore } from '../../store/useAppStore';
import { Hash, Users, Pin,Send,PencilLine,MessageCircle,MessageSquare } from 'lucide-react';
import { BrowserOpenURL } from '../../../wailsjs/runtime'

export const TopBar: React.FC = () => {
  const { currentCommunity, user } = useAppStore();

  if (!currentCommunity) return null;

  return (
    <motion.div
      initial={{ y: -50, opacity: 0 }}
      animate={{ y: 0, opacity: 1 }}
      className="h-16 bg-card border-b border-border/50 flex items-center justify-between px-5"
    >
      <div className="flex items-center space-x-4">
        <div className="flex items-center h-16 space-x-3">
          <div className="w-8 h-8 bg-libr-accent1/20 rounded-lg flex items-center justify-center">
            <Hash className="w-5 h-5 text-libr-accent1" />
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
        
        {currentCommunity.topic && (
          <div className="hidden md:flex items-center space-x-2 bg-muted/30 px-3 py-1 rounded-full">
            <Pin className="w-4 h-4 text-muted-foreground" />
            <span className="text-sm text-muted-foreground">
              {currentCommunity.topic}
            </span>
          </div>
        )}
      </div>
      <button
          onClick={() => BrowserOpenURL("https://forms.gle/Uchqc6Z49aoJwjvZ9")}
          className="libr-button bg-muted/30 hover:bg-muted/80 flex items-center space-x-3 translate-x-2"
        >
          <PencilLine className="w-5 h-5" />
          <span className='mt-0.5'>Feedback</span>
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
