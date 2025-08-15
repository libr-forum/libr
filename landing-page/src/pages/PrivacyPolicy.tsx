import React, { useEffect } from 'react';
import { motion } from 'framer-motion';
import { Link } from 'react-router-dom';
import { ArrowLeft, Shield, Eye, Database, Share2, Bell, Settings } from 'lucide-react';
import { Header } from '../components/LandingPageSections';
import { Footer } from '../components/LandingPageExtended';

interface PrivacyPolicyProps {
  isDarkMode?: boolean;
  toggleTheme?: () => void;
}

const PrivacyPolicy: React.FC<PrivacyPolicyProps> = ({ isDarkMode = true, toggleTheme = () => {} }) => {
  useEffect(() => {
    window.scrollTo(0, 0);
  }, []);

  return (
    <div className="min-h-screen bg-libr-primary/50 text-foreground">
      <Header isDark={isDarkMode} toggleTheme={toggleTheme} />
      <div className="container mx-auto px-4 py-8 max-w-4xl mt-20">{/* Added mt-20 for header spacing */}
          {/* Header */}
          <motion.div
            initial={{ opacity: 0, y: -20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.6 }}
            className="mb-8"
          >
            <Link
              to="/"
              className="inline-flex items-center text-libr-accent1 hover:text-libr-accent2 transition-colors mb-6"
            >
              <ArrowLeft className="w-4 h-4 mr-2" />
              Back to Home
            </Link>
            
            <div className="flex items-center mb-4">
              <Shield className="w-8 h-8 text-libr-accent1 mr-3" />
              <h1 className="text-4xl font-bold bg-gradient-to-r from-libr-accent1 to-libr-accent2 bg-clip-text text-transparent">
                Privacy Policy
              </h1>
            </div>
            
            <p className="text-muted-foreground text-lg">
              Last updated: August 15, 2025
            </p>
          </motion.div>

        {/* Content */}
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.6, delay: 0.2 }}
          className="prose prose-slate dark:prose-invert max-w-none"
        >
          <div className="bg-card/50 rounded-xl p-8 border border-border mb-8">
            <div className="flex items-center mb-4">
              <Eye className="w-6 h-6 text-libr-accent1 mr-3" />
              <h2 className="text-2xl font-semibold text-foreground">Overview</h2>
            </div>
            <p className="text-muted-foreground leading-relaxed">
              libr is a decentralized, open-source social network framework designed with privacy and censorship resistance at its core. 
              This Privacy Policy explains how we handle your information when you use our protocol, applications, and services.
            </p>
          </div>

          <div className="space-y-8">
            <section className="bg-card/30 rounded-xl p-6 border border-border">
              <div className="flex items-center mb-4">
                <Database className="w-6 h-6 text-libr-accent1 mr-3" />
                <h3 className="text-xl font-semibold text-foreground">Information We Collect</h3>
              </div>
              <div className="space-y-4 text-muted-foreground">
                <div>
                  <h4 className="font-medium text-foreground mb-2">Account Information</h4>
                  <p>When you create an account, we may collect your username, email address (optional), and cryptographic public keys generated locally on your device.</p>
                </div>
                <div>
                  <h4 className="font-medium text-foreground mb-2">Content Data</h4>
                  <p>Posts, messages, and interactions you create are stored in a distributed manner across the network. Content is cryptographically signed but may be publicly visible depending on your privacy settings.</p>
                </div>
                <div>
                  <h4 className="font-medium text-foreground mb-2">Technical Information</h4>
                  <p>Network metadata, IP addresses for direct peer connections, and usage analytics to improve the protocol's performance and security.</p>
                </div>
              </div>
            </section>

            <section className="bg-card/30 rounded-xl p-6 border border-border">
              <div className="flex items-center mb-4">
                <Share2 className="w-6 h-6 text-libr-accent1 mr-3" />
                <h3 className="text-xl font-semibold text-foreground">How We Use Your Information</h3>
              </div>
              <ul className="space-y-3 text-muted-foreground">
                <li className="flex items-start">
                  <div className="w-2 h-2 bg-libr-accent1 rounded-full mt-2 mr-3 flex-shrink-0"></div>
                  <span>Facilitate secure, peer-to-peer communication and content sharing</span>
                </li>
                <li className="flex items-start">
                  <div className="w-2 h-2 bg-libr-accent1 rounded-full mt-2 mr-3 flex-shrink-0"></div>
                  <span>Enable content moderation through community-driven mechanisms</span>
                </li>
                <li className="flex items-start">
                  <div className="w-2 h-2 bg-libr-accent1 rounded-full mt-2 mr-3 flex-shrink-0"></div>
                  <span>Maintain network integrity and prevent spam or malicious behavior</span>
                </li>
                <li className="flex items-start">
                  <div className="w-2 h-2 bg-libr-accent1 rounded-full mt-2 mr-3 flex-shrink-0"></div>
                  <span>Improve protocol performance and develop new features</span>
                </li>
              </ul>
            </section>

            <section className="bg-card/30 rounded-xl p-6 border border-border">
              <div className="flex items-center mb-4">
                <Settings className="w-6 h-6 text-libr-accent1 mr-3" />
                <h3 className="text-xl font-semibold text-foreground">Data Storage and Security</h3>
              </div>
              <div className="space-y-4 text-muted-foreground">
                <p>
                  libr employs a distributed architecture where data is stored across multiple nodes in the network. 
                  Your private keys remain on your device and are never transmitted to our servers.
                </p>
                <p>
                  All communications are encrypted end-to-end, and content is cryptographically signed to ensure authenticity. 
                  We implement industry-standard security measures to protect against unauthorized access.
                </p>
              </div>
            </section>

            <section className="bg-card/30 rounded-xl p-6 border border-border">
              <div className="flex items-center mb-4">
                <Bell className="w-6 h-6 text-libr-accent1 mr-3" />
                <h3 className="text-xl font-semibold text-foreground">Your Rights and Choices</h3>
              </div>
              <ul className="space-y-3 text-muted-foreground">
                <li className="flex items-start">
                  <div className="w-2 h-2 bg-libr-accent2 rounded-full mt-2 mr-3 flex-shrink-0"></div>
                  <span><strong className="text-foreground">Access and Portability:</strong> Export your data at any time using standard protocols</span>
                </li>
                <li className="flex items-start">
                  <div className="w-2 h-2 bg-libr-accent2 rounded-full mt-2 mr-3 flex-shrink-0"></div>
                  <span><strong className="text-foreground">Deletion:</strong> Delete your account and associated data from nodes you control</span>
                </li>
                <li className="flex items-start">
                  <div className="w-2 h-2 bg-libr-accent2 rounded-full mt-2 mr-3 flex-shrink-0"></div>
                  <span><strong className="text-foreground">Privacy Controls:</strong> Configure visibility and sharing settings for your content</span>
                </li>
                <li className="flex items-start">
                  <div className="w-2 h-2 bg-libr-accent2 rounded-full mt-2 mr-3 flex-shrink-0"></div>
                  <span><strong className="text-foreground">Opt-out:</strong> Disable analytics and telemetry in your client settings</span>
                </li>
              </ul>
            </section>

            <section className="bg-gradient-to-r from-libr-accent1/10 to-libr-accent2/10 rounded-xl p-6 border border-libr-accent1/20">
              <h3 className="text-xl font-semibold text-foreground mb-4">Decentralized Nature</h3>
              <p className="text-muted-foreground mb-4">
                As a decentralized protocol, libr operates differently from traditional social networks. 
                Once content is distributed across the network, complete removal may not be technically possible, 
                similar to how email or other internet protocols work.
              </p>
              <p className="text-muted-foreground">
                We encourage users to carefully consider what they share and to use the privacy controls available 
                to limit the distribution of sensitive information.
              </p>
            </section>

            <section className="bg-card/30 rounded-xl p-6 border border-border">
              <h3 className="text-xl font-semibold text-foreground mb-4">Contact Information</h3>
              <p className="text-muted-foreground mb-4">
                If you have questions about this Privacy Policy or how we handle your data, please contact us:
              </p>
              <div className="space-y-2 text-muted-foreground">
                <p><strong className="text-foreground">Email:</strong> libr.forum@gmail.com</p>
                <p><strong className="text-foreground">GitHub:</strong> https://github.com/devlup-labs/Libr</p>
                <p><strong className="text-foreground">Documentation:</strong> Available in our repository</p>
              </div>
            </section>

            <section className="bg-card/30 rounded-xl p-6 border border-border">
              <h3 className="text-xl font-semibold text-foreground mb-4">Changes to This Policy</h3>
              <p className="text-muted-foreground">
                We may update this Privacy Policy from time to time. We will notify users of any material changes 
                through our official communication channels and update the "Last updated" date at the top of this policy.
              </p>
            </section>
          </div>
        </motion.div>
      </div>
      <Footer />
    </div>
  );
};

export default PrivacyPolicy;
