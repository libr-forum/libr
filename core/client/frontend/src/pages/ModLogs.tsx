
import React, { useEffect, useState } from 'react';
import { motion } from 'framer-motion';
import { useAppStore } from '../store/useAppStore';
import { apiService } from '../services/api';
import { Message } from '../store/useAppStore';
import { Shield, Check, X, Filter, Search, Clock, MessageSquare } from 'lucide-react';

export const ModLogs: React.FC = () => {
  const { user } = useAppStore();
  const [logs, setLogs] = useState<Message[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [filter, setFilter] = useState<'all' | 'pending' | 'approved' | 'rejected'>('all');
  const [searchTerm, setSearchTerm] = useState('');

  useEffect(() => {
    loadModerationLogs();
  }, []);

  const loadModerationLogs = async () => {
    setIsLoading(true);
    try {
      const moderationLogs = await apiService.getModerationLogs();
      setLogs(moderationLogs);
    } catch (error) {
      console.error('Failed to load moderation logs:', error);
    } finally {
      setIsLoading(false);
    }
  };

  const handleModerate = async (messageId: string, action: 'approve' | 'reject', note?: string) => {
    try {
      await apiService.moderateMessage(messageId, action, note);
      // Update local state
      setLogs(logs.map(log => 
        log.id === messageId 
          ? { ...log, status: action === 'approve' ? 'approved' : 'rejected', moderationNote: note }
          : log
      ));
    } catch (error) {
      console.error('Failed to moderate message:', error);
    }
  };

  const filteredLogs = logs.filter(log => {
    const matchesFilter = filter === 'all' || log.status === filter;
    const matchesSearch = searchTerm === '' || 
      log.content.toLowerCase().includes(searchTerm.toLowerCase()) ||
      log.authorAlias.toLowerCase().includes(searchTerm.toLowerCase());
    return matchesFilter && matchesSearch;
  });

  if (user?.role !== 'moderator') {
    return (
      <div className="flex-1 flex items-center justify-center bg-libr-primary">
        <motion.div
          initial={{ opacity: 0, scale: 0.9 }}
          animate={{ opacity: 1, scale: 1 }}
          className="text-center"
        >
          <div className="w-20 h-20 bg-red-500/20 rounded-2xl flex items-center justify-center mx-auto mb-4">
            <X className="w-10 h-10 text-red-500" />
          </div>
          <h2 className="text-xl font-semibold text-foreground mb-2">
            Access Denied
          </h2>
          <p className="text-muted-foreground">
            You need moderator privileges to access this page
          </p>
        </motion.div>
      </div>
    );
  }

  return (
    <div className="flex-1 flex flex-col bg-libr-primary h-screen">
      <div className="flex-1 overflow-y-auto">
        <div className="p-6 pb-24">
          <motion.div
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            className="max-w-6xl mx-auto"
          >
        {/* Header */}
        <div className="mb-8">
          <div className="flex items-center space-x-3 mb-4">
            <div className="w-12 h-12 bg-libr-accent2/20 rounded-xl flex items-center justify-center">
              <Shield className="w-6 h-6 text-libr-accent2" />
            </div>
            <div>
              <h1 className="text-2xl font-bold text-foreground">Moderation Logs</h1>
              <p className="text-muted-foreground">
                Review and moderate community messages
              </p>
            </div>
          </div>

          {/* Controls */}
          <div className="flex flex-col sm:flex-row gap-4">
            <div className="flex-1 relative">
              <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 w-4 h-4 text-muted-foreground" />
              <input
                type="text"
                placeholder="Search messages or authors..."
                value={searchTerm}
                onChange={(e) => setSearchTerm(e.target.value)}
                className="w-full pl-10 pr-4 py-2 bg-muted/30 border border-border/50 rounded-lg focus:outline-none focus:ring-2 focus:ring-libr-accent1/50"
              />
            </div>
            
            <div className="flex items-center space-x-2">
              <Filter className="w-4 h-4 text-muted-foreground" />
              <select
                value={filter}
                onChange={(e) => setFilter(e.target.value as any)}
                className="px-3 py-2 bg-muted/30 border border-border/50 rounded-lg focus:outline-none focus:ring-2 focus:ring-libr-accent1/50"
              >
                <option value="all">All Messages</option>
                <option value="pending">Pending</option>
                <option value="approved">Approved</option>
                <option value="rejected">Rejected</option>
              </select>
            </div>
          </div>
        </div>

        {/* Logs */}
        {isLoading ? (
          <div className="flex items-center justify-center py-12">
            <motion.div
              animate={{ rotate: 360 }}
              transition={{ duration: 1, repeat: Infinity, ease: "linear" }}
              className="w-8 h-8 border-2 border-libr-accent2 border-t-transparent rounded-full"
            />
          </div>
        ) : filteredLogs.length === 0 ? (
          <motion.div
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            className="text-center py-12"
          >
            <div className="w-16 h-16 bg-libr-accent2/20 rounded-2xl flex items-center justify-center mx-auto mb-4">
              <MessageSquare className="w-8 h-8 text-libr-accent2" />
            </div>
            <h3 className="text-lg font-medium text-foreground mb-2">
              No messages found
            </h3>
            <p className="text-muted-foreground">
              {filter === 'all' ? 'No moderation logs available' : `No ${filter} messages found`}
            </p>
          </motion.div>
        ) : (
          <div className="grid gap-4">
            {filteredLogs.map((log, index) => (
              <motion.div
                key={log.id}
                initial={{ opacity: 0, y: 20 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ delay: index * 0.05 }}
                className="libr-card p-6"
              >
                <div className="flex items-start justify-between">
                  <div className="flex-1">
                    <div className="flex items-center space-x-3 mb-3">
                      <div className="w-8 h-8 bg-libr-accent1 rounded-full flex items-center justify-center">
                        <span className="text-white text-sm font-medium">
                          {log.authorAlias.charAt(0).toUpperCase()}
                        </span>
                      </div>
                      <div>
                        <p className="font-medium text-foreground">{log.authorAlias}</p>
                        <div className="flex items-center space-x-2 text-sm text-muted-foreground">
                          <Clock className="w-3 h-3" />
                          <span>{new Date(log.timestamp).toLocaleString()}</span>
                        </div>
                      </div>
                      <div className="flex items-center space-x-2">
                        {log.status === 'pending' && (
                          <span className="px-2 py-1 bg-yellow-500/20 text-yellow-600 text-xs rounded-full">
                            Pending
                          </span>
                        )}
                        {log.status === 'approved' && (
                          <span className="px-2 py-1 bg-green-500/20 text-green-600 text-xs rounded-full">
                            Approved
                          </span>
                        )}
                        {log.status === 'rejected' && (
                          <span className="px-2 py-1 bg-red-500/20 text-red-600 text-xs rounded-full">
                            Rejected
                          </span>
                        )}
                      </div>
                    </div>
                    
                    <div className="bg-muted/20 rounded-lg p-4 mb-4">
                      <p className="text-foreground leading-relaxed">
                        {log.content}
                      </p>
                    </div>

                    {log.moderationNote && (
                      <div className="bg-libr-accent2/10 border-l-4 border-libr-accent2 pl-4 py-2 mb-4">
                        <p className="text-sm text-foreground">
                          <strong>Moderation Note:</strong> {log.moderationNote}
                        </p>
                      </div>
                    )}
                  </div>

                  {log.status === 'pending' && (
                    <div className="flex items-center space-x-2 ml-4">
                      <motion.button
                        whileHover={{ scale: 1.05 }}
                        whileTap={{ scale: 0.95 }}
                        onClick={() => handleModerate(log.id, 'approve')}
                        className="libr-button bg-green-500 hover:bg-green-600 text-white flex items-center space-x-1"
                      >
                        <Check className="w-4 h-4" />
                        <span>Approve</span>
                      </motion.button>
                      <motion.button
                        whileHover={{ scale: 1.05 }}
                        whileTap={{ scale: 0.95 }}
                        onClick={() => handleModerate(log.id, 'reject', 'Inappropriate content')}
                        className="libr-button bg-red-500 hover:bg-red-600 text-white flex items-center space-x-1"
                      >
                        <X className="w-4 h-4" />
                        <span>Reject</span>
                      </motion.button>
                    </div>
                  )}
                </div>
              </motion.div>
            ))}
          </div>
        )}
        </motion.div>
      </div>
    </div>
    </div>
  );
};
