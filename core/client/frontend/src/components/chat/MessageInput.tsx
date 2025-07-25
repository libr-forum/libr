import React, { useState,useRef } from 'react';
import { motion } from 'framer-motion';
import { Send, Smile, Paperclip } from 'lucide-react';
import { useAppStore } from '../../store/useAppStore';
import { useSidebarStore } from '../../store/useSidebarStore';
import { apiService } from '../../services/api';

export const MessageInput: React.FC = () => {
  const [message, setMessage] = useState('');
  const [isSending, setIsSending] = useState(false);
  const { currentCommunity, addMessage } = useAppStore();
  const { isCollapsed } = useSidebarStore();

  const textareaRef = useRef<HTMLTextAreaElement>(null);
  React.useEffect(() => {
    if (message === '') {
      textareaRef.current?.focus();
    }
  }, [message]);


  const handleSend = async () => {
    if (!message.trim() || !currentCommunity || isSending) return;

    setIsSending(true);
    try {
      const newMessage = await apiService.sendMessage(currentCommunity.id, message);
      addMessage(newMessage);
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

  return (
    <motion.div
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
      <div className="flex items-end space-x-3 max-w-full">
        <div className="flex-1">
          <div className="relative">
            <textarea
              ref={textareaRef}
              value={message}
              onChange={(e) => setMessage(e.target.value)}
              onKeyPress={handleKeyPress}
              placeholder={`Message #${currentCommunity?.name || 'community'}...`}
              disabled={isSending}
              className="w-full min-h-[44px] max-h-32 p-3 bg-muted/30 border border-border/50 rounded-2xl resize-none focus:outline-none focus:ring-2 focus:ring-libr-accent1/50 focus:border-libr-accent1/50 transition-all duration-200"
              rows={1}
            />
            <div className="absolute right-3 bottom-3 flex items-center space-x-2">
              <button
                type="button"
                className="text-muted-foreground hover:text-libr-accent1 transition-colors duration-200"
              >
                <Smile className="w-5 h-5" />
              </button>
              <button
                type="button"
                className="text-muted-foreground hover:text-libr-accent1 transition-colors duration-200"
              >
                <Paperclip className="w-5 h-5" />
              </button>
            </div>
          </div>
        </div>
        
        <motion.button
          whileHover={{ scale: 1.05 }}
          whileTap={{ scale: 0.95 }}
          onClick={handleSend}
          disabled={!message.trim() || isSending}
          className="w-11 h-11 bg-libr-accent1 hover:bg-libr-accent1/80 disabled:bg-muted disabled:cursor-not-allowed rounded-2xl flex items-center justify-center transition-all duration-200"
        >
          {isSending ? (
            <div className="w-5 h-5 border-2 border-white border-t-transparent rounded-full animate-spin" />
          ) : (
            <Send className="w-5 h-5 text-white" />
          )}
        </motion.button>
      </div>
    </motion.div>
  );
};
