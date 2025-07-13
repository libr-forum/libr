import React from 'react';
import { motion } from 'framer-motion';
import { Shield, Zap, Globe, Users, Lock, Code } from 'lucide-react';

const TechStackSection: React.FC = () => {
  const techStack = [
    {
      category: "Core Protocol",
      icon: Code,
      technologies: ["Go Language", "Docker", "PostgreSQL", "Cobra CLI"]
    },
    {
      category: "Consensus",
      icon: Shield,
      technologies: ["Byzantine BCB", "Proof-of-Work", "Blockshare Fork", "Smart Contracts"]
    },
    {
      category: "Network",
      icon: Globe,
      technologies: ["Distributed DHT", "Chord/Kademlia", "Hashchains", "P2P Architecture"]
    },
    {
      category: "Crypto",
      icon: Zap,
      technologies: ["Digital Signatures", "Public Key Crypto", "Hash Functions", "ModCerts"]
    }
  ];

  return (
    <section className="py-20 section-padding">
      <div className="container mx-auto">
        <motion.div 
          className="text-center mb-16"
          initial={{ y: 50, opacity: 0 }}
          whileInView={{ y: 0, opacity: 1 }}
          transition={{ duration: 0.8 }}
          viewport={{ once: true }}
        >
          <h2 className="text-4xl lg:text-5xl font-bold text-libr-secondary mb-4">
            Technical Implementation
          </h2>
          <p className="text-xl text-muted-foreground max-w-3xl mx-auto">
            LIBR leverages proven distributed systems technologies and novel consensus mechanisms 
            to achieve both censorship resistance and effective community moderation.
          </p>
        </motion.div>
        
        <div className="grid md:grid-cols-2 lg:grid-cols-4 gap-6">
          {techStack.map((stack, index) => (
            <motion.div
              key={stack.category}
              className="libr-card p-6 text-center"
              initial={{ y: 50, opacity: 0 }}
              whileInView={{ y: 0, opacity: 1 }}
              transition={{ duration: 0.6, delay: index * 0.1 }}
              viewport={{ once: true }}
            >
              <div className="w-12 h-12 bg-gradient-to-r from-libr-accent1 to-libr-accent2 rounded-lg flex items-center justify-center mx-auto mb-4">
                <stack.icon className="w-6 h-6 text-white" />
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
    <section className="py-20 section-padding bg-muted/30">
      <div className="container mx-auto">
        <div className="grid lg:grid-cols-2 gap-12 items-center">
          <motion.div
            initial={{ x: -100, opacity: 0 }}
            whileInView={{ x: 0, opacity: 1 }}
            transition={{ duration: 0.8 }}
            viewport={{ once: true }}
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
                  transition={{ duration: 0.6, delay: index * 0.2 }}
                  viewport={{ once: true }}
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
            transition={{ duration: 0.8 }}
            viewport={{ once: true }}
            className="space-y-6"
          >
            {securityFeatures.slice(2).map((feature, index) => (
              <motion.div
                key={feature.title}
                className="flex gap-4"
                initial={{ y: 30, opacity: 0 }}
                whileInView={{ y: 0, opacity: 1 }}
                transition={{ duration: 0.6, delay: (index + 2) * 0.2 }}
                viewport={{ once: true }}
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
              transition={{ duration: 0.6, delay: 0.8 }}
              viewport={{ once: true }}
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
          transition={{ duration: 0.8 }}
          viewport={{ once: true }}
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
            of censorship-resistant communication systems.
          </motion.p>
          
          <motion.div 
            className="flex flex-col sm:flex-row gap-4 justify-center items-center"
            initial={{ y: 30, opacity: 0 }}
            whileInView={{ y: 0, opacity: 1 }}
            transition={{ duration: 0.6, delay: 0.6 }}
            viewport={{ once: true }}
          >
            <button className="libr-button-primary text-lg">
              <Users className="w-5 h-5 mr-2" />
              View Documentation
            </button>
            <button className="libr-button-secondary text-lg">
              <Code className="w-5 h-5 mr-2" />
              Explore Code
            </button>
          </motion.div>
          
          <motion.div 
            className="mt-8 text-sm text-muted-foreground"
            initial={{ opacity: 0 }}
            whileInView={{ opacity: 1 }}
            transition={{ duration: 0.6, delay: 0.8 }}
            viewport={{ once: true }}
          >
            <p>Open Research • Academic Project • MIT Licensed</p>
          </motion.div>
        </motion.div>
      </div>
    </section>
  );
};

export { TechStackSection, SecuritySection, CallToActionSection };
