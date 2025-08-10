import React, { useEffect, useState } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import { Header, Hero, TechArch,WhatIsLIBR} from './components/LandingPageSections';
import { HowItWorks, Community, Footer } from './components/LandingPageExtended';
import { TechModules, HowToUse } from './components/AdditionalSections';
import {BackgroundEffect} from './components/BackgroundEffect';
import {Analytics} from '@vercel/analytics/react';

const ScrollProgress: React.FC = () => {
  const [scrollProgress, setScrollProgress] = useState(0);

  useEffect(() => {
    const updateScrollProgress = () => {
      const scrollPx = document.documentElement.scrollTop;
      const winHeightPx = document.documentElement.scrollHeight - document.documentElement.clientHeight;
      const scrolled = scrollPx / winHeightPx;
      setScrollProgress(scrolled);
    };

    window.addEventListener('scroll', updateScrollProgress);
    return () => window.removeEventListener('scroll', updateScrollProgress);
  }, []);

  return (
    <motion.div
      className="fixed top-0 left-0 right-0 h-1 bg-gradient-to-r from-libr-accent1 to-libr-accent2 z-50 origin-left"
      style={{ scaleX: scrollProgress }}
    />
  );
};

const BackToTop: React.FC = () => {
  const [isVisible, setIsVisible] = useState(false);

  useEffect(() => {
    const toggleVisibility = () => {
      if (window.pageYOffset > 300) {
        setIsVisible(true);
      } else {
        setIsVisible(false);
      }
    };

    window.addEventListener('scroll', toggleVisibility);
    return () => window.removeEventListener('scroll', toggleVisibility);
  }, []);

  const scrollToTop = () => {
    window.scrollTo({
      top: 0,
      behavior: 'smooth'
    });
  };

  return (
    <AnimatePresence>
      {isVisible && (
        <motion.button
          initial={{ opacity: 0, scale: 0 }}
          animate={{ opacity: 1, scale: 1 }}
          exit={{ opacity: 0, scale: 0 }}
          onClick={scrollToTop}
          className="fixed bottom-8 right-8 w-12 h-12 bg-libr-secondary text-libr-primary rounded-full shadow-lg hover:shadow-xl transition-all duration-200 flex items-center justify-center z-40"
          whileHover={{ scale: 1.1 }}
          whileTap={{ scale: 0.9 }}
        >
          ↑
        </motion.button>
      )}
    </AnimatePresence>
  );
};

const App: React.FC = () => {
  const [isDarkMode, setIsDarkMode] = useState(true);

  useEffect(() => {
    // Check for saved theme preference or default to light mode
    const savedTheme = localStorage.getItem('libr-theme');
    const prefersDark = window.matchMedia('(prefers-color-scheme: dark)').matches;
    
    if (savedTheme === 'dark' || (!savedTheme && prefersDark)) {
      setIsDarkMode(true);
    }
  }, []);

  useEffect(() => {
    // Apply theme
    document.documentElement.classList.toggle('dark', isDarkMode);
    localStorage.setItem('libr-theme', isDarkMode ? 'dark' : 'light');
    
    // Update meta theme-color for mobile browsers
    const metaThemeColor = document.querySelector('meta[name="theme-color"]');
    if (metaThemeColor) {
      metaThemeColor.setAttribute('content', isDarkMode ? '#080c18' : '#fdfcf7');
    }
  }, [isDarkMode]);

  const toggleTheme = () => {
    setIsDarkMode(!isDarkMode);
  };

  useEffect(() => {
    // Add keyboard shortcut (Ctrl/Cmd + Shift + T) for theme toggle
    const handleKeyDown = (event: KeyboardEvent) => {
      if ((event.ctrlKey || event.metaKey) && event.shiftKey && event.key === 'T') {
        event.preventDefault();
        setIsDarkMode(prev => !prev);
      }
    };

    document.addEventListener('keydown', handleKeyDown);
    return () => document.removeEventListener('keydown', handleKeyDown);
  }, []);

  useEffect(() => {
    // Disable horizontal scroll globally
    document.documentElement.style.overflowX = 'hidden';
    document.body.style.overflowX = 'hidden';
    // Optionally, remove scroll if previously set
    document.documentElement.style.width = '100%';
    document.body.style.width = '100%';
    return () => {
      document.documentElement.style.overflowX = '';
      document.body.style.overflowX = '';
      document.documentElement.style.width = '';
      document.body.style.width = '';
    };
  }, []);

  return (
    <>
      <BackgroundEffect />
      <Analytics />
      <div className="min-h-screen bg-libr-primary/50 text-foreground relative overflow-x-hidden">
        <ScrollProgress />
        <BackToTop />
        <main className="flex flex-col flex-nowrap">
          <Header isDark={isDarkMode} toggleTheme={toggleTheme} />
          <Hero />  
          <WhatIsLIBR/>
          <HowToUse />
          <TechArch />
          <HowItWorks />
          <TechModules />
          <Community />
          <Footer />
        </main>
      </div>
    </>
  );
};

export default App;
