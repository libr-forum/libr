import React from 'react';
import { motion } from 'framer-motion';
import { logger } from '../logger/logger';

interface GlowingBadgeProps {
  role: 'moderator' | 'member' | 'admin';
  className?: string;
}

export const GlowingBadge: React.FC<GlowingBadgeProps> = ({ role, className = '' }) => {
  logger.debug("Rendering GlowingBadge", { role, className });

  const getBadgeConfig = () => {
    logger.debug("Determining badge config for role", { role });
    switch (role) {
      case 'moderator':
        logger.debug("Badge config selected: moderator");
        return {
          icon: '🛡',
          text: 'Moderator',
          colors: 'from-libr-accent2 to-purple-600',
          shadow: 'shadow-libr-accent2/50',
        };
      case 'admin':
        logger.debug("Badge config selected: admin");
        return {
          icon: '👑',
          text: 'Admin',
          colors: 'from-yellow-500 to-orange-600',
          shadow: 'shadow-yellow-500/50',
        };
      default:
        logger.debug("Badge config selected: member");
        return {
          icon: '🗣',
          text: 'Member',
          colors: 'from-libr-accent1 to-blue-600',
          shadow: 'shadow-libr-accent1/50',
        };
    }
  };

  const config = getBadgeConfig();
  logger.debug("Final badge config", config);

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
