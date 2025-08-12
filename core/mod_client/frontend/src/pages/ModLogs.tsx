import React, { useEffect, useState } from 'react';
import { motion } from 'framer-motion';
import { useAppStore } from '../store/useAppStore';
import { apiService } from '../services/api';
import { Message,ModLogEntry } from '../store/useAppStore';
import { Sidebar } from '@/components/layout/Sidebar';
import { Shield, Check, X, Filter, Search, Clock, MessageSquare } from 'lucide-react';

export const ModLogs: React.FC = () => {
  const { user } = useAppStore();
  const [logs, setLogs] = useState<ModLogEntry[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [filter, setFilter] = useState<'all' | 'approved' | 'rejected'>('all');
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

  const filteredLogs = logs.filter(log => {
    const matchesFilter =
      filter === 'all' ||
      (filter === 'approved' && log.status === '1') ||
      (filter === 'rejected' && log.status === '0');
    const matchesSearch =
      searchTerm === '' ||
      log.content.toLowerCase().includes(searchTerm.toLowerCase());
    return matchesFilter && matchesSearch;
  });

  if (user?.role !== 'moderator') {
    return (
      <div className='flex flex-row'>
        <div className='w-[19.4%]'>
          <Sidebar />
        </div>
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
      </div>
    );
  }

  return (
    <div className='flex flex-row'>
      <div className='w-[19.4%]'>
        <Sidebar />
      </div>
      <div className="flex-1 flex flex-col w-full bg-libr-primary h-screen">
        <div className="flex-1 overflow-y-auto">
          <div className="pt-6 pb-24">
            <motion.div
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              className="max-w-7xl mx-4"
            >
              {/* Header */}
              <div className="mb-8 w-full">
                <div className="flex items-center space-x-3 mb-4">
                  <div className="w-12 h-12 bg-libr-accent2/20 rounded-xl flex items-center justify-center">
                    <Shield className="w-6 h-6 text-libr-accent2" />
                  </div>
                  <div>
                    <h1 className="text-2xl font-bold text-foreground">Moderation Logs</h1>
                    <p className="text-muted-foreground">
                      Automatically recorded logs from your moderation
                    </p>
                  </div>
                </div>

                {/* Controls */}
                <div className="flex flex-col sm:flex-row gap-4">
                  <div className="flex-1 relative">
                    <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 w-4 h-4 text-muted-foreground" />
                    <input
                      type="text"
                      placeholder="Search messages..."
                      value={searchTerm}
                      onChange={(e) => setSearchTerm(e.target.value)}
                      className="w-full pl-10 pr-4 py-2 bg-muted/30 border border-border/50 rounded-lg focus:outline-none focus:ring-2 focus:ring-libr-accent1/50"
                    />
                  </div>
                  <button
                    onClick={loadModerationLogs}
                    className="px-3 py-2 bg-muted/30 border border-border/50 text-base rounded-lg text-foreground hover:bg-muted/40 transition-colors"
                  >
                    Refresh
                  </button>
                  <div className="flex items-center space-x-2">
                    <Filter className="w-4 h-4 text-muted-foreground" />
                    <select
                      value={filter}
                      onChange={(e) => setFilter(e.target.value as any)}
                      className="px-3 py-2 bg-muted/30 border border-border/50 rounded-lg focus:outline-none focus:ring-2 focus:ring-libr-accent1/50"
                    >
                      <option value="all">All Messages</option>
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
                    transition={{ duration: 1, repeat: Infinity, ease: 'linear' }}
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
                    {filter === 'all'
                      ? 'No moderation logs available'
                      : `No ${filter} messages found`}
                  </p>
                </motion.div>
              ) : (
                <div className="grid gap-3">
                  {filteredLogs.map((log, index) => (
                    <motion.div
                      key={index}
                      initial={{ opacity: 0, y: 20 }}
                      animate={{ opacity: 1, y: 0 }}
                      transition={{ delay: index * 0.03 }}
                      className="bg-muted/30 border border-border/40 rounded-xl p-4"
                    >
                      <div className="flex justify-between items-start">
                        <div className="flex items-center space-x-3">
                          <div className="flex items-center text-xs text-muted-foreground space-x-1">
                            <Clock className="w-3 h-3" />
                            <span>{new Date(log.timestamp * 1000).toLocaleString()}</span>
                          </div>
                        </div>
                        <div>
                          {log.status === '1' ? (
                            <span className="px-2 py-0.5 text-xs rounded-full bg-green-500/20 text-green-600">
                              Approved
                            </span>
                          ) : (
                            <span className="px-2 py-0.5 text-xs rounded-full bg-red-500/20 text-red-600">
                              Rejected
                            </span>
                          )}
                        </div>
                      </div>

                      <div className="mt-3 bg-background/50 p-3 rounded-md">
                        {(() => {
                          const titleMatch = log.content.match(/<HEAD>(.*?)<\/HEAD>/s);
                          const bodyMatch = log.content.match(/<BODY>(.*?)<\/BODY>/s);
                          const title = titleMatch?.[1]?.trim() || '';
                          const rawBody = bodyMatch?.[1]?.trim() || '';

                          const cleanBody = rawBody
                            .replace(/<\/p>\s*<p>/g, '<br><br>') // Convert paragraphs to double line break
                            .replace(/^<p>/, '')
                            .replace(/<\/p>$/, '');

                          return (
                            <>
                              {title && (
                                <h4 className="text-sm font-semibold text-foreground mb-1">
                                  {title}
                                </h4>
                              )}
                              <div
                                className="text-sm text-foreground leading-snug whitespace-pre-wrap [&_strong]:font-semibold [&_u]:underline [&_em]:italic modlog-body-content"
                                dangerouslySetInnerHTML={{ __html: cleanBody }}
                              />
                            </>
                          );
                        })()}
                      </div>
                      {/* Add bullet styling for modlog content */}
                      <style>
                        {`
                          .modlog-body-content ul {
                            list-style-type: disc;
                            margin-left: 1.5em;
                            padding-left: 1.5em;
                          }
                          .modlog-body-content ol {
                            list-style-type: decimal;
                            margin-left: 1.5em;
                            padding-left: 1.5em;
                          }
                          .modlog-body-content li {
                            margin-bottom: 0.25em;
                          }
                        `}
                      </style>
                    </motion.div>
                  ))}
                </div>
              )}
            </motion.div>
          </div>
        </div>
      </div>
    </div>
  );
};