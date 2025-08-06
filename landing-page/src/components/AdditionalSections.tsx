import React from 'react';
import { motion } from 'framer-motion';
import { Shield, Zap, Globe, Users, Lock, Code, ShieldCheck, Monitor, Database,KeyRound } from 'lucide-react';

const TechModules: React.FC = () => {
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
    <section id="technical-modules"className="py-20 section-padding">
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
          <div className='flex flex-row items-center justify-center w-full space-x-6'>
            {techStack1.map((stack, index) => (
              <motion.div
                key={stack.category}
                className="feature-card p-6 w-[20%] h-full text-center"
                initial={{ y: 50, opacity: 0 }}
                whileInView={{ y: 0, opacity: 1 }}
                transition={{
                  duration: 0.8,
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
          <div className='flex flex-row items-center justify-center w-full space-x-6'>
            {techStack2.map((stack, index) => (
              <motion.div
                key={stack.category}
                className="feature-card p-6 w-[20%] h-full text-center"
                initial={{ y: 50, opacity: 0 }}
                whileInView={{ y: 0, opacity: 1 }}
                transition={{
                  duration: 0.8,
                  ease: [0.4, 0, 0.2, 1],
                  delay: (index * 0.1)+ 0.2,
                }}
                viewport={{ once: false }}
              >
                <div className="w-12 h-12 bg-libr-secondary rounded-lg flex items-center justify-center mx-auto mb-4">
                  <stack.icon className="w-6 h-6 text-libr-primary" />
                </div>
                <h3 className="text-lg font-semibold text-libr-secondary mb-3">{stack.category}</h3>
                <div className="space-y-2">
                  {stack.technologies.map((tech) => (
                    <div key={tech} className="text-sm text-muted-foreground bg-muted/30 rounded-full px-3 py-1">
                      {tech}
                    </div>
                  ))}
                </div>
              </motion.div>
            ))}
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
                    <feature.icon className="w-5 h-5 text-white" />
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
                  <feature.icon className="w-5 h-5 text-white" />
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
            <button onClick={() => window.open('https://github.com/devlup-labs/Libr/blob/main/README.md', '_blank')} className="flex flex-row items-center libr-button-primary text-lg">
              <Users className="w-5 h-5 mr-3" />
              View Documentation
            </button>
            <button onClick={() => window.open('https://github.com/devlup-labs/Libr', '_blank')}className="flex flex-row items-center libr-button-secondary text-lg">
              <Code className="w-5 h-5 mr-3" />
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

export { TechModules, SecuritySection, CallToActionSection };
