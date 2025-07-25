import React, { useState, useRef,useImperativeHandle,forwardRef } from 'react';
import { motion } from 'framer-motion';
import { Send, Smile, Paperclip, User } from 'lucide-react';
import { useAppStore } from '../../store/useAppStore';
import { useSidebarStore } from '../../store/useSidebarStore';
import { apiService } from '../../services/api';
import { cn } from '../../lib/utils'; // Assuming you use `cn` for conditional classes

export const MessageInput= forwardRef<HTMLDivElement>((props, ref) => {
  const [title, setTitle] = useState('');
  const [message, setMessage] = useState('');
  const [isSending, setIsSending] = useState(false);
  const [shake, setShake] = useState(false);
  
  const { currentCommunity, addMessage } = useAppStore();
  const { isCollapsed } = useSidebarStore();
  const textareaRef = useRef<HTMLTextAreaElement>(null);
  const [isOverflowing, setIsOverflowing] = React.useState(false);
  

  React.useEffect(() => {
    if (message === '') {
      textareaRef.current?.focus();
    }
  }, [message]);

  React.useEffect(() => {
    if (!textareaRef.current) return;

    const el = textareaRef.current;
    el.style.height = 'auto';
    el.style.height = el.scrollHeight + 'px';
  }, [message]);

  React.useEffect(() => {
    const el = textareaRef.current;
    if (!el) return;

    el.style.height = 'auto'; // reset
    el.style.height = `${el.scrollHeight}px`;

    setIsOverflowing(el.scrollHeight > 200); // mark overflow if beyond 200px
  }, [message]);

  const handleSend = async () => {
    const trimmedTitle = title.trim();
    const trimmedMessage = message.trim();

    if (!trimmedMessage && trimmedTitle) {
      setShake(true);
      textareaRef.current?.focus();
      setTimeout(() => setShake(false), 400);
      return;
    }

    if (!trimmedMessage || !currentCommunity || isSending) return;

    setIsSending(true);
    try {
      const formatted = trimmedTitle
        ? `<HEAD>${trimmedTitle}</HEAD><BODY>${trimmedMessage}</BODY>`
        : `<BODY>${trimmedMessage}</BODY>`;

      const newMessage = await apiService.sendMessage(currentCommunity.id, formatted);
      addMessage(newMessage);
      setTitle('');
      setMessage('');
    } catch (error) {
      console.error('Failed to send message:', error);
    } finally {
      setIsSending(false);
    }
  };

  const handleKeyPress = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      handleSend();
    }
  };

  const containerRef = useRef<HTMLDivElement>(null);
  
  useImperativeHandle(ref, () => containerRef.current!);

  return (
    <motion.div
      ref={containerRef}
      initial={{ y: 50, opacity: 0 }}
      animate={{ 
        y: 0, 
        opacity: 1,
        x: isCollapsed ? 80 : 320,
        width: isCollapsed ? 'calc(100vw - 80px)' : 'calc(100vw - 320px)'
      }}
      transition={{ duration: 0.3 }}
      className="bg-card border-t border-border/50 p-4 fixed bottom-0 left-0 z-50"
    >
      <div className="flex flex-col space-y-2">
        <input
          type="text"
          value={title}
          onChange={(e) => setTitle(e.target.value)}
          placeholder="Title (optional)"
          className="w-full px-3 py-2 text-sm border rounded-2xl bg-muted/30 text-foreground placeholder-muted-foreground focus:outline-none focus:ring-2 focus:ring-libr-accent1/50 focus:border-libr-accent1/50 mb-1"
        />

        <div className="flex items-start space-x-3 max-w-full">
          <div className="flex-1 h-auto">
            <div className="relative h-full translate-y-1">
              <textarea
                ref={textareaRef}
                value={message}
                onChange={(e) => setMessage(e.target.value)}
                onKeyDown={handleKeyPress}
                placeholder={`Message #${currentCommunity?.name || 'community'}...`}
                disabled={isSending}
                className={cn(
                  "w-full min-h-[40px] max-h-[200px] p-3 bg-muted/30 border border-border/50 rounded-2xl resize-none focus:outline-none focus:ring-2 focus:ring-libr-accent1/50 focus:border-libr-accent1/50 transition-all duration-200",
                  isOverflowing ? "overflow-y-auto" : "overflow-hidden",
                  !isOverflowing && "scrollbar-none",
                  shake && "border-red-500 ring-red-500 animate-shake"
                )}
                rows={1}
              />
            </div>
          </div>

          <motion.button
            whileHover={{ scale: 1.05 }}
            whileTap={{ scale: 0.95 }}
            onClick={handleSend}
            disabled={!message.trim() || isSending}
            className="w-14 h-14 bg-libr-accent1 hover:bg-libr-accent1/80 disabled:bg-muted disabled:cursor-not-allowed rounded-2xl flex items-center justify-center transition-all duration-200"
          >
            {isSending ? (
              <div className="w-5 h-5 border-2 border-white border-t-transparent rounded-full animate-spin" />
            ) : (
              <Send className="w-5 h-5 text-white" />
            )}
          </motion.button>
        </div>
      </div>
    </motion.div>
  );
});
