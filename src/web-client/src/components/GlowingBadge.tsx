
import React from 'react';
import { motion } from 'framer-motion';

interface GlowingBadgeProps {
  role: 'moderator' | 'member' | 'admin';
  className?: string;
}

export const GlowingBadge: React.FC<GlowingBadgeProps> = ({ role, className = '' }) => {
  const getBadgeConfig = () => {
    switch (role) {
      case 'moderator':
        return {
          icon: '🛡',
          text: 'Moderator',
          colors: 'from-libr-accent2 to-purple-600',
          shadow: 'shadow-libr-accent2/50',
        };
      case 'admin':
        return {
          icon: '👑',
          text: 'Admin',
          colors: 'from-yellow-500 to-orange-600',
          shadow: 'shadow-yellow-500/50',
        };
      default:
        return {
          icon: '🗣',
          text: 'Member',
          colors: 'from-libr-accent1 to-blue-600',
          shadow: 'shadow-libr-accent1/50',
        };
    }
  };

  const config = getBadgeConfig();

  return (
    <motion.div
      initial={{ scale: 0.8, opacity: 0 }}
      animate={{ scale: 1, opacity: 1 }}
      whileHover={{ scale: 1.05 }}
      className={`inline-flex items-center space-x-1 px-3 py-1 rounded-full bg-gradient-to-r ${config.colors} text-white text-sm font-medium shadow-lg ${config.shadow} libr-glow ${className}`}
    >
      <span>{config.icon}</span>
      <span>{config.text}</span>
    </motion.div>
  );
};
