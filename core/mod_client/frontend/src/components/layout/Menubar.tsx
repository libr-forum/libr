import React from 'react';
import { BrowserOpenURL } from '../../../wailsjs/runtime';
import { PencilLine, Globe, Database, Copyright, ChevronDown, ChevronRight, ImageIcon, FileText, Flag, CheckCircle2, XCircle, X } from 'lucide-react';
import { logger } from '../../logger/logger';
import { EventsOn } from "../../../wailsjs/runtime";
import { toast } from "sonner";
import {
  GetOnlineMods,
  GenerateAlias,
  GenerateAvatar,
  GetPendingModerationStats
} from "../../../wailsjs/go/main/App";
import { useAppStore, PendingModeration } from '../../store/useAppStore';

type ModDisplay = {
  key: string;
  alias: string;
  avatarSvg: string;
};

type ResolvedItem = PendingModeration & { finalStatus: 'approved' | 'rejected' };
type DisplayQueueItem =
  | { kind: 'pending'; item: PendingModeration }
  | { kind: 'resolved'; item: ResolvedItem };

const ComingSoonDialog: React.FC<{ open: boolean; onClose: () => void }> = ({ open, onClose }) => {
  if (!open) return null;
  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40">
      <div className="bg-card border border-border/50 rounded-2xl shadow-xl text-foreground p-6 w-[90%] max-w-md flex flex-col ">
        <span className="text-lg font-semibold mb-4 text-libr-secondary">Feature Coming Soon</span>
        <p className="text-muted-foreground mb-6 text-left">
          This feature is not available yet. Stay tuned for updates!
        </p>
        <div className="flex justify-end space-x-2">
          <button
            onClick={onClose}
            className="libr-button bg-muted hover:bg-muted/70 text-foreground px-6 py-2"
          >
            Close
          </button>
        </div>
      </div>
    </div>
  );
};

function getQueueText(item: PendingModeration | ResolvedItem) {
  const isImage = item.reason === 'Image attached';
  const isReport = !isImage && !!item.reason;

  const rawText = item.content
    ? item.content
        .replace(/<HEAD>[\s\S]*?<\/HEAD>/gi, '')
        .replace(/<BODY>([\s\S]*?)<\/BODY>/gi, '$1')
        .replace(/<[^>]+>/g, '')
        .trim()
    : '';

  const preview = rawText.length > 0
    ? rawText.substring(0, 40) + (rawText.length > 40 ? '...' : '')
    : null;

  return { rawText, preview, isImage, isReport };
}

export const Menubar: React.FC = () => {
  const [mods, setMods] = React.useState<ModDisplay[]>([]);
  const [dialogOpen, setDialogOpen] = React.useState(false);
  const [queueOpen, setQueueOpen] = React.useState(true);
  const [expandedItem, setExpandedItem] = React.useState<string | null>(null);
  const [resolvedItems, setResolvedItems] = React.useState<ResolvedItem[]>([]);
  const [displayOrder, setDisplayOrder] = React.useState<string[]>([]);
  const { pendingQueue, setPendingQueue } = useAppStore();

  const pendingQueueRef = React.useRef<PendingModeration[]>(pendingQueue);
  React.useEffect(() => { pendingQueueRef.current = pendingQueue; }, [pendingQueue]);

  React.useEffect(() => {
    const activeIds = new Set([
      ...pendingQueue.map(item => item.id),
      ...resolvedItems.map(item => item.id),
    ]);

    setDisplayOrder(prev => {
      const next = prev.filter(id => activeIds.has(id));

      pendingQueue.forEach(item => {
        if (!next.includes(item.id)) {
          next.push(item.id);
        }
      });

      resolvedItems.forEach(item => {
        if (!next.includes(item.id)) {
          next.push(item.id);
        }
      });

      return next;
    });
  }, [pendingQueue, resolvedItems]);

  const pendingMap = React.useMemo(
    () => new Map(pendingQueue.map(item => [item.id, item])),
    [pendingQueue]
  );

  const resolvedMap = React.useMemo(
    () => new Map(resolvedItems.map(item => [item.id, item])),
    [resolvedItems]
  );

  const displayItems = React.useMemo<DisplayQueueItem[]>(
    () =>
      displayOrder.flatMap((id) => {
        const resolved = resolvedMap.get(id);
        if (resolved) {
          return [{ kind: 'resolved', item: resolved }];
        }

        const pending = pendingMap.get(id);
        if (pending) {
          return [{ kind: 'pending', item: pending }];
        }

        return [];
      }),
    [displayOrder, pendingMap, resolvedMap]
  );

  React.useEffect(() => {
    GetPendingModerationStats().then((stats) => {
      if (!stats?.items) return;
      const queue: PendingModeration[] = stats.items.map((item: any) => ({
        id: item.msg_sign,
        ts: item.ts ?? 0,
        content: item.content ?? '',
        reason: item.reason ?? (item.is_image ? 'Image attached' : ''),
        approved: item.approved,
        rejected: item.rejected,
        totalMods: item.approved + item.rejected + item.awaiting,
        ackCount: item.approved + item.rejected,
        awaitingMods: item.awaiting,
      }));
      setPendingQueue(queue);
    }).catch(() => {});

    const unsubCronStatus = EventsOn("cron_status_update", (queue: PendingModeration[]) => {
      setPendingQueue(queue || []);
    });

    const unsubFinalized = EventsOn("moderation_finalized", (event: { status: string; id: string }) => {
      const finalStatus = event.status === "approved" ? 'approved' : 'rejected';

      if (finalStatus === 'approved') {
        toast.success(`Message Approved (${event.id.substring(0, 8)}...)`);
      } else {
        toast.error(`Message Rejected (${event.id.substring(0, 8)}...)`);
      }

      const source = pendingQueueRef.current.find(q => q.id === event.id);
      const resolved: ResolvedItem = source
        ? { ...source, finalStatus }
        : { id: event.id, ts: 0, content: '', reason: '', approved: 0, rejected: 0, totalMods: 0, ackCount: 0, awaitingMods: 0, finalStatus };

      setResolvedItems(prev => [...prev.filter(r => r.id !== event.id), resolved]);

      setTimeout(() => {
        setResolvedItems(prev => prev.filter(r => r.id !== event.id));
      }, 30_000);
    });

    return () => {
      unsubCronStatus();
      unsubFinalized();
    };
  }, [setPendingQueue]);

  React.useEffect(() => {
    logger.debug('[Menubar] Component mounted.');
    async function fetchMods() {
      try {
        const keys = await GetOnlineMods();
        logger.debug('[Menubar] Received keys:', keys);
        const resolved = await Promise.all(
          keys.map(async (key) => {
            const alias = await GenerateAlias(key);
            const avatarSvg = await GenerateAvatar(key);
            logger.debug(`[Menubar] Processed mod: ${alias}`);
            return { key, alias, avatarSvg };
          })
        );
        setMods(resolved);
        logger.info('[Menubar] Mods loaded successfully.');
      } catch (err) {
        logger.error('[Menubar] Failed to load online mods:', err);
      }
    }

    fetchMods();
    return () => logger.debug('[Menubar] Component unmounted.');
  }, []);

  return (
    <div className="w-full p-1 bg-transparent items-center rounded-3xl h-full flex flex-col z-50">
      <ComingSoonDialog open={dialogOpen} onClose={() => {
        logger.info('[ComingSoonDialog] Closed.');
        setDialogOpen(false);
      }} />
      <div className="flex-1 overflow-y-auto flex flex-col w-full items-center gap-4 px-1 pt-2">
        <div className="w-full rounded-3xl bg-card border border-border/50 px-4 py-4">
          <div className="text-left w-full mb-4 flex items-center">
            <h3 className="text-sm font-semibold text-muted-foreground">
              Moderators
            </h3>
          </div>

          <div className="flex flex-col gap-3 w-full">
            {mods.map(({ key, alias, avatarSvg }) => (
              <div key={key} className="flex items-center justify-start space-x-3 py-1">
                {avatarSvg && avatarSvg !== "unknown" ? (
                  <img
                    src={`data:image/svg+xml;base64,${avatarSvg}`}
                    alt="avatar"
                    className="w-10 h-10 rounded-xl"
                  />
                ) : (
                  <div className="w-10 h-10 bg-libr-accent1 rounded-full flex items-center justify-center">
                    <span className="text-white text-sm font-medium">
                      {alias.charAt(0).toUpperCase()}
                    </span>
                  </div>
                )}
                <span className="text-sm font-medium text-foreground">{alias}</span>
              </div>
            ))}
          </div>
        </div>

        <div className="w-full rounded-3xl bg-card border border-border/50 px-4 py-4">
          <button
            onClick={() => setQueueOpen(prev => !prev)}
            className="text-left w-full mb-2 flex items-center hover:opacity-80 transition-opacity"
          >
            <div className="text-sm font-semibold text-muted-foreground flex items-center justify-between w-full">
              <span className="flex items-center gap-1.5">
                <ChevronDown
                  className={`w-3.5 h-3.5 transition-transform duration-200 ${queueOpen ? '' : '-rotate-90'}`}
                />
                Moderation Queue
              </span>
              {displayItems.length > 0 && (
                <span className="bg-libr-accent1 text-white text-xs px-2 py-0.5 rounded-full">
                  {displayItems.length}
                </span>
              )}
            </div>
          </button>

          {queueOpen && (
            <div className="flex flex-col gap-2 w-full">
              {displayItems.length === 0 ? (
                <p className="text-xs text-muted-foreground pl-1">No items pending.</p>
              ) : (
                <>
                  {displayItems.map((entry) => {
                    const item = entry.item;
                    const isResolved = entry.kind === 'resolved';
                    const isApproved = isResolved && item.finalStatus === 'approved';
                    const isOpen = expandedItem === item.id;
                    const { rawText, preview, isImage, isReport } = getQueueText(item);
                    const TypeIcon = isImage ? ImageIcon : isReport ? Flag : FileText;
                    const StatusIcon = isApproved ? CheckCircle2 : XCircle;
                    const statusColor = isApproved ? 'text-green-400' : 'text-red-400';

                    return (
                      <div
                        key={item.id}
                        className={
                          isResolved
                            ? `rounded-lg border ${isApproved ? 'border-green-500/40 bg-green-500/5' : 'border-red-500/40 bg-red-500/5'} overflow-hidden`
                            : "bg-muted/30 rounded-lg border border-border/50 overflow-hidden cursor-pointer hover:border-libr-accent1/40 transition-colors"
                        }
                      >
                        <div
                          className="flex items-center justify-between p-3 text-xs gap-2 cursor-pointer"
                          onClick={() => setExpandedItem(isOpen ? null : item.id)}
                        >
                          <div className="flex items-center gap-1.5 min-w-0">
                            <ChevronRight
                              className={`w-3 h-3 text-muted-foreground flex-shrink-0 transition-transform duration-150 ${isOpen ? 'rotate-90' : ''}`}
                            />
                            {isResolved ? (
                              <StatusIcon className={`w-3.5 h-3.5 ${statusColor} flex-shrink-0`} />
                            ) : (
                              <TypeIcon className="w-3 h-3 text-libr-accent1 flex-shrink-0" />
                            )}
                            <span className={`font-semibold truncate ${isResolved ? statusColor : 'text-libr-accent1'}`}>
                              {preview ?? (isImage ? 'Image message' : isReport ? item.reason : 'Pending message')}
                            </span>
                          </div>

                          <div className="flex items-center gap-1.5 flex-shrink-0">
                            {isResolved ? (
                              <>
                                <span className={`font-semibold uppercase text-[10px] tracking-wide ${statusColor}`}>
                                  {item.finalStatus}
                                </span>
                                <button
                                  onClick={(e) => {
                                    e.stopPropagation();
                                    setResolvedItems(prev => prev.filter(r => r.id !== item.id));
                                  }}
                                  className="text-muted-foreground/50 hover:text-muted-foreground transition-colors"
                                >
                                  <X className="w-3 h-3" />
                                </button>
                              </>
                            ) : (
                              <span className="text-muted-foreground/70 tabular-nums">
                                {item.ts > 0
                                  ? new Date(item.ts * 1000).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
                                  : '---'}
                              </span>
                            )}
                          </div>
                        </div>

                        {isOpen && (
                          <div className={`px-3 pb-3 pt-1 text-xs flex flex-col gap-2 ${isResolved ? 'border-t border-border/20' : 'border-t border-border/30'}`}>
                            {isImage ? (
                              <div className="flex items-center gap-1.5 text-muted-foreground italic">
                                <ImageIcon className="w-3 h-3" />
                                <span>Image attachment</span>
                              </div>
                            ) : rawText.length > 0 ? (
                              <div className={`leading-relaxed break-words rounded-md px-2.5 py-2 ${isResolved ? 'text-foreground/70 bg-background/30 border border-border/20' : 'text-foreground/80 bg-background/40 border border-border/30'}`}>
                                {rawText.substring(0, 200)}{rawText.length > 200 ? '...' : ''}
                              </div>
                            ) : null}

                            {isReport && (
                              <div className="flex items-center gap-1">
                                <Flag className="w-2.5 h-2.5 text-yellow-400" />
                                <span className="text-yellow-400/80">{item.reason}</span>
                              </div>
                            )}

                            {!isResolved && (
                              <div className="flex flex-col gap-1.5">
                                <div className="w-full h-1 overflow-hidden rounded-full bg-background/40">
                                  <div className="flex h-full w-full">
                                    <div
                                      className="bg-green-400/90"
                                      style={{ width: `${item.totalMods > 0 ? (item.approved / item.totalMods) * 100 : 0}%` }}
                                    />
                                    <div
                                      className="bg-red-400/90"
                                      style={{ width: `${item.totalMods > 0 ? (item.rejected / item.totalMods) * 100 : 0}%` }}
                                    />
                                    <div
                                      className="bg-yellow-400/90"
                                      style={{ width: `${item.totalMods > 0 ? (item.awaitingMods / item.totalMods) * 100 : 0}%` }}
                                    />
                                  </div>
                                </div>
                                <div className="flex items-center gap-4 text-[10px] leading-none">
                                  <span className="text-green-400/80">
                                    {item.approved} approved
                                  </span>
                                  <span className="text-red-400/80">
                                    {item.rejected} rejected
                                  </span>
                                  <span className="text-yellow-400/80">
                                    {item.awaitingMods} waiting
                                  </span>
                                </div>
                              </div>
                            )}

                            <div className={`font-mono text-[10px] break-all ${isResolved ? 'text-muted-foreground/40' : 'text-muted-foreground/50'}`}>
                              {item.id.substring(0, 24)}...
                            </div>
                          </div>
                        )}
                      </div>
                    );
                  })}
                </>
              )}
            </div>
          )}
        </div>
      </div>

      <div className="w-full px-2 pb-3 pt-4 mt-auto">
        <div className="w-full px-2 py-2 space-y-1">
          <button
            onClick={() => {
              logger.info('[Menubar] Feedback button clicked.');
              BrowserOpenURL("https://libr-ashen.vercel.app/eula");
            }}
            className='flex justify-start hover:bg-muted/30 libr-button w-full items-center space-x-2 rounded-2xl px-3 py-3 text-muted-foreground hover:text-foreground/90'
          >
            <Copyright className="aspect-square h-[40%] opacity-80" />
            <span className="mt-0.5">License & Agreement</span>
          </button>
          <button
            onClick={() => {
              logger.info('[Menubar] Feedback button clicked.');
              BrowserOpenURL("https://forms.gle/Uchqc6Z49aoJwjvZ9");
            }}
            className='flex justify-start hover:bg-muted/30 libr-button w-full items-center space-x-2 rounded-2xl px-3 py-3 text-muted-foreground hover:text-foreground/90'
          >
            <PencilLine className="aspect-square h-[40%] opacity-80" />
            <span className="mt-0.5">Feedback</span>
          </button>
          <button
            onClick={() => {
              logger.info('[Menubar] Website link clicked.');
              BrowserOpenURL("https://libr-ashen.vercel.app/");
            }}
            className="flex justify-start hover:bg-muted/30 libr-button w-full items-center space-x-2 rounded-2xl px-3 py-3 text-muted-foreground hover:text-foreground/90"
          >
            <Globe className="aspect-square h-[40%] opacity-80" />
            <span className="mt-0.5">Visit Website</span>
          </button>
          <button
            onClick={() => {
              logger.info('[Menubar] Open host database dialog.');
              setDialogOpen(true);
            }}
            className="flex justify-start libr-button hover:bg-muted/30 w-full items-center space-x-2 rounded-2xl px-3 py-3 text-muted-foreground hover:text-foreground/90"
          >
            <Database className="aspect-square h-[40%] opacity-80" />
            <span className="mt-0.5">Host a database</span>
          </button>
        </div>
      </div>
    </div>
  );
};
