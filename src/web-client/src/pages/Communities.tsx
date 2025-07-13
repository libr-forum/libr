import React from 'react';
import { motion } from 'framer-motion';
import { useAppStore } from '../store/useAppStore';
import { Hash, Users, ArrowLeft, Search, Filter, RotateCcw } from 'lucide-react';
import { useNavigate } from 'react-router-dom';

export const Communities: React.FC = () => {
  const { communities, joinCommunity } = useAppStore();
  const navigate = useNavigate();
  const [searchTerm, setSearchTerm] = React.useState('');
  const [filter, setFilter] = React.useState<'all' | 'needsApproval' | 'openJoin'>('all');
  const [isReloading, setIsReloading] = React.useState(false);

  const availableCommunities = communities.filter(c => !c.isJoined);

  const filteredCommunities = React.useMemo(() => {
    let filtered = availableCommunities;

    if (searchTerm) {
      filtered = filtered.filter(community =>
        community.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
        community.topic.toLowerCase().includes(searchTerm.toLowerCase())
      );
    }

    if (filter === 'needsApproval') {
      filtered = filtered.filter(community => community.requiresApproval);
    } else if (filter === 'openJoin') {
      filtered = filtered.filter(community => !community.requiresApproval);
    }

    return filtered;
  }, [availableCommunities, searchTerm, filter]);

  const handleJoinCommunity = (communityId: string) => {
    joinCommunity(communityId);
  };

  const handleReload = async () => {
    setIsReloading(true);
    try {
      // Simulate API call
      await new Promise(resolve => setTimeout(resolve, 1000));
      // Replace with actual API call
    } catch (error) {
      console.error('Failed to reload communities:', error);
    } finally {
      setIsReloading(false);
    }
  };

  return (
    <div className="flex-1 flex flex-col bg-libr-primary h-screen">
      <div className="flex-1 overflow-y-auto">
        <div className="p-6 pb-24">
          <motion.div
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            className="max-w-4xl mx-auto"
          >
        {/* Header */}
        <div className="flex items-center justify-between mb-6">
          <div className="flex items-center space-x-4">
            <button
              onClick={() => navigate(-1)}
              className="libr-button bg-muted hover:bg-muted/80 p-2"
            >
              <ArrowLeft className="w-5 h-5" />
            </button>
            <div>
              <h1 className="text-2xl font-bold text-foreground">All Communities</h1>
              <p className="text-muted-foreground">
                {filteredCommunities.length} of {availableCommunities.length} communities available to join
              </p>
            </div>
          </div>
          
          <button
            onClick={handleReload}
            disabled={isReloading}
            className="libr-button bg-libr-accent1 text-white hover:bg-libr-accent1/80 flex items-center space-x-2"
          >
            <RotateCcw className={`w-4 h-4 ${isReloading ? 'animate-spin' : ''}`} />
            <span>Refresh</span>
          </button>
        </div>

        {/* Search and Filters */}
        <div className="flex flex-col md:flex-row gap-4 mb-6">
          <div className="flex-1 relative">
            <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 w-4 h-4 text-muted-foreground" />
            <input
              type="text"
              placeholder="Search communities..."
              value={searchTerm}
              onChange={(e) => setSearchTerm(e.target.value)}
              className="w-full pl-10 pr-4 py-2 bg-muted border border-border rounded-lg focus:outline-none focus:ring-2 focus:ring-libr-accent1/50"
            />
          </div>
          
          <div className="flex items-center space-x-2">
            <Filter className="w-4 h-4 text-muted-foreground" />
            <select
              value={filter}
              onChange={(e) => setFilter(e.target.value as any)}
              className="bg-muted border border-border rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-libr-accent1/50"
            >
              <option value="all">All Communities</option>
              <option value="openJoin">Open Join</option>
              <option value="needsApproval">Needs Approval</option>
            </select>
          </div>
        </div>

        {/* Communities Grid */}
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
          {filteredCommunities.map((community, index) => (
            <motion.div
              key={community.id}
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ delay: index * 0.05 }}
              className="bg-card border border-border/50 rounded-lg p-4 hover:border-libr-accent1/30 transition-all duration-200"
            >
              <div className="flex items-start justify-between mb-3">
                <div className="flex items-center space-x-3">
                  <div className="w-10 h-10 bg-libr-accent1/20 rounded-lg flex items-center justify-center">
                    <Hash className="w-5 h-5 text-libr-accent1" />
                  </div>
                  <div>
                    <h3 className="font-semibold text-foreground">{community.name}</h3>
                    <div className="flex items-center space-x-1 text-xs text-muted-foreground">
                      <Users className="w-3 h-3" />
                      <span>{community.memberCount} members</span>
                    </div>
                  </div>
                </div>
                
                {community.requiresApproval && (
                  <span className="text-xs bg-yellow-500/20 text-yellow-600 px-2 py-1 rounded-full">
                    Approval Required
                  </span>
                )}
              </div>
              
              <p className="text-sm text-muted-foreground mb-4 line-clamp-2">
                {community.topic}
              </p>
              
              <button
                onClick={() => handleJoinCommunity(community.id)}
                className="w-full libr-button bg-libr-accent1 text-white hover:bg-libr-accent1/80"
              >
                {community.requiresApproval ? 'Request to Join' : 'Join Community'}
              </button>
            </motion.div>
          ))}
        </div>

        {filteredCommunities.length === 0 && (
          <motion.div
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            className="text-center py-12"
          >
            <div className="w-16 h-16 bg-libr-accent1/20 rounded-2xl flex items-center justify-center mx-auto mb-4">
              <Hash className="w-8 h-8 text-libr-accent1" />
            </div>
            <h3 className="text-lg font-medium text-foreground mb-2">
              No Communities Found
            </h3>
            <p className="text-muted-foreground">
              {searchTerm || filter !== 'all' 
                ? 'Try adjusting your search or filters'
                : 'All communities have been joined or none are available'
              }
            </p>
          </motion.div>
        )}
          </motion.div>
        </div>
      </div>
    </div>
  );
};
