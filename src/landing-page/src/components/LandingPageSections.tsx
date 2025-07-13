import React from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import { Shield, Users, Globe, Lock, MessageSquare, Zap, Moon, Sun } from 'lucide-react';

interface HeaderProps {
  isDark?: boolean;
  toggleTheme?: () => void;
}

const Header: React.FC<HeaderProps> = ({ isDark = false, toggleTheme }) => {
  return (
    <motion.header 
      initial={{ y: -100, opacity: 0 }}
      animate={{ y: 0, opacity: 1 }}
      transition={{ duration: 0.8 }}
      className="fixed top-0 left-0 right-0 z-50 bg-libr-primary/80 backdrop-blur-lg border-b border-border/50"
    >
      <nav className="container mx-auto section-padding py-4 flex items-center justify-between">
        <motion.div 
          whileHover={{ scale: 1.05 }}
          className="flex items-center space-x-2"
        >
          <div className="w-8 h-8 bg-gradient-to-r from-libr-accent1 to-libr-accent2 rounded-lg flex items-center justify-center">
            <Shield className="w-5 h-5 text-white" />
          </div>
          <span className="text-2xl font-bold text-libr-secondary">Libr</span>
        </motion.div>
        
        <div className="hidden md:flex items-center space-x-8">
          <a href="#features" className="text-foreground hover:text-libr-accent1 transition-colors">Architecture</a>
          <a href="#how-it-works" className="text-foreground hover:text-libr-accent1 transition-colors">Protocol</a>
          <a href="#community" className="text-foreground hover:text-libr-accent1 transition-colors">Research</a>
          <a href="#roadmap" className="text-foreground hover:text-libr-accent1 transition-colors">Roadmap</a>
        </div>
        
        <div className="flex items-center gap-4">
          {/* Theme Toggle Button */}
          {toggleTheme && (
            <motion.button
              onClick={toggleTheme}
              className="w-10 h-10 rounded-lg bg-card border border-border/50 shadow-sm hover:shadow-md flex items-center justify-center transition-all duration-200 backdrop-blur-sm hover:border-libr-accent1/30"
              whileHover={{ scale: 1.05 }}
              whileTap={{ scale: 0.95 }}
              title={isDark ? "Switch to light mode (Ctrl+Shift+T)" : "Switch to dark mode (Ctrl+Shift+T)"}
            >
              <AnimatePresence mode="wait" initial={false}>
                <motion.div
                  key={isDark ? 'moon' : 'sun'}
                  initial={{ y: -20, opacity: 0, rotate: -90 }}
                  animate={{ y: 0, opacity: 1, rotate: 0 }}
                  exit={{ y: 20, opacity: 0, rotate: 90 }}
                  transition={{ duration: 0.2, ease: "easeInOut" }}
                >
                  {isDark ? (
                    <Moon className="w-5 h-5 text-libr-accent1" />
                  ) : (
                    <Sun className="w-5 h-5 text-libr-accent2" />
                  )}
                </motion.div>
              </AnimatePresence>
            </motion.button>
          )}
          
          <motion.button 
            whileHover={{ scale: 1.05 }}
            whileTap={{ scale: 0.95 }}
            className="libr-button-primary"
          >
            View Docs
          </motion.button>
        </div>
      </nav>
    </motion.header>
  );
};

const Hero: React.FC = () => {
  return (
    <section className="hero-gradient min-h-screen flex items-center section-padding pt-20">
      <div className="container mx-auto">
        <div className="grid lg:grid-cols-2 gap-12 items-center">
          <motion.div
            initial={{ x: -100, opacity: 0 }}
            animate={{ x: 0, opacity: 1 }}
            transition={{ duration: 0.8, delay: 0.2 }}
          >
            <motion.h1 
              className="text-5xl lg:text-7xl font-bold text-libr-secondary mb-6"
              initial={{ y: 50, opacity: 0 }}
              animate={{ y: 0, opacity: 1 }}
              transition={{ duration: 0.8, delay: 0.4 }}
            >
              Censorship-Resistant
              <span className="bg-gradient-to-r from-libr-accent1 to-libr-accent2 bg-clip-text text-transparent"> Forums</span>
            </motion.h1>
            
            <motion.p 
              className="text-xl text-muted-foreground mb-8 leading-relaxed"
              initial={{ y: 30, opacity: 0 }}
              animate={{ y: 0, opacity: 1 }}
              transition={{ duration: 0.8, delay: 0.6 }}
            >
              A novel framework for creating censorship-resilient yet moderated public forums. LIBR combines distributed hash tables, consensus protocols, and community-driven moderation to preserve free expression while ensuring constructive dialogue.
            </motion.p>
            
            <motion.div 
              className="flex flex-col sm:flex-row gap-4"
              initial={{ y: 30, opacity: 0 }}
              animate={{ y: 0, opacity: 1 }}
              transition={{ duration: 0.8, delay: 0.8 }}
            >
              <button className="libr-button-primary text-lg">
                <MessageSquare className="w-5 h-5 mr-2" />
                Join Beta Community
              </button>
              <button className="libr-button-secondary text-lg">
                Read Documentation
              </button>
            </motion.div>
            
            <motion.div 
              className="flex items-center gap-6 mt-8 text-sm text-muted-foreground"
              initial={{ opacity: 0 }}
              animate={{ opacity: 1 }}
              transition={{ duration: 0.8, delay: 1 }}
            >
              <div className="flex items-center gap-2">
                <Users className="w-4 h-4 text-libr-accent1" />
                <span>Open Source Project</span>
              </div>
              <div className="flex items-center gap-2">
                <Globe className="w-4 h-4 text-libr-accent1" />
                <span>Built with Go</span>
              </div>
            </motion.div>
          </motion.div>
          
          <motion.div
            initial={{ x: 100, opacity: 0 }}
            animate={{ x: 0, opacity: 1 }}
            transition={{ duration: 0.8, delay: 0.4 }}
            className="relative"
          >
            <div className="libr-card p-8 bg-gradient-to-br from-card to-muted/30">
              <motion.div 
                className="space-y-4"
                initial={{ opacity: 0 }}
                animate={{ opacity: 1 }}
                transition={{ duration: 0.8, delay: 1.2 }}
              >
                <div className="flex items-center justify-between">
                  <div className="flex items-center gap-3">
                    <div className="w-10 h-10 bg-libr-accent1 rounded-full flex items-center justify-center">
                      <Users className="w-5 h-5 text-white" />
                    </div>
                    <div>
                      <p className="font-semibold">Community: Research Network</p>
                      <p className="text-sm text-muted-foreground">12 moderators active</p>
                    </div>
                  </div>
                  <div className="w-3 h-3 bg-green-500 rounded-full animate-pulse-glow"></div>
                </div>
                
                <div className="space-y-3">
                  <div className="bg-muted/50 rounded-lg p-3">
                    <p className="text-sm">@researcher_a: The DHT replication factor should be adjusted based on network size</p>
                  </div>
                  <div className="bg-libr-accent1/10 rounded-lg p-3 ml-6">
                    <p className="text-sm">@mod_bob: Validated. This follows our technical governance protocols.</p>
                  </div>
                  <div className="bg-muted/50 rounded-lg p-3">
                    <p className="text-sm">@charlie: Byzantine fault tolerance ensures consistency even with malicious nodes.</p>
                  </div>
                </div>
                
                <div className="flex items-center gap-2 pt-2">
                  <Lock className="w-4 h-4 text-libr-accent1" />
                  <span className="text-sm text-muted-foreground">Cryptographically signed & validated</span>
                </div>
              </motion.div>
            </div>
            
            <motion.div 
              className="absolute -top-4 -right-4 w-20 h-20 bg-gradient-to-r from-libr-accent1 to-libr-accent2 rounded-full opacity-20 animate-float"
              animate={{ rotate: 360 }}
              transition={{ duration: 20, repeat: Infinity, ease: "linear" }}
            />
          </motion.div>
        </div>
      </div>
    </section>
  );
};

const Features: React.FC = () => {
  const features = [
    {
      icon: Shield,
      title: "Censorship Resistant",
      description: "Built on distributed hash tables (DHTs) for immutable message storage with partial immutability for efficient forum operations."
    },
    {
      icon: Users,
      title: "Community Moderation",
      description: "Byzantine Consistent Broadcast ensures democratic moderation quorums with 2f+1 moderator consensus for content validation."
    },
    {
      icon: Lock,
      title: "Cryptographic Security",
      description: "Digital signatures and moderation certificates provide tamper-proof validation with public key cryptography."
    },
    {
      icon: Zap,
      title: "Efficient Protocol",
      description: "Go-based implementation with optimized DHT lookup and concurrent message processing for high performance."
    },
    {
      icon: Globe,
      title: "Decentralized Architecture",
      description: "Role-based node system with clients, database nodes, and moderators operating without central authority."
    },
    {
      icon: MessageSquare,
      title: "Message Integrity",
      description: "Hashchain-based state reconstruction and replicated storage ensure data availability and consistency."
    }
  ];

  return (
    <section id="features" className="py-20 section-padding">
      <div className="container mx-auto">
        <motion.div 
          className="text-center mb-16"
          initial={{ y: 50, opacity: 0 }}
          whileInView={{ y: 0, opacity: 1 }}
          transition={{ duration: 0.8 }}
          viewport={{ once: true }}
        >
          <h2 className="text-4xl lg:text-5xl font-bold text-libr-secondary mb-4">
            Technical Architecture
          </h2>
          <p className="text-xl text-muted-foreground max-w-3xl mx-auto">
            LIBR leverages cutting-edge distributed systems concepts to create a platform where 
            censorship resistance meets community-driven governance through innovative protocol design.
          </p>
        </motion.div>
        
        <div className="grid md:grid-cols-2 lg:grid-cols-3 gap-8">
          {features.map((feature, index) => (
            <motion.div
              key={feature.title}
              className="feature-card"
              initial={{ y: 50, opacity: 0 }}
              whileInView={{ y: 0, opacity: 1 }}
              transition={{ duration: 0.6, delay: index * 0.1 }}
              viewport={{ once: true }}
            >
              <div className="w-12 h-12 bg-gradient-to-r from-libr-accent1 to-libr-accent2 rounded-lg flex items-center justify-center mb-4">
                <feature.icon className="w-6 h-6 text-white" />
              </div>
              <h3 className="text-xl font-semibold text-libr-secondary mb-2">{feature.title}</h3>
              <p className="text-muted-foreground">{feature.description}</p>
            </motion.div>
          ))}
        </div>
      </div>
    </section>
  );
};

export { Header, Hero, Features };
