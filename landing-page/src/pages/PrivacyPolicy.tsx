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
      <div className="container mx-auto px-4 py-8 max-w-4xl mt-20">
        {/* Header */}
        <motion.div
          initial={{ opacity: 0, y: -20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.6 }}
          className="mb-8"
        >
          <Link
            to="/"
            className="inline-flex items-center text-libr-foreground hover:text-libr-secondary transition-colors mb-6"
          >
            <ArrowLeft className="w-4 h-4 mr-2" />
            Back to Home
          </Link>

          <div className="flex items-center mb-4 text-libr-foreground">
            <Shield className="w-8 h-8 mr-3" />
            <h1 className="text-4xl font-bold">
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
              <Eye className="w-6 h-6 text-libr-secondary mr-3" />
              <h2 className="text-2xl font-semibold text-foreground">Introduction</h2>
            </div>
            <p className="text-muted-foreground leading-relaxed mb-4">
              libr is a peer-to-peer (P2P) social network framework designed for censorship resilience and community moderation. 
              It enables decentralized content storage and verifiable messaging, with transparent, community-driven governance—not centralized control.
            </p>
            <p className="text-muted-foreground leading-relaxed">
              This Privacy Policy explains how your data is collected, used, and what remains beyond our control due to the decentralized nature of the network. 
              By using libr, you agree to the terms outlined in this policy.
            </p>
          </div>

          <div className="space-y-8">
            <section className="bg-card/30 rounded-xl p-6 border border-border">
              <div className="flex items-center mb-4">
                <Database className="w-6 h-6 text-libr-secondary mr-3" />
                <h3 className="text-xl font-semibold text-foreground">Information We Collect or Expose</h3>
              </div>
              <div className="space-y-4 text-muted-foreground">
                <div>
                  <h4 className="font-medium text-foreground mb-2">IP Address Exposure</h4>
                  <p>Your IP address is exposed to public relays you connect with. This is necessary for routing messages in a P2P architecture and enables direct peer-to-peer communication.</p>
                </div>
                <div>
                  <h4 className="font-medium text-foreground mb-2">Public Content</h4>
                  <p>Any content you publish in libr is publicly accessible to the entire community and may be stored, copied, and redistributed by peers indefinitely. All content is cryptographically signed to ensure authenticity.</p>
                </div>
                <div>
                  <h4 className="font-medium text-foreground mb-2">Pseudonym Identifier</h4>
                  <p>Your identity in libr is a cryptographically-generated pseudonym created locally on your device. Only you can use it to post content. If you reset your identity, future messages cannot be straightforwardly linked to your previous pseudonym.</p>
                </div>
                <div>
                  <h4 className="font-medium text-foreground mb-2">Client Analytics (Optional)</h4>
                  <p>Our client applications may collect anonymous usage analytics to improve performance and identify issues. This can be disabled in your client settings and does not include personal content.</p>
                </div>
              </div>
            </section>

            <section className="bg-gradient-to-r from-orange-500/10 to-red-500/10 rounded-xl p-6 border border-orange-500/20">
              <div className="flex items-center mb-4">
                <Shield className="w-6 h-6 text-orange-500 mr-3" />
                <h3 className="text-xl font-semibold text-foreground">What We Do Not Control</h3>
              </div>
              <div className="space-y-4 text-muted-foreground">
                <div className="bg-orange-500/5 rounded-lg p-4 border border-orange-500/20">
                  <h4 className="font-medium text-foreground mb-2 flex items-center">
                    <Bell className="w-4 h-4 mr-2 text-orange-500" />
                    Important: Decentralization Limitations
                  </h4>
                  <ul className="space-y-2">
                    <li>• Due to decentralization, content you share may remain stored on others' devices permanently</li>
                    <li>• We have no power to guarantee content deletion or modification across the network</li>
                    <li>• Peer-hosted data (on relays or other users' devices) is entirely outside our control</li>
                    <li>• Node operators may have their own data retention and privacy policies</li>
                  </ul>
                </div>
                <p>
                  This is an inherent characteristic of peer-to-peer networks that prioritize censorship resistance. 
                  Please consider this carefully before sharing sensitive information.
                </p>
              </div>
            </section>

            <section className="bg-card/30 rounded-xl p-6 border border-border">
              <div className="flex items-center mb-4">
                <Share2 className="w-6 h-6 text-libr-secondary mr-3" />
                <h3 className="text-xl font-semibold text-foreground">How We Use Information</h3>
              </div>
              <div className="space-y-4 text-muted-foreground">
                <p className="mb-4">Our use of any collected or exposed data is strictly limited to:</p>
                <ul className="space-y-3">
                  <li className="flex items-start">
                    <div className="w-2 h-2 bg-libr-secondary rounded-full mt-2 mr-3 flex-shrink-0"></div>
                    <span>Enabling message and content delivery within the P2P network</span>
                  </li>
                  <li className="flex items-start">
                    <div className="w-2 h-2 bg-libr-secondary rounded-full mt-2 mr-3 flex-shrink-0"></div>
                    <span>Supporting secure and spam-resistant communication</span>
                  </li>
                  <li className="flex items-start">
                    <div className="w-2 h-2 bg-libr-secondary rounded-full mt-2 mr-3 flex-shrink-0"></div>
                    <span>Maintaining the integrity of the community moderation framework</span>
                  </li>
                  <li className="flex items-start">
                    <div className="w-2 h-2 bg-libr-secondary rounded-full mt-2 mr-3 flex-shrink-0"></div>
                    <span>Improving protocol performance and developing new features</span>
                  </li>
                </ul>
                <p className="font-medium text-foreground mt-4">We do not sell or commercially share user data.</p>
              </div>
            </section>

            <section className="bg-card/30 rounded-xl p-6 border border-border">
              <div className="flex items-center mb-4">
                <Database className="w-6 h-6 text-libr-secondary mr-3" />
                <h3 className="text-xl font-semibold text-foreground">Data Sharing</h3>
              </div>
              <div className="space-y-4 text-muted-foreground">
                <p>Data is shared only when required for:</p>
                <ul className="space-y-2">
                  <li>• Delivering messages to peers in the network</li>
                  <li>• Relay-hosted functionality, subject to relay operators' own policies</li>
                  <li>• Community moderation processes (voting, reputation systems)</li>
                </ul>
                <p className="font-medium text-foreground">
                  No centralized storage of content is maintained by libr beyond temporary relay communication, if any.
                </p>
              </div>
            </section>

            <section className="bg-card/30 rounded-xl p-6 border border-border">
              <div className="flex items-center mb-4">
                <Settings className="w-6 h-6 text-libr-secondary mr-3" />
                <h3 className="text-xl font-semibold text-foreground">Data Retention</h3>
              </div>
              <div className="space-y-4 text-muted-foreground">
                <ul className="space-y-3">
                  <li className="flex items-start">
                    <div className="w-2 h-2 bg-libr-accent2 rounded-full mt-2 mr-3 flex-shrink-0"></div>
                    <span>We do not and cannot centrally store user content</span>
                  </li>
                  <li className="flex items-start">
                    <div className="w-2 h-2 bg-libr-accent2 rounded-full mt-2 mr-3 flex-shrink-0"></div>
                    <span>Any relay logs are retained only as needed for operational purposes</span>
                  </li>
                  <li className="flex items-start">
                    <div className="w-2 h-2 bg-libr-accent2 rounded-full mt-2 mr-3 flex-shrink-0"></div>
                    <span>Peer-stored data may persist indefinitely across the network</span>
                  </li>
                  <li className="flex items-start">
                    <div className="w-2 h-2 bg-libr-accent2 rounded-full mt-2 mr-3 flex-shrink-0"></div>
                    <span>Client-side data (keys, settings) remains on your device until manually deleted</span>
                  </li>
                </ul>
              </div>
            </section>

            <section className="bg-card/30 rounded-xl p-6 border border-border">
              <div className="flex items-center mb-4">
                <Shield className="w-6 h-6 text-libr-secondary mr-3" />
                <h3 className="text-xl font-semibold text-foreground">Security</h3>
              </div>
              <div className="space-y-4 text-muted-foreground">
                <ul className="space-y-3">
                  <li className="flex items-start">
                    <div className="w-2 h-2 bg-libr-secondary rounded-full mt-2 mr-3 flex-shrink-0"></div>
                    <span>Your pseudonym and cryptographic identity are secure and unique to you</span>
                  </li>
                  <li className="flex items-start">
                    <div className="w-2 h-2 bg-libr-secondary rounded-full mt-2 mr-3 flex-shrink-0"></div>
                    <span>All content is cryptographically signed to ensure authenticity and prevent tampering</span>
                  </li>
                  <li className="flex items-start">
                    <div className="w-2 h-2 bg-libr-secondary rounded-full mt-2 mr-3 flex-shrink-0"></div>
                    <span>Private keys are generated and stored locally on your device only</span>
                  </li>
                </ul>
                <p className="mt-4">
                  However, P2P communication inherently exposes your IP address and relies on peers for content storage—both beyond our full control.
                </p>
              </div>
            </section>

            <section className="bg-card/30 rounded-xl p-6 border border-border">
              <div className="flex items-center mb-4">
                <Bell className="w-6 h-6 text-libr-secondary mr-3" />
                <h3 className="text-xl font-semibold text-foreground">Your Rights</h3>
              </div>
              <div className="space-y-4 text-muted-foreground">
                <p className="mb-4">
                  You may have the right, depending on jurisdiction, to access, correct, or request deletion of data. 
                  However, due to decentralization, deleting peer-stored data may not be feasible.
                </p>
                <ul className="space-y-3">
                  <li className="flex items-start">
                    <div className="w-2 h-2 bg-libr-accent2 rounded-full mt-2 mr-3 flex-shrink-0"></div>
                    <span><strong className="text-foreground">Access and Portability:</strong> Export your data at any time using standard protocols</span>
                  </li>
                  <li className="flex items-start">
                    <div className="w-2 h-2 bg-libr-accent2 rounded-full mt-2 mr-3 flex-shrink-0"></div>
                    <span><strong className="text-foreground">Identity Reset:</strong> Generate a new cryptographic identity at any time</span>
                  </li>
                  <li className="flex items-start">
                    <div className="w-2 h-2 bg-libr-accent2 rounded-full mt-2 mr-3 flex-shrink-0"></div>
                    <span><strong className="text-foreground">Client Control:</strong> Configure privacy settings and disable analytics</span>
                  </li>
                  <li className="flex items-start">
                    <div className="w-2 h-2 bg-libr-accent2 rounded-full mt-2 mr-3 flex-shrink-0"></div>
                    <span><strong className="text-foreground">Data Requests:</strong> Contact us to request removal of any data we control</span>
                  </li>
                </ul>
              </div>
            </section>

            <section className="bg-gradient-to-r from-red-500/10 to-orange-500/10 rounded-xl p-6 border border-red-500/20">
              <div className="flex items-center mb-4">
                <Bell className="w-6 h-6 text-red-500 mr-3" />
                <h3 className="text-xl font-semibold text-foreground">Persistence Warning</h3>
              </div>
              <div className="bg-red-500/5 rounded-lg p-4 border border-red-500/20">
                <p className="text-muted-foreground font-medium">
                  <strong className="text-foreground">Important:</strong> Because libr is peer-to-peer, any content you share may be stored by others indefinitely. 
                  We cannot ensure deletion or modification of data on other users' devices.
                </p>
                <p className="text-muted-foreground mt-3">
                  This is an inherent characteristic of decentralized networks designed for censorship resistance. 
                  Please think carefully about what you share and use appropriate privacy controls.
                </p>
              </div>
            </section>

            <section className="bg-card/30 rounded-xl p-6 border border-border">
              <h3 className="text-xl font-semibold text-foreground mb-4">Governing Law</h3>
              <p className="text-muted-foreground">
                This Privacy Policy is governed by and construed in accordance with the laws of India, 
                without regard to its conflict of law principles.
              </p>
            </section>

            <section className="bg-card/30 rounded-xl p-6 border border-border">
              <h3 className="text-xl font-semibold text-foreground mb-4">Changes to This Policy</h3>
              <p className="text-muted-foreground">
                We may update this Privacy Policy from time to time. Changes will be published on our website and within the app. 
                Continued use after updates constitutes acceptance of the new terms. Material changes will be communicated through 
                our official channels with reasonable notice.
              </p>
            </section>

            <section className="bg-card/30 rounded-xl p-6 border border-border">
              <h3 className="text-xl font-semibold text-foreground mb-4">Contact and Governance</h3>
              <p className="text-muted-foreground mb-4">
                As libr is decentralized, there is no single corporate operator. For governance-related queries, appeals, or issues, 
                please engage with the relevant Community's official channels, as specified in the libr protocol documentation.
              </p>
              <div className="space-y-2 text-muted-foreground">
                <p><strong className="text-foreground">Email:</strong> <a href="mailto:libr.forum@gmail.com" target="_blank" rel="noopener noreferrer" className="text-libr-accent1 hover:text-libr-accent2 hover:underline">libr.forum@gmail.com</a></p>
                <p><strong className="text-foreground">GitHub:</strong> <a href="https://github.com/libr-forum/libr" target="_blank" rel="noopener noreferrer" className="text-libr-accent1 hover:text-libr-accent2 hover:underline">https://github.com/libr-forum/libr</a></p>
                <p><strong className="text-foreground">Website:</strong> <a href="https://libr-ashen.vercel.app/" target="_blank" rel="noopener noreferrer" className="text-libr-accent1 hover:text-libr-accent2 hover:underline">https://libr-ashen.vercel.app/</a></p>
                <p><strong className="text-foreground">Apache 2.0 License:</strong> <a href="http://www.apache.org/licenses/LICENSE-2.0" target="_blank" rel="noopener noreferrer" className="text-libr-accent1 hover:text-libr-accent2 hover:underline">View Full License Text</a></p>
                <p><strong className="text-foreground">Terms & Conditions:</strong> <a href="/terms-and-conditions" className="text-libr-accent1 hover:text-libr-accent2 hover:underline">View our Terms & Conditions</a></p>
                <p><strong className="text-foreground">EULA:</strong> <a href="/eula" className="text-libr-accent1 hover:text-libr-accent2 hover:underline">View our End User License Agreement</a></p>
              </div>
            </section>
          </div>
        </motion.div>
      </div>
      <Footer />
    </div>
  );
};

export default PrivacyPolicy;
