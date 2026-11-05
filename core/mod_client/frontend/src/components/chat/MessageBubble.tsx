import React, { useEffect, useState } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import { Message, User, useAppStore } from '../../store/useAppStore';
import { Clock, Check, AlertCircle, MoreVertical, Cross, X, ImageIcon } from 'lucide-react';
import DOMPurify from 'dompurify';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import { emojify } from 'node-emoji';
import { Delete, Report, GenerateAlias } from 'wailsjs/go/main/App';
import { types } from 'wailsjs/go/models';
import { parseFormatting, apiService } from '@/services/api';
import { Tooltip, TooltipProvider, TooltipTrigger } from '../ui/tooltip';
import { TooltipContent } from '@radix-ui/react-tooltip';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Textarea } from "@/components/ui/textarea";
import { Label } from "@/components/ui/label";
import { toast } from "sonner";

interface MessageBubbleProps {
  message: Message;
}

export const MessageBubble: React.FC<MessageBubbleProps> = ({ message }) => {
  const formatTime = (unixTimestamp: bigint) => {
    const timestampNumber = Number(unixTimestamp);
    const date = new Date(timestampNumber * 1000);

    return new Intl.DateTimeFormat('en-US', {
      hour: '2-digit',
      minute: '2-digit',
      second: '2-digit',
      year: 'numeric',
      month: 'short',
      day: '2-digit',
    }).format(date);
  };

  const { setMessages } = useAppStore();

  const getStatus = () => {
    switch (message.status) {
      case 'approved':
        return { icon: <Check className="w-3 h-3 text-green-500" />, label: 'Approved' };
      case 'pending':
        return { icon: <Clock className="w-3 h-3 text-yellow-500" />, label: 'Pending' };
      case 'rejected':
        return { icon: <AlertCircle className="w-3 h-3 text-red-500" />, label: 'Rejected' };
      default:
        return null;
    }
  };

  const parseMessage = (raw: string): { title?: string; body: string } => {
    const titleMatch = raw.match(/<HEAD>(.*?)<\/HEAD>/s);
    const bodyMatch = raw.match(/<BODY>(.*?)<\/BODY>/s);

    return {
      title: titleMatch?.[1]?.trim(),
      body: bodyMatch?.[1]?.trim() || raw,
    };
  };

  const { title, body } = parseMessage(message.content);

  // Allow DOMPurify to pass through img tags with data: src for base64 images
  const safeHtml = DOMPurify.sanitize(parseFormatting(body), {
    ADD_TAGS: ['img'],
    ADD_ATTR: ['src', 'alt', 'class', 'style'],
    ALLOW_DATA_ATTR: true,
    FORCE_BODY: true,
  });

  // Extract image srcs for separate, richer rendering
  const imgSrcList: string[] = [];
  const tempDiv = document.createElement('div');
  tempDiv.innerHTML = safeHtml;
  tempDiv.querySelectorAll('img').forEach(img => {
    imgSrcList.push(img.src);
    img.remove();
  });
  const safeTextHtml = tempDiv.innerHTML;

  const status = getStatus();
  const { user, isIncognito } = useAppStore();

  const [lightboxSrc, setLightboxSrc] = useState<string | null>(null);
  const [isReportDialogOpen, setIsReportDialogOpen] = useState(false);
  const [reportReason, setReportReason] = useState("");

  const handleReportSubmit = async () => {
    if (isIncognito) {
      setIsReportDialogOpen(false);
      setReportReason("");
      toast.error("Reporting disabled", {
        description: "Turn off incognito before sending reports.",
      });
      return;
    }

    const msg: types.Msg = {
      content: message.content,
      ts: Number(message.timestamp),
    };
    const msgcert = new types.MsgCert({
      public_key: message.authorPublicKey,
      msg: msg,
      mod_certs: message.moderationNote,
      sign: message.sign,
    });

    const result = await Report(msgcert, reportReason || "No reason provided");
    setIsReportDialogOpen(false);
    setReportReason("");
    toast.message("Report submitted", {
      description: typeof result === 'string' && result.length > 0
        ? result.replace(/:[^ ]+:/g, '').trim()
        : "Your report has been sent to moderators for review.",
    });
  };

  return (
    <>
      <motion.div
        initial={{ opacity: 0, y: 20, scale: 0.95 }}
        animate={{ opacity: 1, y: 0, scale: 1 }}
        transition={{ duration: 0.3 }}
        className="flex w-full px-4 mb-3 box-border"
      >
        <div className="w-full relative rounded-3xl px-5 py-4 bg-card shadow-md border border-border/20 break-words overflow-hidden box-border">
            {message.authorPublicKey === user.publicKey && (
              <div className="absolute left-0 top-0 bottom-0 w-[4px] bg-libr-accent1 rounded-l-3xl" />
            )}
            <div className="absolute top-3 right-3">
              <DropdownMenu>
                <DropdownMenuTrigger asChild>
                  <button className="p-1 rounded-full hover:bg-muted transition">
                    <MoreVertical className="w-4 h-4 text-foreground" />
                  </button>
                </DropdownMenuTrigger>
                <DropdownMenuContent
                  side="right"
                  align="start"
                  sideOffset={8}
                  className="z-50 bg-popover border border-border rounded-md shadow-lg p-2 text-sm w-64"
                >
                  <DropdownMenuItem disabled className="flex items-center justify-between">
                    <span className="text-foreground">Time</span>
                    <span>{formatTime(message.timestamp)}</span>
                  </DropdownMenuItem>

                  {status && (
                    <DropdownMenuItem disabled className="flex items-center justify-between">
                      <span className="flex items-center gap-1 text-foreground">
                        {status.icon}
                        {status.label}
                      </span>
                    </DropdownMenuItem>
                  )}
                  {message.status === 'pending' && (
                    <DropdownMenuItem
                      onClick={async () => {
                        let retried: Message;
                        const { messages, setMessages } = useAppStore.getState();
                        if (message.moderationNote && message.moderationNote.length > 0) {
                          // Resend only to DB
                          retried = await apiService.resendToDb(message);
                        } else {
                          // Full moderation resend
                          retried = await apiService.sendMessage(message.communityId, message.content);
                        }
                        setMessages(messages.map(m => m.id === message.id ? retried : m));
                      }}
                      className="text-sm cursor-pointer hover:bg-muted px-2 py-1"
                    >
                      Retry Send
                    </DropdownMenuItem>
                  )}
                  {message.moderationNote && (
                    <div className="px-2 py-1 mt-2 bg-muted/20 text-foreground text-xs rounded">
                      {message.moderationNote.map((cert, index) => (
                        <div key={index} className='flex flex-row gap-4 items-center justify-between'>
                          <TooltipProvider>
                            <Tooltip>
                              <TooltipTrigger className="cursor-pointer text-muted-foreground">
                                <CertAlias publicKey={cert.public_key} />
                              </TooltipTrigger>
                              <TooltipContent>
                                <span className="break-all">{cert.public_key}</span>
                              </TooltipContent>
                            </Tooltip>
                          </TooltipProvider>
                          <p>
                            {cert.status === "1"
                              ? <Check className="w-3 h-3 text-green-500" />
                              : <Cross className="w-3 h-3 text-red-500" />}
                          </p>
                          <TooltipProvider>
                            <Tooltip>
                              <TooltipTrigger className="cursor-pointer text-muted-foreground">
                                sign
                              </TooltipTrigger>
                              <TooltipContent>
                                <span className="break-all">{cert.sign}</span>
                              </TooltipContent>
                            </Tooltip>
                          </TooltipProvider>
                        </div>
                      ))}
                    </div>
                  )}

                  {message.authorPublicKey === user.publicKey ? (
                    <DropdownMenuItem
                      onClick={async () => {
                        const msg: types.Msg = {
                          content: message.content,
                          ts: Number(message.timestamp),
                        }
                        const msgcert = new types.MsgCert({
                          public_key: message.authorPublicKey,
                          msg: msg,
                          mod_certs: message.moderationNote,
                          sign: message.sign,

                        });
                        const result = await Delete(msgcert);
                        toast.message("Delete requested", {
                          description: typeof result === 'string' && result.length > 0
                            ? result.replace(/:[^ ]+:/g, '').trim()
                            : "Your delete request has been sent.",
                        });
                      }}
                      className="text-destructive cursor-pointer hover:bg-destructive/10"
                    >
                      Delete
                    </DropdownMenuItem>
                  ) : (
                    <DropdownMenuItem
                      onClick={() => {
                        if (isIncognito) {
                          toast.error("Reporting disabled", {
                            description: "Turn off incognito before sending reports.",
                          });
                          return;
                        }
                        setIsReportDialogOpen(true);
                      }}
                      className="text-destructive cursor-pointer hover:bg-destructive/10"
                    >
                      {isIncognito ? 'Blocked In Incognito' : 'Report'}
                    </DropdownMenuItem>
                  )}

                </DropdownMenuContent>
              </DropdownMenu>
            </div>

            <div className="flex items-start space-x-3">
              {/* Avatar */}
              {message.avatarSvg && message.avatarSvg !== 'unknown' ? (
                <img
                  src={`data:image/svg+xml;base64,${message.avatarSvg}`}
                  alt="avatar"
                  className="w-8 h-8 rounded-full"
                />
              ) : (
                <div className="w-8 h-8 bg-libr-accent1 rounded-full flex items-center justify-center">
                  <span className="text-white text-sm font-medium">
                    {message.authorAlias.charAt(0).toUpperCase()}
                  </span>
                </div>
              )}

              {/* Header Info */}
              <div className="flex flex-col w-full">
                <span className="text-sm font-medium text-libr-secondary">
                  {message.authorAlias}
                </span>

                {title && (
                  <p className="text-lg font-semibold text-foreground mt-1">{title}</p>
                )}

                <div
                  className="text-sm leading-relaxed text-foreground mt-1 break-words max-w-[55vw] whitespace-pre-wrap message-bubble-content"
                  dangerouslySetInnerHTML={{ __html: safeTextHtml }}
                />

                {/* Image gallery */}
                {imgSrcList.length > 0 && (
                  <div className="mt-3 flex flex-wrap gap-2">
                    {imgSrcList.map((src, idx) => (
                      <div key={idx} className="relative group cursor-pointer" onClick={() => setLightboxSrc(src)}>
                        <img
                          src={src}
                          alt={`image-${idx + 1}`}
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
              </div>
            </div>
          </div>
          {/* Add bullet styling for message content */}
          <style>
            {`
              .message-bubble-content ul {
                list-style-type: disc;
                margin-left: 1.5em;
                padding-left: 1.5em;
              }
              .message-bubble-content ol {
                list-style-type: decimal;
                margin-left: 1.5em;
                padding-left: 1.5em;
              }
              .message-bubble-content li {
                margin-bottom: 0.25em;
              }
            `}
          </style>
      </motion.div>
      {/* Lightbox */}
      <AnimatePresence>
        {lightboxSrc && (
          <motion.div
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            exit={{ opacity: 0 }}
            className="fixed inset-0 z-[100] flex items-center justify-center bg-black/80 backdrop-blur-sm"
            onClick={() => setLightboxSrc(null)}
          >
            <motion.div
              initial={{ scale: 0.85, opacity: 0 }}
              animate={{ scale: 1, opacity: 1 }}
              exit={{ scale: 0.85, opacity: 0 }}
              transition={{ type: 'spring', stiffness: 300, damping: 25 }}
              className="relative max-w-[90vw] max-h-[90vh]"
              onClick={e => e.stopPropagation()}
            >
              <img
                src={lightboxSrc}
                alt="Full size"
                className="max-w-[90vw] max-h-[90vh] rounded-2xl shadow-2xl object-contain"
              />
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

      <Dialog open={isReportDialogOpen} onOpenChange={setIsReportDialogOpen}>
        <DialogContent className="sm:max-w-[425px]">
          <DialogHeader>
            <DialogTitle>Report Message</DialogTitle>
            <DialogDescription>
              Why are you reporting this message? This will be sent to moderators for review.
            </DialogDescription>
          </DialogHeader>
          <div className="grid gap-4 py-4">
            <div className="flex flex-col gap-2">
              <Label htmlFor="reason">Reason</Label>
              <Textarea
                id="reason"
                placeholder="Spam, harassment, hate speech, etc."
                value={reportReason}
                onChange={(e) => setReportReason(e.target.value)}
                className="h-24"
              />
            </div>
          </div>
          <DialogFooter>
            <Button variant="ghost" onClick={() => setIsReportDialogOpen(false)}>Cancel</Button>
            <Button variant="destructive" onClick={handleReportSubmit}>Submit Report</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </>
  );
};

// Helper component to fetch and display alias asynchronously
const CertAlias: React.FC<{ publicKey: string }> = ({ publicKey }) => {
  const [alias, setAlias] = useState<string>(publicKey);

  useEffect(() => {
    let mounted = true;
    (async () => {
      try {
        const result = await GenerateAlias(publicKey);
        if (mounted && result) setAlias(result);
      } catch {
        // fallback to publicKey
      }
    })();
    return () => { mounted = false; };
  }, [publicKey]);

  return <span className="font-semibold">{alias}</span>;
};
