import React from 'react';
import { Header, Hero, TechArch, WhatIsLIBR } from './LandingPageSections';
import { HowItWorks, Community, Footer } from './LandingPageExtended';
import { TechModules, HowToUse, InstallGuide } from './AdditionalSections';

interface LandingPageProps {
  isDarkMode: boolean;
  toggleTheme: () => void;
}

const LandingPage: React.FC<LandingPageProps> = ({ isDarkMode, toggleTheme }) => {
  return (
    <div className="flex flex-col min-h-screen">
      <Header isDark={isDarkMode} toggleTheme={toggleTheme} />
      <main className="flex-1">
        <Hero />  
        <WhatIsLIBR/>
        <HowToUse />
        <TechArch />
        <HowItWorks />
        <TechModules />
        <Community />
      </main>
      <Footer />
    </div>
  );
};

export default LandingPage;
