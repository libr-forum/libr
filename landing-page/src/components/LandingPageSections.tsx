import React,{useRef,useEffect,useState} from 'react';
import { Link, useLocation, useNavigate } from 'react-router-dom';
import { motion, AnimatePresence } from 'framer-motion';
import { Shield, Users, Globe, Lock, Moon, Sun, DatabaseZap, VenetianMask, Waypoints, Download, X, Menu,ChevronDown} from 'lucide-react';
import icon_transparent from "../assets/icon_transparent.png"
import { FaWindows, FaApple, FaLinux } from 'react-icons/fa';

interface HeaderProps {
  isDark?: boolean;
  toggleTheme?: () => void;
}
const isMobile = typeof window !== "undefined" ? window.innerWidth < 768 : false;

const Header: React.FC<HeaderProps> = ({ isDark = false, toggleTheme }) => {
  const [isMenuOpen, setIsMenuOpen] = useState(false);
  const menuRef = useRef<HTMLDivElement>(null);
  const location = useLocation();
  const navigate = useNavigate();
  const isLegalPage = ['/privacy-policy', '/terms-and-conditions', '/eula'].includes(location.pathname);

  const handleSectionNavigation = (href: string) => {
    if (isLegalPage && href.startsWith('#')) {
      // Navigate to home page and then scroll to section
      navigate('/');
      setTimeout(() => {
        const element = document.querySelector(href);
        if (element) {
          element.scrollIntoView({ behavior: 'smooth' });
        }
      }, 100);
    }
  };

  // Close menu on outside click
  useEffect(() => {
    const handleClickOutside = (e: MouseEvent) => {
      if (menuRef.current && !menuRef.current.contains(e.target as Node)) {
        setIsMenuOpen(false);
      }
    };

    if (isMenuOpen) {
      document.addEventListener("mousedown", handleClickOutside);
    } else {
      document.removeEventListener("mousedown", handleClickOutside);
    }

    return () => document.removeEventListener("mousedown", handleClickOutside);
  }, [isMenuOpen]);

  const navLinks = [
    { href: "#what-is-libr", label: "Product" },
    { href: "#how-to-use", label: "How To" },
    { href: "#features", label: "Architecture" },
    { href: "#how-it-works", label: "Protocol" },
    { href: "#technical-modules", label: "Modules" },
    { href: "https://medium.com/@libr.forum/libr-a-moderated-censorship-resilient-social-network-framework-ecfcffb3fdae", label: "Docs", external: true },
    { href: "https://github.com/libr-forum/libr", label: "GitHub", external: true },
    { href: "https://forms.gle/Uchqc6Z49aoJwjvZ9", label: "Feedback", external: true }
  ];

  return (
    <motion.header
      initial={{ y: -100, opacity: 0 }}
      animate={{ y: 0, opacity: 1 }}
      transition={{ duration: 0.8 }}
      className="fixed top-0 left-0 right-0 z-50 bg-libr-primary/80 backdrop-blur-lg border-b border-border/50"
    >
      <nav className="container mx-auto section-padding py-4 flex items-center justify-between">
        {/* Logo */}
        <div className="flex items-center space-x-2">
          {isLegalPage ? (
            <Link to="/" className="flex items-center">
              <div className="w-8 h-8 rounded-lg flex items-center justify-center">
                <img src={icon_transparent} className="w-8 h-8 rounded-lg" />
              </div>
              <span className="text-2xl font-bold text-libr-secondary">libr</span>
            </Link>
          ) : (
            <a href="#welcome" className="flex items-center">
              <div className="w-8 h-8 rounded-lg flex items-center justify-center">
                <img src={icon_transparent} className="w-8 h-8 rounded-lg" />
              </div>
              <span className="text-2xl font-bold text-libr-secondary">libr</span>
            </a>
          )}
        </div>

        {/* Desktop Nav Links */}
        <div className="hidden 980:flex items-center space-x-8">
          {navLinks.map(({ href, label, external }) => {
            if (isLegalPage && !external && href.startsWith('#')) {
              return (
                <button
                  key={href}
                  onClick={() => handleSectionNavigation(href)}
                  className="text-foreground hover:text-libr-accent1 transition-colors bg-transparent border-none cursor-pointer"
                >
                  {label}
                </button>
              );
            }
            return (
              <a
                key={href}
                href={href}
                target={external ? "_blank" : undefined}
                rel={external ? "noopener noreferrer" : undefined}
                className="text-foreground hover:text-libr-accent1 transition-colors"
              >
                {label}
              </a>
            );
          })}
        </div>

        {/* Mobile Menu Toggle + Theme Toggle */}
        <div className="flex items-center gap-4 980:gap-2">
          {/* Hamburger Menu (mobile only) */}
          <motion.button
            onClick={() => setIsMenuOpen(prev => !prev)}
            className="w-10 h-10 rounded-lg bg-card border border-border/50 shadow-sm hover:shadow-md flex items-center justify-center transition-all duration-200 backdrop-blur-sm hover:border-libr-accent1/30 sm:hidden"
          >
            {isMenuOpen ? <X className="w-5 h-5 text-libr-secondary" /> : <Menu className="w-5 h-5 text-libr-secondary" />}
          </motion.button>

          {/* Theme Toggle */}
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
                  key={isDark ? "moon" : "sun"}
                  initial={{ y: -20, opacity: 0, rotate: -90 }}
                  animate={{ y: 0, opacity: 1, rotate: 0 }}
                  exit={{ y: 20, opacity: 0, rotate: 90 }}
                  transition={{ duration: 0.2, ease: "easeInOut" }}
                >
                  {isDark ? (
                    <Sun className="w-5 h-5 text-libr-secondary" />
                  ) : (
                    <Moon className="w-5 h-5 text-libr-secondary" />
                  )}
                </motion.div>
              </AnimatePresence>
            </motion.button>
          )}
        </div>
      </nav>

      {/* Mobile Menu Dropdown */}
      <AnimatePresence>
        {isMenuOpen && (
          <motion.div
            ref={menuRef}
            initial={{ height: 0, opacity: 0 }}
            animate={{ height: "auto", opacity: 1 }}
            exit={{ height: 0, opacity: 0 }}
            transition={{ duration: 0.4, ease: [0.4, 0, 0.2, 1] }}
            className="980:hidden bg-libr-primary/90 backdrop-blur-xl border-t border-border/50 px-6 pb-4 pt-2 flex flex-col space-y-3"
          >
            {navLinks.map(({ href, label, external }) => {
              if (isLegalPage && !external && href.startsWith('#')) {
                return (
                  <button
                    key={href}
                    onClick={() => {
                      handleSectionNavigation(href);
                      setIsMenuOpen(false);
                    }}
                    className="text-foreground hover:text-libr-accent1 transition-colors text-base bg-transparent border-none cursor-pointer text-left"
                  >
                    {label}
                  </button>
                );
              }
              return (
                <a
                  key={href}
                  href={href}
                  target={external ? "_blank" : undefined}
                  rel={external ? "noopener noreferrer" : undefined}
                  onClick={() => setIsMenuOpen(false)}
                  className="text-foreground hover:text-libr-accent1 transition-colors text-base"
                >
                  {label}
                </a>
              );
            })}
          </motion.div>
        )}
      </AnimatePresence>
    </motion.header>
  );
};

const JoinBetaDropdown = () => {
  const [open, setOpen] = React.useState(false);
  const dropdownRef = React.useRef<HTMLDivElement>(null);

  // Optional: close dropdown when clicking outside
  React.useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (dropdownRef.current && !dropdownRef.current.contains(event.target as Node)) {
        setOpen(false);
      }
    };
    document.addEventListener("mousedown", handleClickOutside);
    return () => document.removeEventListener("mousedown", handleClickOutside);
  }, []);

  return (
    <div id="join-beta" className="relative flex flex-row h-full items-center justify-center" ref={dropdownRef}>
      <button
        onClick={() => setOpen(!open)}
        className="libr-button bg-libr-secondary flex flex-row items-center justify-center gap-2 text-libr-primary"
      >
        <Download className='w-5 h-5 mr-3'/>
        Join Beta
        <ChevronDown size={16} className={`transition-transform ${open ? "rotate-180" : ""}`} />
      </button>

      {open && (
        <div className="absolute top-[60%] mt-2 w-40 bg-libr-secondary rounded shadow-lg z-50">
          <div className="flex flex-col text-sm text-center text-libr-primary">
            <div
              onClick={() => window.open("https://github.com/libr-forum/libr/releases/download/v1.0.0-beta/libr-amd64-installer.exe", "_blank")}
              className="flex flex-row items-center justify-center gap-2 libr-button px-4 py-2 cursor-pointer"
            >
              <FaWindows/>
              Windows
            </div>
            <div
              onClick={() => window.open("https://github.com/libr-forum/libr/releases/download/v1.0.0-beta/libr-linux-amd64", "_blank")}
              className="flex flex-row items-center justify-center gap-2 libr-button px-4 py-2 cursor-pointer"
            >
              <FaLinux/>
              Linux
            </div>
            <div
              onClick={() => window.open("https://github.com/libr-forum/libr/releases/download/v1.0.0-beta/libr-darwin-amd64.out", "_blank")}
              className="flex flex-row items-center justify-center gap-2 libr-button px-4 py-2 cursor-pointer"
            >
              <FaApple/>
              MacOS
            </div>
          </div>
        </div>
      )}
    </div>
  );
};

const Hero:React.FC = () => {
  const audioRef = React.useRef<HTMLAudioElement>(null);

  // Store gain node and audio context in a ref to avoid TS errors
  const audioBoostRef = React.useRef<{ctx?: AudioContext, gainNode?: GainNode}>({});
  const handlePlay = () => {
    const audio = audioRef.current;
    if (audio) {
      audio.currentTime = 0;
      audio.volume = 1.0;
      // Try to boost volume using Web Audio API if possible
      try {
        if (!audioBoostRef.current.ctx || !audioBoostRef.current.gainNode) {
          const ctx = new window.AudioContext();
          const source = ctx.createMediaElementSource(audio);
          const gainNode = ctx.createGain();
          gainNode.gain.value = 2.5; // 2.5x boost (max safe for most browsers)
          source.connect(gainNode).connect(ctx.destination);
          audioBoostRef.current = { ctx, gainNode };
        } else {
          audioBoostRef.current.gainNode.gain.value = 2.5;
        }
      } catch (e) {
        // Fallback: do nothing if Web Audio API fails
      }
      audio.play().catch((err) => {
        console.error("Playback failed", err);
      });
    }
  };
  const scaleRef = React.useRef<HTMLDivElement>(null);
  React.useEffect(() => {
    const el = scaleRef.current;
    if (!el) return;
    const resize = () => {
      if (!el.parentElement) return;
      const parentWidth = el.parentElement.offsetWidth;
      const elWidth = el.scrollWidth;
      let scale = 1;
      if (elWidth > parentWidth) {
        scale = parentWidth / elWidth;
      }
      el.style.transform = scale < 1 ? `scale(${scale})` : '';
    };
    resize();
    window.addEventListener('resize', resize);
    return () => window.removeEventListener('resize', resize);
  }, []);
  return(
    <section id="welcome" className="min-h-screen flex items-center justify-center section-padding pt-20">
      <div className="h-screen w-screen pb-20 flex items-center justify-center">
        <motion.div
          initial={{ x: -100, opacity: 0 }}
          whileInView={{ x: 0, opacity: 1 }}
          transition={{
            duration: 0.8,
            ease: [0.4, 0, 0.2, 1],
          }}
          viewport={{ once: false }}
          className="w-full h-full flex items-center justify-center"
        >
          <div className='flex flex-col p-0 items-center justify-center w-full h-full'>
            <div className="pl-6 p-4 w-full flex flex-col 980:flex-row items-center justify-center">
              <div
                onClick={handlePlay}
                className="flex rounded-3xl pl-0 980:pl-10 mt-20 380:-mt-10 md:mt-0 flex-col h-full justify-center cursor-pointer w-fit "
                style={{
                  minWidth: 0,
                  width: 'fit-content',
                  display: 'flex',
                  flexDirection: 'column',
                  alignItems: 'flex-start',
                  transition: 'transform 0.2s',
                  transformOrigin: 'left',
                }}
                ref={scaleRef}
                data-libr-scale-listener
              >
                <span className="text-libr-secondary/20 text-3xl translate-y-16">свобода</span>
                <span className="text-libr-secondary/30 text-4xl translate-y-16 tracking-wider">स्वतंत्रता</span>
                <span className="text-libr-secondary/40 text-5xl translate-y-14">Liberté</span>
                <div className='flex flex-row items-center -translate-x-2'>
                  <span className="text-libr-secondary text-11xl ">libr</span>
                  <audio ref={audioRef} src="../src/assets/libr.mp3" preload="auto" />
                </div>
                <span className="text-libr-secondary/40 text-4xl -translate-y-16 tracking-wider">স্বাধীনতা</span>
                <span className="text-libr-secondary/30 text-3xl -translate-y-17">Libertad</span>
                <span className="text-libr-secondary/20 text-2xl -translate-y-18">స్వేచ్ఛ</span>
              </div>
              <div className="flex flex-row justify-start mb-4 items-center md:justify-end md:p-8 w-full min-w-0">
                <p
                  className="text-muted-foreground opacity-50 md:blur-sm whitespace-nowrap text-[clamp(2rem,8vw,8rem)] pr-0 980:pr-10 text-center 980:text-left"
                  style={{
                    lineHeight: 1.1,
                    textAlign: 'left',
                  }}
                >
                  Your Space.<br/>
                  Your Quorum.<br/>
                  Your Rules.
                </p>
              </div>
            </div>
            <div id="join-beta" className="flex flex-col gap-4 w-full items-center justify-center sm:flex-row sm:gap-4 sm:items-center sm:justify-center">
              <JoinBetaDropdown />
              {/* <button onClick={() => window.open("https://forms.gle/udt5zATFogCGQtUTA", '_blank')} className="libr-button bg-libr-secondary text-libr-primary flex flex-row items-center w-full max-w-xs mx-auto sm:w-full sm:max-w-xs sm:mx-auto md:w-auto md:max-w-none md:mx-0">
                <Download className="w-5 h-5 mr-3" />
                Join Beta
              </button> */}
              <button onClick={() => window.open("https://medium.com/@libr.forum/libr-a-moderated-censorship-resilient-social-network-framework-ecfcffb3fdae", '_blank')} className="flex flex-row items-center libr-button-secondary text-libr-secondary border-xl border-libr-secondary w-full max-w-xs mx-auto sm:w-full sm:max-w-xs sm:mx-auto md:w-auto md:max-w-none md:mx-0">
                <Users className="w-5 h-5 mr-3" />
                View Docs
              </button>
            </div>
          </div>
        </motion.div>
      </div>
    </section>
  );
};

const WhatIsLIBR: React.FC = () => {
  const [isMobile, setIsMobile] = useState(window.innerWidth < 1024);

  useEffect(() => {
    const handleResize = () => setIsMobile(window.innerWidth < 1024);
    window.addEventListener("resize", handleResize);
    return () => window.removeEventListener("resize", handleResize);
  }, []);

  return (
    <section id="what-is-libr" className="flex items-center pt-20 pb-20">
      <div className="container mx-auto">
        <div
          className={`flex ${
            isMobile ? "flex-col" : "flex-row"
          } gap-12 items-center`}
        >
          {/* Left Column */}
          <motion.div
            initial={
              isMobile
                ? { y: 100, opacity: 0 }
                : { x: -100, opacity: 0 }
            }
            whileInView={{ x: 0, y: 0, opacity: 1 }}
            transition={{
              duration: 0.8,
              ease: [0.4, 0, 0.2, 1],
            }}
            viewport={{ once: false }}
          >
            <h2 className="text-4xl lg:text-5xl font-bold text-libr-secondary mb-6 w-full">
              Do we have the freedom of speech?
            </h2>
            <p className="text-xl text-muted-foreground mb-8">
              Tired of platforms that quietly delete your posts?
              <br />
              Or rules that change depending on who’s watching?
              <br />
              libr is a new kind of social platform.
              <br />
              Built on transparency, community, and proof.
            </p>
            <p className="text-md mb-8">
              libr is a <b>censorship-resilient yet moderated</b> forum
              protocol where communities set their own rules — and every
              moderation decision is{" "}
              <b>cryptographically verifiable</b>.
            </p>
            <div className="flex flex-col gap-4">
              <div className="flex flew-row gap-2">
                <Shield /> No Shadow Bans
              </div>
              <div className="flex flew-row gap-2">
                <Users /> Moderation Per Community Rules
              </div>
              <div className="flex flew-row gap-2">
                <VenetianMask /> Pseudonomity
              </div>
            </div>
          </motion.div>

          {/* <motion.div
            initial={
              isMobile
                ? { y: 100, opacity: 0 }
                : { x: 100, opacity: 0 }
            }
            whileInView={{ x: 0, y: 0, opacity: 1 }}
            transition={{
              duration: 0.8,
              ease: [0.4, 0, 0.2, 1],
            }}
            viewport={{ once: false }}
            className="space-y-6 w-full"
          >
            {[
              {
                text: `"The hybrid approach of combining DHTs with Byzantine consensus is innovative. 
                This research addresses real challenges in decentralized systems."`,
                initial: "A",
                name: "Dr. Alice Research",
                role: "Distributed Systems",
                color: "bg-libr-accent1",
              },
              {
                text: `"Impressive protocol design. The modular Go implementation makes it easy to understand and extend."`,
                initial: "S",
                name: "Sam Developer",
                role: "Open Source Contributor",
                color: "bg-libr-accent2",
              },
            ].map((t, idx) => (
              <div className="testimonial-card" key={idx}>
                <p className="text-muted-foreground mb-4">{t.text}</p>
                <div className="flex items-center justify-center gap-3">
                  <div
                    className={`w-10 h-10 ${t.color} rounded-full flex items-center justify-center text-white font-semibold`}
                  >
                    {t.initial}
                  </div>
                  <div>
                    <p className="font-semibold">{t.name}</p>
                    <p className="text-sm text-muted-foreground">
                      {t.role}
                    </p>
                  </div>
                </div>
              </div>
            ))}
          </motion.div> */}
        </div>
      </div>
    </section>
  );
};

const TechArch: React.FC = () => {
  

  const features = [
    {
      icon: Shield,
      title: "Censorship resilient",
      description: "Built on DHTs for immutable message storage with partial immutability for efficient forum operations."
    },
    {
      icon: Users,
      title: "Community Moderated",
      description: "Byzantine Consistent Broadcast ensures democratic moderation quorums with 2f+1 moderator consensus for content validation."
    },
    {
      icon: Lock,
      title: "Cryptographically Secure",
      description: "Digital signatures with ed25519 keys used at each stage ensuring end-to-end immutability."
    },
    {
      icon: Waypoints,
      title: "Modern Web Net Infra",
      description: "Websockets based p2p network with fallback mechanisms and support over VPNs."
    },
    {
      icon: DatabaseZap,
      title: "Replicated DHT",
      description: "Distributed hash table with replication ensures permanent data availability."
    },
    {
      icon: Globe,
      title: "Decentralized Architecture",
      description: "No central servers, ever.\nRelays and databases also run on community nodes."
    }
  ];

  return (
    <section id="features" className="py-20 section-padding">
      <div className="container mx-auto">
        <motion.div 
          className="text-center mb-16"
          initial={{ y: 50, opacity: 0 }}
          whileInView={{ y: 0, opacity: 1 }}
          transition={{
            duration: 0.8,
            ease: [0.4, 0, 0.2, 1],
          }}
          viewport={{ once: false }}
        >
          <h2 className="text-4xl lg:text-5xl font-bold text-libr-secondary mb-4">
            Technical Architecture
          </h2>
          <p className="text-xl text-muted-foreground max-w-3xl mx-auto">
            A Moderated, Censorship-Resilient Social Network Framework
          </p>
        </motion.div>
        
        <div className="grid md:grid-cols-2 lg:grid-cols-3 gap-8">
          {features.map((feature, index) => (
            <motion.div
              key={feature.title}
              className="feature-card"
              initial={isMobile?{ y: 50, opacity: 1 }:{ y: 50, opacity: 0 }}
              whileInView={{ y: 0, opacity: 1 }}
              transition={{
                ease: [0.4, 0, 0.2, 1],
                delay: index * 0.1
              }}
              viewport={{ once: false }}
            >
              <div className='flex flex-row items-center space-x-2'>
              <div className="w-12 h-12 bg-libr-secondary rounded-lg flex items-center justify-center mb-4">
                <feature.icon className="w-6 h-6 text-libr-primary" />
              </div>
              <h3 className="text-xl font-semibold text-libr-secondary mb-2">{feature.title}</h3>
              </div>
              <p className="text-muted-foreground">{feature.description}</p>
            </motion.div>
          ))}
        </div>
      </div>
    </section>
  );
};

export { Header, Hero, TechArch, WhatIsLIBR };
