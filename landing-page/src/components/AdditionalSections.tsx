import React,{useState,useEffect} from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import { Shield, Globe, Users, Lock, Code, ShieldCheck, Monitor, Database,KeyRound, Download, Play } from 'lucide-react';
import { FaApple, FaLinux, FaTools, FaWindows } from 'react-icons/fa';
const isMobile = typeof window !== "undefined" ? window.innerWidth < 768 : false;
const TechModules: React.FC = () => {
  const [marginTop, setMarginTop] = useState(0);
  
  // Realtime resize handling
  useEffect(() => {
    const handleResize = () => {
      if (window.innerWidth < 400) {
        setMarginTop(100);
      } else if (window.innerWidth < 768) {
        setMarginTop(-100);
      } else if (window.innerWidth < 1200) {
        // Linear interpolation from -200 at 768px to 800 at 1199px
        const ratio = (window.innerWidth - 768) / (1199 - 768); // 0 → 1
        const margin = -200 + ratio * (800 - -200); // -200 to 800
        setMarginTop(margin);
      } else if (window.innerWidth === 1200) {
        setMarginTop(-600);
      } else if (window.innerWidth > 1200 && window.innerWidth < 1750) {
        // Linear interpolation from -200 at 1201px to -100 at 1750px
        const ratio = (window.innerWidth - 1201) / (1750 - 1201); // 0 → 1
        const margin = -200 + ratio * 100; // -200 to -100
        setMarginTop(margin);
      } else {
        setMarginTop(-200);
      }
    };
    handleResize(); // Run once on mount
    window.addEventListener("resize", handleResize);
    return () => window.removeEventListener("resize", handleResize);
  }, []);

  const techStack1 = [
    {
      category: "Client",
      icon: Monitor,
      technologies: ["Go+React", "Wails Integration", "Pseudonymous"]
    },
    {
      category: "Moderator",
      icon: ShieldCheck,
      technologies: ["Community Ruled", "Pluggable", "Google Cloud NLP"]
    },
    {
      category: "Network",
      icon: Globe,
      technologies: ["libp2p", "WebSockets", "Relayed Connections"]
    }    
  ];

  const techStack2=[
    {
      category: "Database",
      icon: Database,
      technologies: ["SQLite 3", "Kademlia", "Replicated DHT"]
    },
    {
      category: "Crypto",
      icon: KeyRound,
      technologies: ["Digital Signatures", "ed25519 Key-Pair", "SHA256 Hashing"]
    }  ];

  return (
    <section id="technical-modules"className="py-20 section-padding"  style={{ marginTop: `${marginTop}px`}}>
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
            Technical Modules
          </h2>
          {/* <p className="text-xl text-muted-foreground max-w-3xl mx-auto">
            LIBR leverages proven distributed systems technologies and novel consensus mechanisms 
            to achieve both censorship resistance and effective community moderation.
          </p> */}
        </motion.div>
        
        <div className="flex flex-col items-center justify-center w-full space-y-6">
          {/* Responsive layout: <768px = 1 column, 768-1299px = 2 rows (3+2), >=1300px = 1 row of 5 */}
          <div className="w-full">
            {/* <1024px: 1 column */}
            <div className="flex flex-col lg:hidden w-full space-y-6">
              {[...techStack1, ...techStack2].map((stack, index) => (
                <motion.div
                  key={stack.category}
                  className="feature-card p-6 w-full h-full text-center"
                  initial={{ y: 50, opacity: 1 }}
                  whileInView={{ y: 0, opacity: 1 }}
                  transition={{
                    ease: [0.4, 0, 0.2, 1],
                    delay: index * 0.1,
                  }}
                  viewport={{ once: false }}
                >
                  <div className="w-12 h-12 bg-libr-secondary rounded-lg flex items-center justify-center mx-auto mb-4">
                    <stack.icon className="w-6 h-6 text-libr-primary" />
                  </div>
                  <h3 className="text-lg font-semibold text-libr-secondary mb-3">{stack.category}</h3>
                  <div className="space-y-2">
                    {stack.technologies.map((tech) => (
                      <div key={tech} className="text-sm text-muted-foreground bg-muted/30 rounded-full px-3 py-2">
                        {tech}
                      </div>
                    ))}
                  </div>
                </motion.div>
              ))}
            </div>

            {/* 1024px-1299px: 2 rows (3+2) */}
            <div className="hidden lg:flex 2xl:hidden flex-col w-full space-y-6">
              <div className="flex flex-row items-center justify-center w-full space-x-6">
                {techStack1.map((stack, index) => (
                  <motion.div
                    key={stack.category}
                    className="feature-card p-6 w-full md:w-[30%] h-full text-center"
                    initial={{ y: 50, opacity: 0 }}
                    whileInView={{ y: 0, opacity: 1 }}
                    transition={{
                      ease: [0.4, 0, 0.2, 1],
                      delay: index * 0.1,
                    }}
                    viewport={{ once: false }}
                  >
                    <div className="w-12 h-12 bg-libr-secondary rounded-lg flex items-center justify-center mx-auto mb-4">
                      <stack.icon className="w-6 h-6 text-libr-primary" />
                    </div>
                    <h3 className="text-lg font-semibold text-libr-secondary mb-3">{stack.category}</h3>
                    <div className="space-y-2">
                      {stack.technologies.map((tech) => (
                        <div key={tech} className="text-sm text-muted-foreground bg-muted/30 rounded-full px-3 py-2">
                          {tech}
                        </div>
                      ))}
                    </div>
                  </motion.div>
                ))}
              </div>
              <div className="flex flex-row items-center justify-center w-full space-x-6">
                {techStack2.map((stack, index) => (
                  <motion.div
                    key={stack.category}
                    className="feature-card p-6 w-full md:w-[30%] h-full text-center"
                    initial={{ y: 50, opacity: 0 }}
                    whileInView={{ y: 0, opacity: 1 }}
                    transition={{
                      ease: [0.4, 0, 0.2, 1],
                      delay: (index + 3) * 0.1,
                    }}
                    viewport={{ once: false }}
                  >
                    <div className="w-12 h-12 bg-libr-secondary rounded-lg flex items-center justify-center mx-auto mb-4">
                      <stack.icon className="w-6 h-6 text-libr-primary" />
                    </div>
                    <h3 className="text-lg font-semibold text-libr-secondary mb-3">{stack.category}</h3>
                    <div className="space-y-2">
                      {stack.technologies.map((tech) => (
                        <div key={tech} className="text-sm text-muted-foreground bg-muted/30 rounded-full px-3 py-2">
                          {tech}
                        </div>
                      ))}
                    </div>
                  </motion.div>
                ))}
              </div>
            </div>

            {/* >=1300px: 1 row of 5 */}
            <div className="hidden 2xl:flex flex-row items-center justify-center w-full space-x-6">
              {[...techStack1, ...techStack2].map((stack, index) => (
                <motion.div
                  key={stack.category}
                  className="feature-card p-6 w-full lg:w-[20%] h-full text-center"
                  initial={{ y: 50, opacity: 0 }}
                  whileInView={{ y: 0, opacity: 1 }}
                  transition={{
                    ease: [0.4, 0, 0.2, 1],
                    delay: index * 0.1,
                  }}
                  viewport={{ once: false }}
                >
                  <div className="w-12 h-12 bg-libr-secondary rounded-lg flex items-center justify-center mx-auto mb-4">
                    <stack.icon className="w-6 h-6 text-libr-primary" />
                  </div>
                  <h3 className="text-lg font-semibold text-libr-secondary mb-3">{stack.category}</h3>
                  <div className="space-y-2">
                    {stack.technologies.map((tech) => (
                      <div key={tech} className="text-sm text-muted-foreground bg-muted/30 rounded-full px-3 py-2">
                        {tech}
                      </div>
                    ))}
                  </div>
                </motion.div>
              ))}
            </div>
          </div>
        </div>
      </div>
    </section>
  );
};

const SecuritySection: React.FC = () => {
  const securityFeatures = [
    {
      title: "Message Certificates",
      description: "Each message requires cryptographic validation from 2M+1 moderators via Byzantine Consistent Broadcast.",
      icon: Lock
    },
    {
      title: "Replicated Storage",
      description: "Messages stored across R database nodes using DHT for fault tolerance and partial immutability.",
      icon: Globe
    },
    {
      title: "Proof of Service",
      description: "Database nodes provide verifiable service proofs while moderators maintain scoring systems.",
      icon: Shield
    },
    {
      title: "Open Research",
      description: "Academic transparency with full protocol specification and implementation details available.",
      icon: Code
    }
  ];

  return (
    <section className="py-20 section-padding">
      <div className="container mx-auto">
        <div className="grid lg:grid-cols-2 gap-12 items-center">
          <motion.div
            initial={{ x: -100, opacity: 0 }}
            whileInView={{ x: 0, opacity: 1 }}
            transition={{
              duration: 0.8,
              ease: [0.4, 0, 0.2, 1],
            }}
            viewport={{ once: false }}
          >
            <h2 className="text-4xl lg:text-5xl font-bold text-libr-secondary mb-6">
              Protocol Security
            </h2>
            <p className="text-xl text-muted-foreground mb-8">
              LIBR implements robust security measures through cryptographic protocols, 
              distributed consensus, and transparent governance mechanisms.
            </p>
            
            <div className="space-y-6">
              {securityFeatures.slice(0, 2).map((feature, index) => (
                <motion.div
                  key={feature.title}
                  className="flex gap-4"
                  initial={{ y: 30, opacity: 0 }}
                  whileInView={{ y: 0, opacity: 1 }}
                  transition={{
                    duration: 0.8,
                    ease: [0.4, 0, 0.2, 1],
                    delay: index * 0.2
                  }}
                  viewport={{ once: false }}
                >
                  <div className="w-10 h-10 bg-libr-accent1 rounded-lg flex items-center justify-center flex-shrink-0">
                    <feature.icon className="w-6 h-6 text-white" />
                  </div>
                  <div>
                    <h3 className="font-semibold text-libr-secondary mb-1">{feature.title}</h3>
                    <p className="text-muted-foreground text-sm">{feature.description}</p>
                  </div>
                </motion.div>
              ))}
            </div>
          </motion.div>
          
          <motion.div
            initial={{ x: 100, opacity: 0 }}
            whileInView={{ x: 0, opacity: 1 }}
            transition={{
              duration: 0.8,
              ease: [0.4, 0, 0.2, 1],
            }}
            viewport={{ once: false }}
            className="space-y-6"
          >
            {securityFeatures.slice(2).map((feature, index) => (
              <motion.div
                key={feature.title}
                className="flex gap-4"
                initial={{ y: 30, opacity: 0 }}
                whileInView={{ y: 0, opacity: 1 }}
                transition={{
                  duration: 0.8,
                  ease: [0.4, 0, 0.2, 1],
                  delay: (index+2) * 0.2
                }}
                viewport={{ once: false }}
              >
                <div className="w-10 h-10 bg-libr-accent2 rounded-lg flex items-center justify-center flex-shrink-0">
                  <feature.icon className="w-6 h-6 text-white" />
                </div>
                <div>
                  <h3 className="font-semibold text-libr-secondary mb-1">{feature.title}</h3>
                  <p className="text-muted-foreground text-sm">{feature.description}</p>
                </div>
              </motion.div>
            ))}
            
            <motion.div 
              className="libr-card p-6 bg-gradient-to-br from-libr-accent1/10 to-libr-accent2/10"
              initial={{ scale: 0.9, opacity: 0 }}
              whileInView={{ scale: 1, opacity: 1 }}
              transition={{
                duration: 0.8,
                ease: [0.4, 0, 0.2, 1],
                delay: 0.8
              }}
              viewport={{ once: false }}
            >
              <div className="flex items-center gap-3 mb-3">
                <div className="w-8 h-8 bg-green-500 rounded-full flex items-center justify-center">
                  ✓
                </div>
                <span className="font-semibold text-libr-secondary">Security Audit Complete</span>
              </div>
              <p className="text-sm text-muted-foreground">
                Libr has undergone comprehensive security audits by independent security firms to ensure the highest standards of protection.
              </p>
            </motion.div>
          </motion.div>
        </div>
      </div>
    </section>
  );
};

const CallToActionSection: React.FC = () => {
  return (
    <section className="py-20 section-padding">
      <div className="container mx-auto">
        <motion.div
          className="libr-card p-12 text-center bg-gradient-to-br from-libr-accent1/5 to-libr-accent2/5 border-2 border-libr-accent1/20"
          initial={{ scale: 0.9, opacity: 0 }}
          whileInView={{ scale: 1, opacity: 1 }}
          transition={{
            duration: 0.8,
            ease: [0.4, 0, 0.2, 1],
          }}
          viewport={{ once: false }}
        >
          <motion.h2 
            className="text-4xl lg:text-5xl font-bold text-libr-secondary mb-6"
            initial={{ y: 30, opacity: 0 }}
            whileInView={{ y: 0, opacity: 1 }}
            transition={{ duration: 0.6, delay: 0.2 }}
            viewport={{ once: true }}
          >
            Explore the LIBR Protocol
          </motion.h2>
          
          <motion.p 
            className="text-xl text-muted-foreground mb-8 max-w-2xl mx-auto"
            initial={{ y: 30, opacity: 0 }}
            whileInView={{ y: 0, opacity: 1 }}
            transition={{ duration: 0.6, delay: 0.4 }}
            viewport={{ once: true }}
          >
            Dive into the research, examine the implementation, and contribute to the future 
            of censorship-resilient yet moderated communication systems.
          </motion.p>
          
          <motion.div 
            className="flex sm:flex-row gap-4 justify-center items-center"
            initial={{ y: 30, opacity: 0 }}
            whileInView={{ y: 0, opacity: 1 }}
            transition={{
              duration: 0.8,
              ease: [0.4, 0, 0.2, 1],
              delay: 0.6
            }}
            viewport={{ once: false }}
          >
            <button onClick={() => window.open('https://github.com/libr-forum/libr/blob/main/README.md', '_blank')} className="flex flex-row items-center libr-button-primary text-lg">
              <Users className="w-6 h-6 mr-3" />
              View Documentation
            </button>
            <button onClick={() => window.open('https://github.com/libr-forum/libr', '_blank')}className="flex flex-row items-center libr-button-secondary text-lg">
              <Code className="w-6 h-6 mr-3" />
              Explore Code
            </button>
          </motion.div>
          
          <motion.div 
            className="mt-8 text-sm text-muted-foreground"
            initial={{ opacity: 0 }}
            whileInView={{ opacity: 1 }}
            transition={{
              duration: 0.8,
              ease: [0.4, 0, 0.2, 1],
            }}
            viewport={{ once: false }}
          >
            <p>Open Research • Academic Project • MIT Licensed</p>
          </motion.div>
        </motion.div>
      </div>
    </section>
  );
};

const HowToUse: React.FC = () => {
  const steps = [
    {
      icon: <Download className="w-6 h-6 text-libr-primary" />,
      title: "Download & Install",
      description: "Get LIBR for your platform and install it in a few clicks.",
    },
    {
      icon: <Play className="w-6 h-6 text-libr-primary" />,
      title: "Run & Start Posting",
      description: "Run the application and start sharing your thoughts.",
    },
    {
      icon: <Database className="w-6 h-6 text-libr-primary" />,
      title: "Host Your Database",
      description: "If you like, host a database node and contribute to the network.",
    },
  ];

  return (
    <section id='how-to-use' className="py-20 section-padding">
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
            How to Use
          </h2>
        </motion.div>

        <div className="flex flex-col md:flex-row items-center justify-center w-full space-y-6 md:space-y-0 md:space-x-6">
          {steps.map((step, index) => {
            const isFirst = index === 0;
            return (
              <motion.div
                key={index}
                className={`feature-card p-6 w-full md:w-64 text-center ${isFirst ? "cursor-pointer" : ""} md:aspect-square`}
                initial={isMobile?{ y: 50, opacity: 1 }:{ y: 50, opacity: 0 }}
                whileInView={{ y: 0, opacity: 1 }}
                transition={{
                  ease: [0.4, 0, 0.2, 1],
                  delay: index * 0.1,
                }}
                viewport={{ once: false }}
                onClick={() => {
                  if (isFirst) {
                    window.location.href = "#welcome";
                  }
                }}
              >
                <div className="w-12 h-12 bg-libr-secondary rounded-lg flex items-center justify-center mx-auto mb-4">
                  {step.icon}
                </div>
                <h3 className="text-lg font-semibold text-libr-secondary mb-3">
                  {step.title}
                </h3>
                <p className="text-muted-foreground">{step.description}</p>
              </motion.div>
            );
          })}
        </div>
      </div>
    </section>
  );
};

const CollapsibleSection: React.FC<{
  title: string;
  icon: React.ReactNode;
  children: React.ReactNode;
  defaultOpen?: boolean;
}> = ({ title, icon, children, defaultOpen = false }) => {
  const [open, setOpen] = useState(defaultOpen);
  return (
    <motion.div
      className="libr-card p-6 mb-6 bg-libr-card rounded-lg shadow"
      initial={{ y: 20, opacity: 0.8, scale: 0.98 }}
      whileInView={{ y: 0, opacity: 1, scale: 1 }}
      transition={{ duration: 0.7, ease: [0.4, 0, 0.2, 1] }}
      viewport={{ once: false }}
    >
      <button
        className="w-full flex items-center justify-between text-left focus:outline-none"
        onClick={() => setOpen((o) => !o)}
      >
        <span className="flex items-center gap-2 text-2xl font-bold text-libr-secondary">
          <span className="w-10 h-10 bg-libr-secondary rounded-lg flex items-center justify-center">
            {/* Icon with libr-primary color */}
            {React.isValidElement(icon)
              ? React.cloneElement(icon as React.ReactElement, { className: "w-6 h-6 text-libr-primary" })
              : icon}
          </span>
          {title}
        </span>
        <span className="ml-2 text-libr-accent1">{open ? '−' : '+'}</span>
      </button>
      <AnimatePresence>
        {open && (
          <motion.div
            initial={{ height: 0, opacity: 0, y: 10 }}
            animate={{ height: 'auto', opacity: 1, y: 0 }}
            exit={{ height: 0, opacity: 0, y: 10 }}
            transition={{ duration: 0.4 }}
            className="mt-4"
          >
            {children}
          </motion.div>
        )}
      </AnimatePresence>
    </motion.div>
  );
};

const InstallGuide: React.FC = () => {
  return (
    <section id='installation-guide' className='py-20 section-padding'>
      <div className="container mx-auto max-w-5xl">
        <motion.div
          className="text-center mb-12"
          initial={{ y: 50, opacity: 0 }}
          whileInView={{ y: 0, opacity: 1 }}
          transition={{ duration: 0.8, ease: [0.4, 0, 0.2, 1] }}
          viewport={{ once: false }}
        >
          <h2 className="text-4xl lg:text-5xl font-bold text-libr-secondary mb-4">
            Installation Guide
          </h2>
        </motion.div>

        <CollapsibleSection
          title="Linux Installation"
          icon={<div className='bg-libr-secondary p-1 rounded-lg'><FaLinux/></div>}
          defaultOpen={false}
        >
          <div className="mb-4">
            <span className="font-semibold text-libr-accent1">Option 1: APT Repository (Ubuntu/Debian) - Recommended</span>
            <pre className="bg-muted/30 rounded-lg p-4 text-sm text-left mt-2 overflow-x-auto">
{`wget -qO- https://libr-forum.github.io/libr-apt-repo/libr-repo-key.gpg | sudo gpg --dearmor -o /usr/share/keyrings/libr-repo-key.gpg
echo "deb [signed-by=/usr/share/keyrings/libr-repo-key.gpg] https://libr-forum.github.io/libr-apt-repo/ ./" | sudo tee /etc/apt/sources.list.d/libr.list
sudo apt update
sudo apt install libr`}
            </pre>
          </div>
          <div className="mb-4">
            <span className="font-semibold text-libr-accent1">Option 2: Direct Download - All Distributions</span>
            <div className="mt-2">
              <span className="font-semibold">Ubuntu/Debian (.deb package):</span>
              <pre className="bg-muted/30 rounded-lg p-4 text-sm text-left mt-2 overflow-x-auto">
{`wget https://github.com/libr-forum/Libr/releases/download/v1.0.0-beta/libr_1.0.0-beta_amd64.deb
sudo dpkg -i libr_1.0.0-beta_amd64.deb
sudo apt-get install -f`}
              </pre>
              <span className="font-semibold">Fedora/RHEL/CentOS (.rpm package):</span>
              <pre className="bg-muted/30 rounded-lg p-4 text-sm text-left mt-2 overflow-x-auto">
{`wget https://github.com/libr-forum/Libr/releases/download/v1.0.0-beta/libr-1.0.0-beta-1.x86_64.rpm
sudo dnf install ./libr-1.0.0-beta-1.x86_64.rpm
# or on older systems: sudo yum install ./libr-1.0.0-beta-1.x86_64.rpm`}
              </pre>
              <span className="font-semibold">Arch Linux (.pkg.tar.zst package):</span>
              <pre className="bg-muted/30 rounded-lg p-4 text-sm text-left mt-2 overflow-x-auto">
{`wget https://github.com/libr-forum/Libr/releases/download/v1.0.0-beta/libr-1.0.0-beta-1-x86_64.pkg.tar.zst
sudo pacman -U libr-1.0.0-beta-1-x86_64.pkg.tar.zst`}
              </pre>
            </div>
          </div>
          <div className="mb-4">
            <span className="font-semibold text-libr-accent1">Option 3: Binary Installation</span>
            <pre className="bg-muted/30 rounded-lg p-4 text-sm text-left mt-2 overflow-x-auto">
{`wget https://github.com/libr-forum/Libr/releases/download/v1.0.0-beta/libr-linux-amd64
chmod +x libr-linux-amd64
sudo mv libr-linux-amd64 /usr/local/bin/libr`}
            </pre>
          </div>
          <div className="mb-4">
            <span className="font-semibold text-libr-accent1">Solving WebKit Library Issues</span>
            <p className="text-muted-foreground text-sm mt-2">Libr uses WebKitGTK for its UI, which may require specific library versions on different distributions:</p>
            <pre className="bg-muted/30 rounded-lg p-4 text-xs text-left mt-2 overflow-x-auto">
{`# Ubuntu 24.04 (Noble) and newer:
sudo apt update
sudo apt install -y libwebkit2gtk-4.1-0 libjavascriptcoregtk-4.1-0
sudo ln -sf /usr/lib/x86_64-linux-gnu/libwebkit2gtk-4.1.so.0 /usr/lib/x86_64-linux-gnu/libwebkit2gtk-4.0.so.37
sudo ln -sf /usr/lib/x86_64-linux-gnu/libjavascriptcoregtk-4.1.so.0 /usr/lib/x86_64-linux-gnu/libjavascriptcoregtk-4.0.so.18

# Fedora 35+ and newer RHEL/CentOS:
sudo dnf install webkit2gtk4.1-devel
sudo ln -sf /usr/lib64/libwebkit2gtk-4.1.so.0 /usr/lib64/libwebkit2gtk-4.0.so.37
sudo ln -sf /usr/lib64/libjavascriptcoregtk-4.1.so.0 /usr/lib64/libjavascriptcoregtk-4.0.so.18

# Arch Linux:
sudo pacman -S webkit2gtk-4.1
sudo ln -sf /usr/lib/libwebkit2gtk-4.1.so.0 /usr/lib/libwebkit2gtk-4.0.so.37
sudo ln -sf /usr/lib/libjavascriptcoregtk-4.1.so.0 /usr/lib/libjavascriptcoregtk-4.0.so.18

# Generic Linux (if above don't work):
sudo apt install libwebkit2gtk-4.0-dev
sudo dnf install webkit2gtk3-devel
sudo zypper install webkit2gtk3-devel
sudo pacman -S webkit2gtk`}
            </pre>
          </div>
          <div className="mb-4">
            <span className="font-semibold text-libr-accent1">Alternative: Automated Installation Script</span>
            <pre className="bg-muted/30 rounded-lg p-4 text-sm text-left mt-2 overflow-x-auto">
{`curl -fsSL https://raw.githubusercontent.com/libr-forum/Libr/main/scripts/install-libr.sh | bash
# Or inspect the script first (recommended for security)
wget https://raw.githubusercontent.com/libr-forum/Libr/main/scripts/install-libr.sh
chmod +x install-libr.sh
./install-libr.sh`}
            </pre>
            <p className="text-muted-foreground text-xs mt-2">
              The script automatically detects your distribution and handles package installation and library dependencies.
            </p>
          </div>
        </CollapsibleSection>

        <CollapsibleSection
          title="Windows Installation"
          icon={<span><FaWindows/></span>}
        >
          <pre className="bg-muted/30 rounded-lg p-4 text-sm text-left mt-2 overflow-x-auto">
{`Download the latest Windows release (libr-win-amd64.exe) from the Releases page
Double-click to run it
If the app doesn't start, try right-click → Run as administrator`}
          </pre>
        </CollapsibleSection>

        <CollapsibleSection
          title="macOS Installation"
          icon={<span><FaApple/></span>}
        >
          <pre className="bg-muted/30 rounded-lg p-4 text-sm text-left mt-2 overflow-x-auto">
{`Download the macOS release (libr-darwin-amd64.out) from Releases
On first run, macOS may block the app. To fix this:
Go to System Settings → Privacy & Security
Allow the app under the "Security" section
Make executable and run:
chmod +x ./libr-darwin-amd64.out
./libr-darwin-amd64.out`}
          </pre>
          <h3 className="text-2xl font-bold text-libr-secondary mb-4 flex items-center gap-2 mt-8">
            📋 Verification
          </h3>
          <pre className="bg-muted/30 rounded-lg p-4 text-sm text-left mt-2 overflow-x-auto">
{`libr --version
libr`}
          </pre>
          <p className="text-muted-foreground text-xs mt-2">
            After installation, you can find Libr in your applications menu under "Network" → "Libr" or run <code>libr</code> from any terminal.
          </p>
        </CollapsibleSection>

        <CollapsibleSection
          title="Troubleshooting"
          icon={<span><FaTools/></span>}
        >
          <ul className="list-disc pl-6 text-left text-muted-foreground space-y-2">
            <li>
              <strong>Command not found:</strong> Make sure the binary is in your PATH or use the full path to the executable.
            </li>
            <li>
              <strong>Permission denied:</strong> Run <code>chmod +x</code> on the downloaded binary.
            </li>
            <li>
              <strong>WebKit library errors on Linux:</strong> Install the WebKitGTK packages for your distribution as shown above. The error typically looks like: <code>libwebkit2gtk-4.0.so.37: cannot open shared object file</code>. The symlink solutions above resolve compatibility issues.
            </li>
            <li>
              <strong>Package installation fails:</strong> On Debian/Ubuntu: Run <code>sudo apt-get install -f</code> to fix dependencies. On Fedora/RHEL: Ensure EPEL repository is enabled for additional packages. On Arch Linux: Update system with <code>sudo pacman -Syu</code> before installing.
            </li>
            <li>
              <strong>macOS security warnings:</strong> Allow the app in System Settings → Privacy & Security.
            </li>
            <li>
              <strong>Distribution-specific Notes:</strong>
              <ul className="list-disc pl-6">
                <li>Ubuntu 24.04+ (Noble): Requires WebKit 4.1 packages and compatibility symlinks.</li>
                <li>Fedora 35+: May need <code>webkit2gtk4.1-devel</code> package.</li>
                <li>Arch Linux: Install <code>webkit2gtk-4.1</code> or <code>webkit2gtk</code> packages.</li>
                <li>RHEL/CentOS 8+: Enable PowerTools/CRB repository for WebKit packages.</li>
              </ul>
            </li>
          </ul>
        </CollapsibleSection>
      </div>
    </section>
  );
};

export { TechModules, SecuritySection, CallToActionSection, HowToUse, InstallGuide };
