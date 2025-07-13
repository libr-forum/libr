
import React from 'react';
import { motion } from 'framer-motion';
import { useAppStore } from '../../store/useAppStore';
import { useSidebarStore } from '../../store/useSidebarStore';
import { Shield, Users, Hash, Plus, Settings, Moon, Sun, RotateCcw, ChevronLeft, Eye, Menu } from 'lucide-react';
import { useNavigate, useLocation } from 'react-router-dom';

export const Sidebar: React.FC = () => {
  const { 
    communities, 
    currentCommunity, 
    setCurrentCommunity, 
    user, 
    isDarkMode, 
    toggleTheme,
    joinCommunity,
    setCommunities
  } = useAppStore();
  
  const { isCollapsed, setIsCollapsed } = useSidebarStore();
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
    <>
      <motion.div
        initial={{ x: -300, opacity: 0 }}
        animate={{ 
          x: isCollapsed ? -240 : 0, 
          opacity: 1,
          width: isCollapsed ? 80 : 320
        }}
        transition={{ duration: 0.3 }}
        className="h-screen bg-card border-r border-border/50 flex flex-col relative z-40"
      >
        {/* Collapse/Expand Button */}
        <button
          onClick={() => setIsCollapsed(!isCollapsed)}
          className="absolute -right-3 top-20 w-6 h-6 bg-libr-accent1 rounded-full flex items-center justify-center text-white shadow-lg hover:bg-libr-accent1/80 transition-colors z-50"
        >
          {isCollapsed ? <Menu className="w-3 h-3" /> : <ChevronLeft className="w-3 h-3" />}
        </button>
      {/* Header */}
      <div className="p-6 border-b border-border/50">
        <motion.div
          initial={{ scale: 0.8, opacity: 0 }}
          animate={{ scale: 1, opacity: 1 }}
          transition={{ delay: 0.2 }}
          className="flex items-center space-x-3"
        >
          <div className="w-10 h-10 bg-gradient-to-br from-libr-accent1 to-libr-accent2 rounded-lg flex items-center justify-center libr-glow">
            <Shield className="w-6 h-6 text-white" />
          </div>
          {!isCollapsed && (
            <div>
              <h1 className="text-xl font-bold text-libr-secondary">libr</h1>
              <p className="text-sm text-muted-foreground">Censorship-resistant</p>
            </div>
          )}
        </motion.div>
      </div>

      {/* User Info */}
      {user && !isCollapsed && (
        <div className="p-4 bg-muted/20">
          <div className="flex items-center space-x-3">
            <div className="w-8 h-8 bg-libr-accent1 rounded-full flex items-center justify-center">
              <span className="text-white text-sm font-medium">
                {user.alias.charAt(0).toUpperCase()}
              </span>
            </div>
            <div className="flex-1 min-w-0">
              <p className="text-sm font-medium text-foreground truncate">
                {user.alias}
              </p>
              <div className="flex items-center space-x-1">
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
        {!isCollapsed && (
          <div>
            <h3 className="text-sm font-semibold text-muted-foreground mb-3 flex items-center">
              <Users className="w-4 h-4 mr-2" />
              Joined Communities
            </h3>
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
                      currentCommunity?.id === community.id ? 'text-libr-accent1' : 'text-muted-foreground'
                    }`} />
                    <div className="flex-1 min-w-0">
                      <p className="font-medium text-foreground truncate">
                        {community.name}
                      </p>
                      <p className="text-xs text-muted-foreground">
                        {community.memberCount} members
                      </p>
                    </div>
                  </div>
                </motion.button>
              ))}
            </div>
          </div>
        )}

        {/* Collapsed view - only show icons */}
        {isCollapsed && (
          <div className="space-y-2">
            {communities.filter(c => c.isJoined).map((community) => (
              <motion.button
                key={community.id}
                whileHover={{ scale: 1.05 }}
                whileTap={{ scale: 0.95 }}
                onClick={() => handleCommunitySelect(community)}
                className={`w-12 h-12 rounded-lg transition-all duration-200 flex items-center justify-center ${
                  currentCommunity?.id === community.id
                    ? 'bg-libr-accent1/20 border border-libr-accent1/30'
                    : 'hover:bg-muted/50'
                }`}
                title={community.name}
              >
                <Hash className={`w-5 h-5 ${
                  currentCommunity?.id === community.id ? 'text-libr-accent1' : 'text-muted-foreground'
                }`} />
              </motion.button>
            ))}
          </div>
        )}

        {/* Available Communities */}
        {!isCollapsed && (
          <div>
            <div className="flex items-center justify-between mb-3">
              <h3 className="text-sm font-semibold text-muted-foreground flex items-center">
                Available Communities
              </h3>
              <button
                onClick={reloadCommunities}
                disabled={isReloadingCommunities}
                className="p-1 hover:bg-muted/50 rounded transition-colors"
                title="Reload Communities"
              >
                <RotateCcw className={`w-4 h-4 text-muted-foreground ${isReloadingCommunities ? 'animate-spin' : ''}`} />
              </button>
            </div>
            <div className="space-y-2">
              {availableCommunities.slice(0, 3).map((community) => (
                <motion.div
                  key={community.id}
                  initial={{ opacity: 0, y: 20 }}
                  animate={{ opacity: 1, y: 0 }}
                  className="p-3 rounded-lg border border-border/50 bg-card/50"
                >
                  <div className="flex items-center justify-between">
                    <div className="flex items-center space-x-3">
                      <Hash className="w-5 h-5 text-muted-foreground" />
                      <div>
                        <p className="font-medium text-foreground">
                          {community.name}
                        </p>
                        <p className="text-xs text-muted-foreground">
                          {community.memberCount} members
                        </p>
                      </div>
                    </div>
                    <button
                      onClick={() => handleJoinCommunity(community.id)}
                      className="libr-button bg-libr-accent1 text-white text-sm hover:bg-libr-accent1/80"
                    >
                      {community.requiresApproval ? 'Request' : 'Join'}
                    </button>
                  </div>
                </motion.div>
              ))}
              
              {showViewAll && (
                <button
                  onClick={() => navigate('/communities')}
                  className="w-full p-3 rounded-lg border border-border/50 bg-card/50 hover:bg-muted/50 transition-colors flex items-center justify-center space-x-2"
                >
                  <Eye className="w-4 h-4" />
                  <span>View All ({availableCommunities.length})</span>
                </button>
              )}
            </div>
          </div>
        )}

        {/* Moderator Section */}
        {user?.role === 'moderator' && !isCollapsed && (
          <div>
            <h3 className="text-sm font-semibold text-muted-foreground mb-3">
              Moderation
            </h3>
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
                  location.pathname === '/modlogs' ? 'text-libr-accent2' : 'text-muted-foreground'
                }`} />
                <span className="font-medium text-foreground">
                  Moderation Logs
                </span>
              </div>
            </motion.button>
          </div>
        )}

        {/* Collapsed Moderator Section */}
        {user?.role === 'moderator' && isCollapsed && (
          <div className="space-y-2">
            <motion.button
              whileHover={{ scale: 1.05 }}
              whileTap={{ scale: 0.95 }}
              onClick={() => navigate('/modlogs')}
              className={`w-12 h-12 rounded-lg transition-all duration-200 flex items-center justify-center ${
                location.pathname === '/modlogs'
                  ? 'bg-libr-accent2/20 border border-libr-accent2/30'
                  : 'hover:bg-muted/50'
              }`}
              title="Moderation Logs"
            >
              <Shield className={`w-5 h-5 ${
                location.pathname === '/modlogs' ? 'text-libr-accent2' : 'text-muted-foreground'
              }`} />
            </motion.button>
          </div>
        )}
      </div>

      {/* Footer */}
      <div className="p-4 border-t border-border/50">
        <div className={`flex items-center ${isCollapsed ? 'flex-col space-y-2' : 'justify-between'}`}>
          <button
            onClick={toggleTheme}
            className={`libr-button bg-muted hover:bg-muted/80 flex items-center ${isCollapsed ? 'w-12 h-12 justify-center' : 'space-x-2'}`}
            title={isDarkMode ? 'Switch to Light Mode' : 'Switch to Dark Mode'}
          >
            {isDarkMode ? <Sun className="w-4 h-4" /> : <Moon className="w-4 h-4" />}
            {!isCollapsed && <span>{isDarkMode ? 'Light' : 'Dark'}</span>}
          </button>
          <button 
            className={`libr-button bg-muted hover:bg-muted/80 ${isCollapsed ? 'w-12 h-12' : ''}`}
            title="Settings"
          >
            <Settings className="w-4 h-4" />
          </button>
        </div>
      </div>
      </motion.div>

      {/* Floating reappear button when collapsed */}
      {isCollapsed && (
        <motion.button
          initial={{ opacity: 0, scale: 0.8 }}
          animate={{ opacity: 1, scale: 1 }}
          onClick={() => setIsCollapsed(false)}
          className="fixed top-4 left-4 w-12 h-12 bg-libr-accent1 rounded-full flex items-center justify-center text-white shadow-lg hover:bg-libr-accent1/80 transition-colors z-50"
          title="Show Sidebar"
        >
          <Menu className="w-5 h-5" />
        </motion.button>
      )}
    </>
  );
};
