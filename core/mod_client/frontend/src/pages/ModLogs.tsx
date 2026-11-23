import React, { useEffect, useState } from 'react';
import { motion } from 'framer-motion';
import { useAppStore } from '../store/useAppStore';
import { apiService } from '../services/api';
import { Message,ModLogEntry } from '../store/useAppStore';
import { Sidebar } from '@/components/layout/Sidebar';
import { Shield, Check, X, Filter, Search, Clock, MessageSquare, RotateCcw } from 'lucide-react';

export const ModLogs: React.FC = () => {
  const { user } = useAppStore();
  const [logs, setLogs] = useState<ModLogEntry[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [filter, setFilter] = useState<'all' | 'approved' | 'rejected'>('all');
  const [searchTerm, setSearchTerm] = useState('');

  const loadModerationLogs = async (background: boolean = false) => {
    if (!background) setIsLoading(true);
    try {
      const moderationLogs = await apiService.getModerationLogs();
      setLogs(moderationLogs);
    } catch (error) {
      console.error('Failed to load moderation logs:', error);
    } finally {
      if (!background) setIsLoading(false);
    }
  };

  useEffect(() => {
    loadModerationLogs(false);
    const intervalId = setInterval(() => {
      loadModerationLogs(true);
    }, 5000);

    return () => clearInterval(intervalId);
  }, []);

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
              <div className="mb-6 w-full flex flex-col lg:flex-row items-start lg:items-center justify-between gap-4">
                <div className="flex items-center space-x-3">
                  <div className="w-12 h-12 bg-libr-accent2/20 rounded-2xl flex items-center justify-center">
                    <Shield className="w-6 h-6 text-libr-accent2" />
                  </div>
                  <div>
                    <h1 className="text-xl font-bold text-foreground">Moderation Logs</h1>
                    <p className="text-xs text-muted-foreground">
                      Automatically recorded logs
                    </p>
                  </div>
                </div>

                {/* Controls */}
                <div className="flex flex-row gap-3 items-center">
                  <div className="relative w-64">
                    <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 w-4 h-4 text-muted-foreground" />
                    <input
                      type="text"
                      placeholder="Search logs..."
                      value={searchTerm}
                      onChange={(e) => setSearchTerm(e.target.value)}
                      className="w-full pl-9 pr-4 h-9 bg-muted/50 hover:bg-muted border-none rounded-2xl text-sm focus:outline-none focus:ring-1 focus:ring-libr-accent1/50 transition-all text-foreground"
                    />
                  </div>
                  <button
                    onClick={() => loadModerationLogs(false)}
                    disabled={isLoading}
                    className={`libr-button bg-libr-accent1/20 hover:bg-muted rounded-2xl flex items-center justify-center space-x-2 text-sm px-4 h-9 ${
                      isLoading ? 'opacity-80 cursor-not-allowed' : ''
                    }`}
                  >
                    <RotateCcw className={`w-4 h-4 ${isLoading ? 'animate-spin text-libr-accent1' : ''}`} />
                    <span className="mt-0.5">{isLoading ? 'Refreshing...' : 'Refresh'}</span>
                  </button>
                  <div className="flex items-center relative">
                    <Filter className="w-4 h-4 text-muted-foreground absolute left-3 pointer-events-none" />
                    <select
                      value={filter}
                      onChange={(e) => setFilter(e.target.value as any)}
                      className="libr-button appearance-none bg-libr-accent1/20 hover:bg-muted rounded-2xl text-sm focus:outline-none pl-9 pr-4 h-9 cursor-pointer transition-colors"
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
                <div className="space-y-2 mt-2">
                  {filteredLogs.map((log, index) => (
                    <motion.div
                      key={index}
                      initial={{ opacity: 0, y: 20 }}
                      animate={{ opacity: 1, y: 0 }}
                      transition={{ delay: index * 0.03 }}
                      className="flex w-full box-border"
                    >
                      <div className="w-full relative rounded-2xl px-6 py-4 bg-card hover:bg-muted/30 transition-colors border border-border/20 break-words overflow-hidden box-border">
                        <div className={`absolute left-0 top-0 bottom-0 w-[4px] ${
                          log.status === '1' ? 'bg-libr-accent1' : log.status === '0' ? 'bg-red-500' : 'bg-amber-400'
                        }`} />

                        <div className="flex w-full justify-between items-center gap-4">
                          <div className="flex-1 min-w-0">
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
                                    className="text-sm text-foreground leading-snug whitespace-pre-wrap [&_strong]:font-semibold [&_u]:underline [&_em]:italic [&_img]:w-48 [&_img]:h-auto [&_img]:rounded-lg [&_img]:mt-2 modlog-body-content"
                                    dangerouslySetInnerHTML={{ __html: cleanBody }}
                                  />
                                </>
                              );
                            })()}
                          </div>

                          <div className="flex items-center gap-3 flex-shrink-0 mt-0.5">
                            <div className="flex items-center text-[10px] opacity-45 text-muted-foreground space-x-1.5 font-mono">
                              <Clock className="w-3 h-3" />
                              <span>{new Date(log.timestamp * 1000).toLocaleString()}</span>
                            </div>
                            <div>
                              {log.status === '1' ? (
                                <span className="px-2 py-0.5 text-[10px] uppercase tracking-wider font-bold rounded-full border border-current text-teal-400">
                                  Approved
                                </span>
                              ) : log.status === '0' ? (
                                <span className="px-2 py-0.5 text-[10px] uppercase tracking-wider font-bold rounded-full border border-current text-rose-500">
                                  Rejected
                                </span>
                              ) : (
                                <span className="px-2 py-0.5 text-[10px] uppercase tracking-wider font-bold rounded-full border border-current text-amber-400">
                                  Pending
                                </span>
                              )}
                            </div>
                          </div>
                        </div>
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