import React, { useEffect, useState } from 'react';
import { motion } from 'framer-motion';
import { ReportedMessage, useAppStore } from '../store/useAppStore';
import { apiService } from '../services/api';
import { Sidebar } from '@/components/layout/Sidebar';
import { Check, X, AlertTriangle, MessageSquare } from 'lucide-react';

export const MsgReports: React.FC = () => {
  const { user } = useAppStore();
  const [reports, setReports] = useState<ReportedMessage[]>([]);
  const [isLoading, setIsLoading] = useState(false);

  useEffect(() => {
    loadReports();
  }, []);

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

  const handleModerate = async (sign: string, action: 'approve' | 'reject', note?: string) => {
    try {
      const report = reports.find(r => r.sign === sign);
      if (!report) return;

      const { MsgCert } = require('../../wailsjs/go/models');
      const cert = new MsgCert({
        public_key: '',
        msg: {
          content: report.content,
          ts: 0,
        },
        mod_certs: [],
        sign: report.sign,
        reason: report.note || '',
      });

      await apiService.manualModerate(cert, action === 'approve' ? 0 : 1);
      setReports(reports.filter(r => r.sign !== sign));
    } catch (error) {
      console.error('Failed to moderate message:', error);
    }
  };

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
                    <MessageSquare className="w-6 h-6 text-libr-accent2" />
                  </div>
                  <div>
                    <h1 className="text-2xl font-bold text-foreground">Message Reports</h1>
                    <p className="text-muted-foreground">
                      These messages have been reported by users and require moderation.
                    </p>
                  </div>
                </div>
                <button
                  onClick={loadReports}
                  className="px-3 py-2 bg-muted/30 border border-border/50 text-base rounded-lg text-foreground hover:bg-muted/40 transition-colors"
                >
                  Refresh
                </button>
              </div>

              {/* Reports */}
              {isLoading ? (
                <div className="flex items-center justify-center py-12">
                  <motion.div
                    animate={{ rotate: 360 }}
                    transition={{ duration: 1, repeat: Infinity, ease: "linear" }}
                    className="w-8 h-8 border-2 border-libr-accent2 border-t-transparent rounded-full"
                  />
                </div>
              ) : reports.length === 0 ? (
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
                    No message reports available
                  </p>
                </motion.div>
              ) : (
                <div className="grid gap-3">
                  {reports.map((report, index) => (
                    <motion.div
                      key={report.sign}
                      initial={{ opacity: 0, y: 20 }}
                      animate={{ opacity: 1, y: 0 }}
                      transition={{ delay: index * 0.03 }}
                      className="bg-muted/30 border border-border/40 rounded-xl space-y-4 p-4"
                    >
                      <div className="mt-3 p-3 rounded-xl">
                        <div className="text-sm text-foreground leading-snug whitespace-pre-wrap">
                          <span className="font-semibold text-libr-secondary">Message:</span>
                          <span className="ml-2">{report.content}</span>
                        </div>
                        <div className="text-sm text-foreground leading-snug whitespace-pre-wrap mt-2">
                          <span className="font-semibold text-libr-secondary">Reason:</span>
                          <span className="ml-2">{report.note || "—"}</span>
                        </div>
                      </div>
                      <div className="flex items-center justify-end gap-2">
                        <motion.button
                          whileHover={{ scale: 1.05 }}
                          whileTap={{ scale: 0.95 }}
                          onClick={() => handleModerate(report.sign, 'approve')}
                          className="libr-button bg-green-500 hover:bg-green-600 text-white flex items-center space-x-1"
                        >
                          <Check className="w-4 h-4" />
                          <span>Approve</span>
                        </motion.button>
                        <motion.button
                          whileHover={{ scale: 1.05 }}
                          whileTap={{ scale: 0.95 }}
                          onClick={() => handleModerate(report.sign, 'reject', 'Inappropriate content')}
                          className="libr-button bg-red-500 hover:bg-red-600 text-white flex items-center space-x-1"
                        >
                          <X className="w-4 h-4" />
                          <span>Reject</span>
                        </motion.button>
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
