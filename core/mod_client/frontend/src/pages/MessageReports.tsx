import React, { useEffect, useState } from 'react';
import { motion } from 'framer-motion';
import { useAppStore } from '../store/useAppStore';
import { apiService } from '../services/api';
import { Message } from '../store/useAppStore';
import { Shield, Check, X, Filter, Search, Clock, MessageSquare, FileWarning, MessageSquareWarning, AlertCircle, AlertTriangle } from 'lucide-react';
import { Sidebar } from '@/components/layout/Sidebar';
export const MsgReports: React.FC = () => {
  const { user } = useAppStore();
  const [reports, setReports] = useState<Message[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [filter, setFilter] = useState<'all' | 'pending' | 'approved' | 'rejected'>('all');
  const [searchTerm, setSearchTerm] = useState('');

  // useEffect(() => {
  //   loadReports();
  // }, []);

  const loadReports = async () => {
    setIsLoading(true);
    try {
      const moderationReports = await apiService.getMessages("1");
      setReports(moderationReports);
    } catch (error) {
      console.error('Failed to load moderation reports:', error);
    } finally {
      setIsLoading(false);
    }
  };

  const handleModerate = async (messageId: string, action: 'approve' | 'reject', note?: string) => {
    try {
      await apiService.moderateMessage(messageId, action, note);
      setReports(reports.map(report => 
        report.id === messageId 
          ? { ...report, status: action === 'approve' ? 'approved' : 'rejected', moderationNote: note }
          : report
      ));
    } catch (error) {
      console.error('Failed to moderate message:', error);
    }
  };

  const filteredReports = reports.filter(report => {
    const matchesFilter = filter === 'all' || report.status === filter;
    const matchesSearch = searchTerm === '' || 
      report.content.toLowerCase().includes(searchTerm.toLowerCase()) ||
      report.authorAlias.toLowerCase().includes(searchTerm.toLowerCase());
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
    <div className="flex-1 flex flex-col bg-libr-primary h-screen">
      <div className="flex-1 overflow-y-auto">
        <div className="p-6 pb-24">
          <motion.div
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            className="max-w-6xl mx-auto"
          >
            {/* (HTML untouched) */}
            {/* It will still show “Moderation Logs”, but logic uses reports */}
            
            {isLoading ? (
              <div className="flex items-center justify-center py-12">
                <motion.div
                  animate={{ rotate: 360 }}
                  transition={{ duration: 1, repeat: Infinity, ease: "linear" }}
                  className="w-8 h-8 border-2 border-libr-accent2 border-t-transparent rounded-full"
                />
              </div>
            ) : filteredReports.length === 0 ? (
              <motion.div
                initial={{ opacity: 0, y: 20 }}
                animate={{ opacity: 1, y: 0 }}
                className="text-center py-12"
              >
                <div className="w-16 h-16 bg-libr-accent2/20 rounded-2xl flex items-center justify-center mx-auto mb-4">
                  <AlertTriangle className="w-8 h-8 text-libr-accent2" />
                </div>
                <h3 className="text-lg font-medium text-foreground mb-2">
                  No reports found
                </h3>
                <p className="text-muted-foreground">
                  {filter === 'all' ? 'No message reports available' : `No ${filter} messages found`}
                </p>
              </motion.div>
            ) : (
              <div className="grid gap-4">
                {filteredReports.map((report, index) => (
                  <motion.div
                    key={report.id}
                    initial={{ opacity: 0, y: 20 }}
                    animate={{ opacity: 1, y: 0 }}
                    transition={{ delay: index * 0.05 }}
                    className="libr-card p-6"
                  >
                    <div className="flex items-start justify-between">
                      <div className="flex-1">
                        {/* JSX untouched */}
                      </div>
                      {report.status === 'pending' && (
                        <div className="flex items-center space-x-2 ml-4">
                          <motion.button
                            whileHover={{ scale: 1.05 }}
                            whileTap={{ scale: 0.95 }}
                            onClick={() => handleModerate(report.id, 'approve')}
                            className="libr-button bg-green-500 hover:bg-green-600 text-white flex items-center space-x-1"
                          >
                            <Check className="w-4 h-4" />
                            <span>Approve</span>
                          </motion.button>
                          <motion.button
                            whileHover={{ scale: 1.05 }}
                            whileTap={{ scale: 0.95 }}
                            onClick={() => handleModerate(report.id, 'reject', 'Inappropriate content')}
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
    </div>
  );
};
