import React from 'react';
import { motion } from 'framer-motion';
import { Message } from '../../store/useAppStore';
import { Clock, Check, AlertCircle } from 'lucide-react';
import DOMPurify from 'dompurify';

interface MessageBubbleProps {
  message: Message;
}

export function parseFormatting(text: string): string {
  return text
    .replace(/\*\*(.+?)\*\*/g, '<strong>$1</strong>')     // Bold
    .replace(/\*(.+?)\*/g, '<em>$1</em>')                  // Italic
    .replace(/_(.+?)_/g, '<u>$1</u>')                    // Underline
    .replace(/~(.+?)~/g, '<s>$1</s>')                    // Strikethrough
    .replace(/\n/g, '<br/>');                              // Line breaks
}

export const MessageBubble: React.FC<MessageBubbleProps> = ({ message }) => {
  const formatTime = (date: Date) => {
    return new Intl.DateTimeFormat('en-US', {
      hour: '2-digit',
      minute: '2-digit',
    }).format(date);
  };

  const getStatusIcon = () => {
    switch (message.status) {
      case 'approved':
        return <Check className="w-3 h-3 text-green-500" />;
      case 'pending':
        return <Clock className="w-3 h-3 text-yellow-500" />;
      case 'rejected':
        return <AlertCircle className="w-3 h-3 text-red-500" />;
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

  return (
    <motion.div
      initial={{ opacity: 0, y: 20, scale: 0.95 }}
      animate={{ opacity: 1, y: 0, scale: 1 }}
      transition={{ duration: 0.3 }}
      className="flex justify-center mb-4"
    >
      <div className="max-w-lg w-full">
        <div className="rounded-2xl px-4 py-3 shadow-md bg-card border border-border/50">
          <div className="flex items-start space-x-3 mb-2">
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
              <div className="flex items-center justify-between">
                <span className="text-sm font-medium text-libr-secondary">
                  {message.authorAlias}
                </span>
                <div className="flex items-center space-x-2">
                  <span className="text-xs text-muted-foreground">
                    {formatTime(message.timestamp)}
                  </span>
                  <div className="flex items-center space-x-1">
                    {getStatusIcon()}
                    {message.status === 'pending' && (
                      <span className="text-xs text-yellow-600">Pending</span>
                    )}
                  </div>
                </div>
              </div>

              {/* Title */}
              {title && (
                <p className="text-lg font-semibold text-foreground mt-1">
                  {title}
                </p>
              )}

              {/* Body */}
              <p className="text-sm leading-relaxed text-foreground mt-1" dangerouslySetInnerHTML={{ __html: safeHtml }}/>
            </div>
          </div>
        </div>
      </div>
    </motion.div>
  );
};
