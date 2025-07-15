
import React from 'react';
import { motion } from 'framer-motion';
import { Message } from '../../store/useAppStore';
import { Clock, Check, AlertCircle } from 'lucide-react';

interface MessageBubbleProps {
  message: Message;
  isOwn?: boolean;
}

export const MessageBubble: React.FC<MessageBubbleProps> = ({ message, isOwn = false }) => {
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

  return (
    <motion.div
      initial={{ opacity: 0, y: 20, scale: 0.95 }}
      animate={{ opacity: 1, y: 0, scale: 1 }}
      transition={{ duration: 0.3 }}
      className={`flex ${isOwn ? 'justify-end' : 'justify-start'} mb-4`}
    >
      <div className={`max-w-lg ${isOwn ? 'order-2' : 'order-1'}`}>
        <div
          className={`rounded-2xl px-4 py-3 shadow-md ${
            isOwn
              ? 'bg-libr-accent1 text-white ml-4'
              : 'bg-card border border-border/50 mr-4'
          }`}
        >
          {!isOwn && (
            <div className="flex items-center space-x-2 mb-2">
              <div className="w-6 h-6 bg-libr-accent1 rounded-full flex items-center justify-center">
                <span className="text-white text-xs font-medium">
                  {message.authorAlias.charAt(0).toUpperCase()}
                </span>
              </div>
              <span className="text-sm font-medium text-libr-secondary">
                {message.authorAlias}
              </span>
            </div>
          )}
          
          <p className={`text-sm leading-relaxed ${isOwn ? 'text-white' : 'text-foreground'}`}>
            {message.content}
          </p>
          
          <div className={`flex items-center justify-between mt-2 space-x-2`}>
            <span className={`text-xs ${isOwn ? 'text-white/70' : 'text-muted-foreground'}`}>
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
      </div>
    </motion.div>
  );
};
