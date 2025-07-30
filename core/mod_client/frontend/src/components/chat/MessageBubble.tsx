import React from 'react';
import { motion } from 'framer-motion';
import { Message,User,useAppStore } from '../../store/useAppStore';
import { Clock, Check, AlertCircle, MoreVertical } from 'lucide-react';
import DOMPurify from 'dompurify';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import {emojify} from 'node-emoji';
import { Report } from 'wailsjs/go/main/App';

interface MessageBubbleProps {
  message: Message;
}

export function parseFormatting(text: string): string {
  // Escape HTML to prevent injection
  const escapeHTML = (str: string) =>
    str.replace(/&/g, '&amp;')
       .replace(/</g, '&lt;')
       .replace(/>/g, '&gt;');

  // Apply emoji replacements first
  let formatted = emojify(text);

  // Code blocks (```...```)
  formatted = formatted.replace(/```([\s\S]*?)```/g, (_match, code) => {
    return `<pre class="bg-muted rounded p-2 overflow-x-auto my-2 text-xs"><code>${escapeHTML(code)}</code></pre>`;
  });

  // Inline code (`...`)
  formatted = formatted.replace(/`([^`\n]+?)`/g, (_match, code) => {
    return `<code class="bg-muted px-1 rounded text-xs">${escapeHTML(code)}</code>`;
  });

  // Bold (**bold**)
  formatted = formatted.replace(/\*\*(.+?)\*\*/g, '<strong>$1</strong>');

  // Italic (*italic*)
  formatted = formatted.replace(/\*(.+?)\*/g, '<em>$1</em>');

  // Underline (_underline_)
  formatted = formatted.replace(/_(.+?)_/g, '<u>$1</u>');

  // Strikethrough (~strike~)
  formatted = formatted.replace(/~(.+?)~/g, '<s>$1</s>');

  // Newlines to <br/>
  formatted = formatted.replace(/\n/g, '<br/>');

  return formatted;
}


export const MessageBubble: React.FC<MessageBubbleProps> = ({ message }) => {
  const formatTime = (date: Date) => {
    return new Intl.DateTimeFormat('en-US', {
      hour: '2-digit',
      minute: '2-digit',
    }).format(date);
  };

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

  return (
    <motion.div
      initial={{ opacity: 0, y: 20, scale: 0.95 }}
      animate={{ opacity: 1, y: 0, scale: 1 }}
      transition={{ duration: 0.3 }}
      className="flex w-full mb-4"
    >
      <div className="w-[99%]">
        <div className="relative rounded-3xl px-4 py-3 border-b">
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
                {message.moderationNote && (
                  <div className="px-2 py-1 mt-2 bg-muted/20 text-foreground text-xs rounded">
                    {message.moderationNote}
                  </div>
                )}

                {message.authorId === user.publicKey ? (
                  <DropdownMenuItem
                    onClick={() => console.log('Delete message:', message.id)}
                    className="text-destructive cursor-pointer hover:bg-destructive/10"
                  >
                    Delete
                  </DropdownMenuItem>
                ) : (
                  <DropdownMenuItem
                    onClick={() => Report(message.content,Math.floor(message.timestamp.getTime() / 1000),'',message.authorId)}
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

              <p
                className="text-sm leading-relaxed text-foreground mt-1"
                dangerouslySetInnerHTML={{ __html: safeHtml }}
              />
            </div>
          </div>
        </div>
      </div>
    </motion.div>
  );
};
