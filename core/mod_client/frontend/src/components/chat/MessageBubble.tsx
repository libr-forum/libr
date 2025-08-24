import React, { useEffect, useState } from 'react';
import { motion } from 'framer-motion';
import { Message,User,useAppStore } from '../../store/useAppStore';
import { Clock, Check, AlertCircle, MoreVertical, Cross } from 'lucide-react';
import DOMPurify from 'dompurify';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import {emojify} from 'node-emoji';
import { Delete, Report,GenerateAlias } from 'wailsjs/go/main/App';
import { types } from 'wailsjs/go/models';
import { parseFormatting,apiService } from '@/services/api';
import { Tooltip, TooltipProvider, TooltipTrigger } from '../ui/tooltip';
import { TooltipContent } from '@radix-ui/react-tooltip';

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
  
  const{setMessages}=useAppStore();

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
  const safeHtml = DOMPurify.sanitize(parseFormatting(body));
  const status = getStatus();
  const user=useAppStore.getState().user;

  const [showReportPopup, setShowReportPopup] = useState(false);

  return (
    <>
      <motion.div
        initial={{ opacity: 0, y: 20, scale: 0.95 }}
        animate={{ opacity: 1, y: 0, scale: 1 }}
        transition={{ duration: 0.3 }}
        className="flex w-full mb-4"
      >
        <div className="w-[99%]">
          <div className="relative rounded-3xl px-4 py-3 bg-card shadow-md border-b max-w-[80vw] break-words">
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
                  {/* Only show time if status is not pending */}
                  {message.status !== 'pending' && (
                    <DropdownMenuItem disabled className="flex items-center justify-between">
                      <span className="text-foreground">Time</span>
                      <span>{formatTime(message.timestamp)}</span>
                    </DropdownMenuItem>
                  )}

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
                        setMessages([]);
                        const retried = await apiService.sendMessage(message.communityId, message.content);
                        // Replace all messages with only the retried message
                        setMessages([retried]);
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
                              <TooltipContent className='bg-muted'>
                                <span className="break-all">{cert.public_key}</span>
                              </TooltipContent>
                            </Tooltip>
                          </TooltipProvider>
                          <p>
                            {cert.status === "1"
                              ? <Check className="w-3 h-3 text-green-500"/>
                              : <Cross className="w-3 h-3 text-red-500"/>}
                          </p>
                          <TooltipProvider>
                            <Tooltip>
                              <TooltipTrigger className="cursor-pointer text-muted-foreground">
                                sign
                              </TooltipTrigger>
                              <TooltipContent className='bg-muted'>
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
                      onClick={() =>{
                        const msg:types.Msg={
                          content:message.content,
                          ts:Number(message.timestamp),
                        }
                        const msgcert=new types.MsgCert({
                          public_key:message.authorPublicKey,
                          msg:msg,
                          mod_certs:message.moderationNote,
                          sign:message.sign,
                          
                        });
                        Delete(msgcert);
                      }}
                      className="text-destructive cursor-pointer hover:bg-destructive/10"
                    >
                      Delete
                    </DropdownMenuItem>
                  ) : (
                    <DropdownMenuItem
                      onClick={() =>{
                        const msg:types.Msg={
                          content:message.content,
                          ts:Number(message.timestamp),
                        }
                        const msgcert=new types.MsgCert({
                          public_key:message.authorPublicKey,
                          msg:msg,
                          mod_certs:message.moderationNote,
                          sign:message.sign,
                          
                        });
                        Report(msgcert,"report reason");
                        setShowReportPopup(true);
                        setTimeout(() => setShowReportPopup(false), 2500);
                      }}
                      className="text-destructive cursor-pointer hover:bg-destructive/10"
                    >
                      Report
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
                  dangerouslySetInnerHTML={{ __html: safeHtml }}
                />
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
        </div>
      </motion.div>
      {showReportPopup && (
        <div className="fixed top-8 left-1/2 transform -translate-x-1/2 z-50 libr-card text-libr-secondary px-6 py-3 rounded-xl shadow-lg font-semibold text-center">
          Your report has been received.<br />
          It will be acted upon soon.
        </div>
      )}
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
