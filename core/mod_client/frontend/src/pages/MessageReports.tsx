import React, { useEffect, useState } from 'react';
import { motion } from 'framer-motion';
import { ReportedMessage, useAppStore } from '../store/useAppStore';
import { apiService } from '../services/api';
import { Message } from '../store/useAppStore';
import { Shield, Check, X, Filter, Search, Clock, MessageSquare, FileWarning, MessageSquareWarning, AlertCircle, AlertTriangle } from 'lucide-react';
import { Sidebar } from '@/components/layout/Sidebar';
export const MsgReports: React.FC = () => {
  const { user } = useAppStore();
  const [reports, setReports] = useState<ReportedMessage[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [filter, setFilter] = useState<'all' | 'pending' | 'approved' | 'rejected'>('all');
  const [searchTerm, setSearchTerm] = useState('');

  // useEffect(() => {
  //   loadReports();
  // }, []);

  const loadReports = async () => {
    setIsLoading(true);
    try {
      const moderationReports = await apiService.getMessageReports("1");
      setReports(moderationReports);
    } catch (error) {
      console.error('Failed to load moderation reports:', error);
    } finally {
      setIsLoading(false);
    }
  };

  const handleModerate = async (messageId: string, action: 'approve' | 'reject', note?: string) => {
    try {
      // Find the report to get its sign
      const report = reports.find(
        r =>
          r.authorPublicKey === messageId.split('_')[0] &&
          r.timestamp.toString() === messageId.split('_')[1]
      );
      if (!report) return;

      const { MsgCert } = require('../../wailsjs/go/models'); // Adjust path as needed
      const cert = new MsgCert({
        public_key: report.authorPublicKey,
        msg: {
          content: report.content,
          ts: Number(report.timestamp),
        },
        mod_certs: report.moderationNote || [],
        sign: report.sign,
      });

      await apiService.manualModerate(cert, action === 'approve' ? 1 : 0);

      setReports(reports.map(r =>
        r.authorPublicKey === messageId.split('_')[0] &&
        r.timestamp.toString() === messageId.split('_')[1]
          ? {
              ...r,
              status: action === 'approve' ? 'approved' : 'rejected',
              note: note || "",
            }
          : r
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
                {filteredReports.map((report, index) => {
                  const uniqueKey = `${report.authorPublicKey}_${report.timestamp.toString()}`;
                  return (
                    <motion.div
                      key={uniqueKey}
                      initial={{ opacity: 0, y: 20 }}
                      animate={{ opacity: 1, y: 0 }}
                      transition={{ delay: index * 0.05 }}
                      className="libr-card p-6"
                    >
                      <div className="flex flex-col gap-2">
                        <div>
                          <span className="font-semibold text-libr-secondary">Message:</span>
                          <span className="ml-2 text-foreground">{report.content}</span>
                        </div>
                        <div>
                          <span className="font-semibold text-libr-secondary">Timestamp:</span>
                          <span className="ml-2 text-foreground">{new Date(Number(report.timestamp)).toLocaleString()}</span>
                        </div>
                        <div>
                          <span className="font-semibold text-libr-secondary">Reason:</span>
                          <span className="ml-2 text-foreground">{report.note || "—"}</span>
                        </div>
                        <div className="flex items-center gap-2 mt-2">
                          {report.status === 'pending' && (
                            <>
                              <motion.button
                                whileHover={{ scale: 1.05 }}
                                whileTap={{ scale: 0.95 }}
                                onClick={() => handleModerate(uniqueKey, 'approve')}
                                className="libr-button bg-green-500 hover:bg-green-600 text-white flex items-center space-x-1"
                              >
                                <Check className="w-4 h-4" />
                                <span>Approve</span>
                              </motion.button>
                              <motion.button
                                whileHover={{ scale: 1.05 }}
                                whileTap={{ scale: 0.95 }}
                                onClick={() => handleModerate(uniqueKey, 'reject', 'Inappropriate content')}
                                className="libr-button bg-red-500 hover:bg-red-600 text-white flex items-center space-x-1"
                              >
                                <X className="w-4 h-4" />
                                <span>Reject</span>
                              </motion.button>
                            </>
                          )}
                          {report.status === 'approved' && (
                            <span className="text-green-600 font-semibold flex items-center gap-1">
                              <Check className="w-4 h-4" /> Approved
                            </span>
                          )}
                          {report.status === 'rejected' && (
                            <span className="text-red-600 font-semibold flex items-center gap-1">
                              <X className="w-4 h-4" /> Rejected
                            </span>
                          )}
                        </div>
                      </div>
                    </motion.div>
                  );
                })}
              </div>
            )}
          </motion.div>
        </div>
      </div>
    </div>
    </div>
  );
};
