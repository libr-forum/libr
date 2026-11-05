import React, { useEffect, useState } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import { ReportedMessage, useAppStore } from '../store/useAppStore';
import { apiService } from '../services/api';
import { Sidebar } from '@/components/layout/Sidebar';
import { Check, X, AlertTriangle, MessageSquare, ImageIcon, Flag, Trash2, ShieldCheck, RotateCcw } from 'lucide-react';
import DOMPurify from 'dompurify';

// Parse <HEAD>title</HEAD><BODY>body</BODY> format
function parseContent(raw: string): { title: string; body: string } {
  const headMatch = raw.match(/<HEAD>([\s\S]*?)<\/HEAD>/i);
  const bodyMatch = raw.match(/<BODY>([\s\S]*?)<\/BODY>/i);
  return {
    title: headMatch?.[1] ?? '',
    body: bodyMatch?.[1] ?? raw,
  };
}

// Extract image srcs from HTML, returning sanitized text HTML and src list
function extractImages(html: string): { textHtml: string; imgSrcs: string[] } {
  const safe = DOMPurify.sanitize(html, {
    ADD_TAGS: ['img'],
    ADD_ATTR: ['src', 'alt', 'class', 'style'],
    ALLOW_DATA_ATTR: true,
    FORCE_BODY: true,
  });
  const div = document.createElement('div');
  div.innerHTML = safe;
  const srcs: string[] = [];
  div.querySelectorAll('img').forEach(img => {
    srcs.push(img.src);
    img.remove();
  });
  return { textHtml: div.innerHTML, imgSrcs: srcs };
}

export const MsgReports: React.FC = () => {
  const { user } = useAppStore();
  const [reports, setReports] = useState<ReportedMessage[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [filter, setFilter] = useState<'all' | 'reports' | 'manual_mod'>('all');
  const [lightboxSrc, setLightboxSrc] = useState<string | null>(null);

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

  const handleModerate = async (sign: string, value: number) => {
    try {
      // value: 1 = proceed (allow image / remove reported message)
      // value: 0 = halt (block image / keep reported message)
      await apiService.moderateBySign(sign, value);
      setReports(prev => prev.filter(r => r.sign !== sign));
    } catch (error) {
      console.error('Failed to moderate message:', error);
    }
  };

  if (user?.role !== 'moderator') {
    return (
      <div className='flex flex-row'>
        <div className='w-[19.4%]'><Sidebar /></div>
        <div className="flex-1 flex items-center justify-center bg-libr-primary">
          <motion.div initial={{ opacity: 0, scale: 0.9 }} animate={{ opacity: 1, scale: 1 }} className="text-center">
            <div className="w-20 h-20 bg-red-500/20 rounded-2xl flex items-center justify-center mx-auto mb-4">
              <X className="w-10 h-10 text-red-500" />
            </div>
            <h2 className="text-xl font-semibold text-foreground mb-2">Access Denied</h2>
            <p className="text-muted-foreground">You need moderator privileges to access this page</p>
          </motion.div>
        </div>
      </div>
    );
  }

  const reportCount = reports.filter(r => r.type !== 'manual_mod').length;
  const imageCount = reports.filter(r => r.type === 'manual_mod').length;

  const filteredReports = reports.filter(r => {
    if (filter === 'reports') return r.type !== 'manual_mod';
    if (filter === 'manual_mod') return r.type === 'manual_mod';
    return true;
  });

  return (
    <>
      <div className='flex flex-row'>
        <div className='w-[19.4%]'><Sidebar /></div>
        <div className="flex-1 flex flex-col w-full bg-libr-primary h-screen">
          <div className="flex-1 overflow-y-auto">
            <div className="pt-6 pb-24">
              <motion.div initial={{ opacity: 0, y: 20 }} animate={{ opacity: 1, y: 0 }} className="max-w-7xl mx-4">

                {/* Header */}
                <div className="mb-6 w-full flex items-center justify-between">
                  <div className="flex items-center space-x-3">
                    <div className="w-12 h-12 bg-libr-accent2/20 rounded-xl flex items-center justify-center">
                      <MessageSquare className="w-6 h-6 text-libr-accent2" />
                    </div>
                    <div>
                      <h1 className="text-2xl font-bold text-foreground">Moderation Queue</h1>
                      <p className="text-muted-foreground">Review incoming images and user-reported live messages.</p>
                    </div>
                  </div>
                  <button
                    onClick={loadReports}
                    disabled={isLoading}
                    className={`px-3 py-2 bg-muted/30 border border-border/50 text-sm rounded-lg text-foreground hover:bg-muted/40 transition-colors flex items-center gap-2 ${
                      isLoading ? 'opacity-80 cursor-not-allowed' : ''
                    }`}
                  >
                    <RotateCcw className={`w-4 h-4 ${isLoading ? 'animate-spin text-libr-accent2' : ''}`} />
                    <span>{isLoading ? 'Refreshing...' : 'Refresh'}</span>
                  </button>
                </div>

                {/* Filter Tabs */}
                <div className="flex items-center gap-2 mb-6">
                  <button
                    onClick={() => setFilter('all')}
                    className={`px-3.5 py-1.5 rounded-xl text-sm font-medium transition ${
                      filter === 'all'
                        ? 'bg-libr-accent2 text-white shadow-sm'
                        : 'bg-muted/30 text-muted-foreground hover:bg-muted/50'
                    }`}
                  >
                    All ({reports.length})
                  </button>
                  <button
                    onClick={() => setFilter('reports')}
                    className={`px-3.5 py-1.5 rounded-xl text-sm font-medium transition ${
                      filter === 'reports'
                        ? 'bg-rose-500 text-white shadow-sm'
                        : 'bg-muted/30 text-muted-foreground hover:bg-muted/50'
                    }`}
                  >
                    Reports ({reportCount})
                  </button>
                  <button
                    onClick={() => setFilter('manual_mod')}
                    className={`px-3.5 py-1.5 rounded-xl text-sm font-medium transition ${
                      filter === 'manual_mod'
                        ? 'bg-sky-500 text-white shadow-sm'
                        : 'bg-muted/30 text-muted-foreground hover:bg-muted/50'
                    }`}
                  >
                    Image Reviews ({imageCount})
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
                ) : filteredReports.length === 0 ? (
                  <motion.div initial={{ opacity: 0, y: 20 }} animate={{ opacity: 1, y: 0 }} className="text-center py-12">
                    <div className="w-16 h-16 bg-libr-accent2/20 rounded-2xl flex items-center justify-center mx-auto mb-4">
                      <AlertTriangle className="w-8 h-8 text-libr-accent2" />
                    </div>
                    <h3 className="text-lg font-medium text-foreground mb-2">Queue is clear</h3>
                    <p className="text-muted-foreground">No pending items to review</p>
                  </motion.div>
                ) : (
                  <div className="grid gap-4">
                    {filteredReports.map((report, index) => {
                      const { title, body } = parseContent(report.content);
                      const { textHtml, imgSrcs } = extractImages(body);
                      const isManualMod = report.type === 'manual_mod';

                      return (
                        <motion.div
                          key={report.sign}
                          initial={{ opacity: 0, y: 20 }}
                          animate={{ opacity: 1, y: 0 }}
                          transition={{ delay: index * 0.03 }}
                          className="bg-card border border-border/20 rounded-3xl space-y-4 p-5 shadow-md"
                        >
                          <div className="rounded-2xl bg-background/30 border border-border/20 p-4">
                            {/* Type Badge */}
                            <div className="flex items-center justify-between mb-3">
                              {isManualMod ? (
                                <span className="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-full text-xs font-semibold bg-sky-500/20 text-sky-400 border border-sky-500/30">
                                  <ImageIcon className="w-3.5 h-3.5" /> Image Review (Pre-publish)
                                </span>
                              ) : (
                                <span className="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-full text-xs font-semibold bg-rose-500/20 text-rose-400 border border-rose-500/30">
                                  <Flag className="w-3.5 h-3.5" /> User Report (Live Message)
                                </span>
                              )}
                            </div>

                            {/* Title */}
                            {title && (
                              <div className="text-sm font-semibold text-foreground mb-1">{title}</div>
                            )}

                            {/* Text body */}
                            <div className="text-sm text-foreground leading-snug">
                              <span className="font-semibold text-libr-secondary">Message: </span>
                              <span
                                className="leading-relaxed whitespace-pre-wrap break-words"
                                dangerouslySetInnerHTML={{ __html: textHtml }}
                              />
                            </div>

                            {/* Image gallery */}
                            {imgSrcs.length > 0 && (
                              <div className="mt-3 flex flex-wrap gap-2">
                                {imgSrcs.map((src, i) => (
                                  <div key={i} className="relative group cursor-pointer" onClick={() => setLightboxSrc(src)}>
                                    <img
                                      src={src}
                                      alt={`attachment-${i + 1}`}
                                      className="max-h-48 max-w-xs rounded-2xl object-cover border border-border/40 shadow-md group-hover:opacity-90 transition-opacity"
                                    />
                                    <div className="absolute inset-0 flex items-center justify-center opacity-0 group-hover:opacity-100 transition-opacity">
                                      <div className="bg-black/50 rounded-full p-2">
                                        <ImageIcon className="w-4 h-4 text-white" />
                                      </div>
                                    </div>
                                  </div>
                                ))}
                              </div>
                            )}

                            {/* Reason */}
                            <div className="text-sm text-foreground leading-snug whitespace-pre-wrap mt-2">
                              <span className="font-semibold text-libr-secondary">
                                {isManualMod ? 'Trigger:' : 'Reason:'}
                              </span>
                              <span className="ml-2">{report.note || (isManualMod ? 'Image attached' : '—')}</span>
                            </div>
                          </div>

                          {/* Action Buttons: 1-word clear labels */}
                          {isManualMod ? (
                            <div className="flex items-center justify-end gap-2">
                              <motion.button
                                whileHover={{ scale: 1.05 }} whileTap={{ scale: 0.95 }}
                                onClick={() => handleModerate(report.sign, 1)}
                                className="libr-button bg-green-500 hover:bg-green-600 text-white flex items-center space-x-1"
                              >
                                <Check className="w-4 h-4" /><span>Allow</span>
                              </motion.button>
                              <motion.button
                                whileHover={{ scale: 1.05 }} whileTap={{ scale: 0.95 }}
                                onClick={() => handleModerate(report.sign, 0)}
                                className="libr-button bg-red-500 hover:bg-red-600 text-white flex items-center space-x-1"
                              >
                                <X className="w-4 h-4" /><span>Block</span>
                              </motion.button>
                            </div>
                          ) : (
                            <div className="flex items-center justify-end gap-2">
                              <motion.button
                                whileHover={{ scale: 1.05 }} whileTap={{ scale: 0.95 }}
                                onClick={() => handleModerate(report.sign, 0)}
                                className="libr-button bg-muted/60 hover:bg-muted text-foreground border border-border/50 flex items-center space-x-1"
                              >
                                <ShieldCheck className="w-4 h-4 text-emerald-400" /><span>Keep</span>
                              </motion.button>
                              <motion.button
                                whileHover={{ scale: 1.05 }} whileTap={{ scale: 0.95 }}
                                onClick={() => handleModerate(report.sign, 1)}
                                className="libr-button bg-red-500 hover:bg-red-600 text-white flex items-center space-x-1"
                              >
                                <Trash2 className="w-4 h-4" /><span>Remove</span>
                              </motion.button>
                            </div>
                          )}
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

      {/* Lightbox */}
      <AnimatePresence>
        {lightboxSrc && (
          <motion.div
            initial={{ opacity: 0 }} animate={{ opacity: 1 }} exit={{ opacity: 0 }}
            className="fixed inset-0 z-[100] flex items-center justify-center bg-black/80 backdrop-blur-sm"
            onClick={() => setLightboxSrc(null)}
          >
            <motion.div
              initial={{ scale: 0.85, opacity: 0 }} animate={{ scale: 1, opacity: 1 }} exit={{ scale: 0.85, opacity: 0 }}
              transition={{ type: 'spring', stiffness: 300, damping: 25 }}
              className="relative max-w-[90vw] max-h-[90vh]"
              onClick={e => e.stopPropagation()}
            >
              <img src={lightboxSrc} alt="Full size" className="max-w-[90vw] max-h-[90vh] rounded-2xl shadow-2xl object-contain" />
              <button
                onClick={() => setLightboxSrc(null)}
                className="absolute -top-3 -right-3 bg-card border border-border rounded-full p-1 shadow-lg hover:bg-muted transition"
              >
                <X className="w-4 h-4" />
              </button>
            </motion.div>
          </motion.div>
        )}
      </AnimatePresence>
    </>
  );
};
