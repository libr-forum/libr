import React, { useEffect } from 'react';
import { motion } from 'framer-motion';
import { Link } from 'react-router-dom';
import { ArrowLeft, Scale, Users, AlertTriangle, Shield, Gavel, FileText } from 'lucide-react';
import { Header } from '../components/LandingPageSections';
import { Footer } from '../components/LandingPageExtended';

interface TermsAndConditionsProps {
  isDarkMode?: boolean;
  toggleTheme?: () => void;
}

const TermsAndConditions: React.FC<TermsAndConditionsProps> = ({ isDarkMode = true, toggleTheme = () => {} }) => {
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
              <Scale className="w-8 h-8 text-libr-accent1 mr-3" />
              <h1 className="text-4xl font-bold bg-gradient-to-r from-libr-accent1 to-libr-accent2 bg-clip-text text-transparent">
                Terms and Conditions
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
              <FileText className="w-6 h-6 text-libr-accent1 mr-3" />
              <h2 className="text-2xl font-semibold text-foreground">Agreement Overview</h2>
            </div>
            <p className="text-muted-foreground leading-relaxed">
              These Terms and Conditions govern your use of the libr protocol, applications, and related services. 
              By accessing or using libr, you agree to be bound by these terms and our Privacy Policy.
            </p>
          </div>

          <div className="space-y-8">
            <section className="bg-card/30 rounded-xl p-6 border border-border">
              <div className="flex items-center mb-4">
                <Users className="w-6 h-6 text-libr-accent1 mr-3" />
                <h3 className="text-xl font-semibold text-foreground">Acceptance of Terms</h3>
              </div>
              <div className="space-y-4 text-muted-foreground">
                <p>
                  By using libr, you confirm that you are at least 13 years old (or the minimum age required in your jurisdiction) 
                  and have the legal capacity to enter into this agreement.
                </p>
                <p>
                  If you are using libr on behalf of an organization, you represent that you have the authority to bind 
                  that organization to these terms.
                </p>
              </div>
            </section>

            <section className="bg-card/30 rounded-xl p-6 border border-border">
              <div className="flex items-center mb-4">
                <Shield className="w-6 h-6 text-libr-accent1 mr-3" />
                <h3 className="text-xl font-semibold text-foreground">Description of Service</h3>
              </div>
              <div className="space-y-4 text-muted-foreground">
                <p>
                  libr is an open-source, decentralized social network protocol that enables censorship-resistant 
                  communication while maintaining community-driven moderation capabilities.
                </p>
                <div>
                  <h4 className="font-medium text-foreground mb-2">Key Features:</h4>
                  <ul className="space-y-2">
                    <li className="flex items-start">
                      <div className="w-2 h-2 bg-libr-accent1 rounded-full mt-2 mr-3 flex-shrink-0"></div>
                      <span>Decentralized content storage and distribution</span>
                    </li>
                    <li className="flex items-start">
                      <div className="w-2 h-2 bg-libr-accent1 rounded-full mt-2 mr-3 flex-shrink-0"></div>
                      <span>Cryptographic authentication and integrity verification</span>
                    </li>
                    <li className="flex items-start">
                      <div className="w-2 h-2 bg-libr-accent1 rounded-full mt-2 mr-3 flex-shrink-0"></div>
                      <span>Community-driven moderation systems</span>
                    </li>
                    <li className="flex items-start">
                      <div className="w-2 h-2 bg-libr-accent1 rounded-full mt-2 mr-3 flex-shrink-0"></div>
                      <span>Peer-to-peer networking and communication</span>
                    </li>
                  </ul>
                </div>
              </div>
            </section>

            <section className="bg-card/30 rounded-xl p-6 border border-border">
              <div className="flex items-center mb-4">
                <Gavel className="w-6 h-6 text-libr-accent1 mr-3" />
                <h3 className="text-xl font-semibold text-foreground">User Responsibilities</h3>
              </div>
              <div className="space-y-4 text-muted-foreground">
                <div>
                  <h4 className="font-medium text-foreground mb-2">Account Security</h4>
                  <p>You are responsible for maintaining the security of your account credentials, private keys, and any devices used to access libr.</p>
                </div>
                <div>
                  <h4 className="font-medium text-foreground mb-2">Content Guidelines</h4>
                  <p>You agree not to use libr to:</p>
                  <ul className="mt-2 space-y-1">
                    <li>• Distribute illegal content or engage in illegal activities</li>
                    <li>• Harass, threaten, or harm other users</li>
                    <li>• Distribute malware, viruses, or other harmful code</li>
                    <li>• Violate intellectual property rights</li>
                    <li>• Attempt to disrupt network operations or security</li>
                    <li>• Impersonate others or provide false information</li>
                  </ul>
                </div>
                <div>
                  <h4 className="font-medium text-foreground mb-2">Compliance</h4>
                  <p>You must comply with all applicable laws and regulations in your jurisdiction when using libr.</p>
                </div>
              </div>
            </section>

            <section className="bg-gradient-to-r from-yellow-500/10 to-orange-500/10 rounded-xl p-6 border border-yellow-500/20">
              <div className="flex items-center mb-4">
                <AlertTriangle className="w-6 h-6 text-yellow-500 mr-3" />
                <h3 className="text-xl font-semibold text-foreground">Decentralized Nature and Limitations</h3>
              </div>
              <div className="space-y-4 text-muted-foreground">
                <p>
                  <strong className="text-foreground">Important:</strong> libr operates as a decentralized protocol. 
                  This means that once content is distributed across the network, complete removal may be technically impossible.
                </p>
                <div>
                  <h4 className="font-medium text-foreground mb-2">Service Availability</h4>
                  <p>
                    We provide the protocol and reference implementations "as is" without guarantees of uptime, 
                    availability, or performance. Network availability depends on the distributed community of node operators.
                  </p>
                </div>
                <div>
                  <h4 className="font-medium text-foreground mb-2">Content Persistence</h4>
                  <p>
                    Content shared on libr may persist on the network indefinitely. Consider carefully what you share 
                    and use appropriate privacy settings to control content distribution.
                  </p>
                </div>
              </div>
            </section>

            <section className="bg-card/30 rounded-xl p-6 border border-border">
              <h3 className="text-xl font-semibold text-foreground mb-4">Intellectual Property</h3>
              <div className="space-y-4 text-muted-foreground">
                <p>
                  The libr protocol is open-source software released under appropriate open-source licenses. 
                  You retain ownership of content you create, but grant necessary licenses for distribution across the network.
                </p>
                <p>
                  By posting content, you grant libr and other network participants the right to store, 
                  distribute, and display your content as necessary for protocol operation.
                </p>
              </div>
            </section>

            <section className="bg-card/30 rounded-xl p-6 border border-border">
              <h3 className="text-xl font-semibold text-foreground mb-4">Disclaimer of Warranties</h3>
              <div className="space-y-4 text-muted-foreground">
                <p>
                  libr is provided "as is" without warranties of any kind, either express or implied. 
                  We do not warrant that the service will be uninterrupted, secure, or error-free.
                </p>
                <p>
                  You use libr at your own risk. We are not responsible for content posted by other users 
                  or actions taken by network participants.
                </p>
              </div>
            </section>

            <section className="bg-card/30 rounded-xl p-6 border border-border">
              <h3 className="text-xl font-semibold text-foreground mb-4">Limitation of Liability</h3>
              <div className="space-y-4 text-muted-foreground">
                <p>
                  To the maximum extent permitted by law, libr and its contributors shall not be liable for any 
                  indirect, incidental, special, consequential, or punitive damages arising from your use of the service.
                </p>
                <p>
                  Our total liability to you for any claims related to libr shall not exceed the amount you have paid 
                  us in the twelve months preceding the claim (which may be zero for free services).
                </p>
              </div>
            </section>

            <section className="bg-card/30 rounded-xl p-6 border border-border">
              <h3 className="text-xl font-semibold text-foreground mb-4">Governing Law and Disputes</h3>
              <div className="space-y-4 text-muted-foreground">
                <p>
                  These terms are governed by the laws of the jurisdiction where the libr development team is based. 
                  Any disputes will be resolved through appropriate legal channels in that jurisdiction.
                </p>
                <p>
                  Given the decentralized nature of libr, disputes with other users should be resolved through 
                  community moderation mechanisms where possible.
                </p>
              </div>
            </section>

            <section className="bg-card/30 rounded-xl p-6 border border-border">
              <h3 className="text-xl font-semibold text-foreground mb-4">Changes to Terms</h3>
              <p className="text-muted-foreground">
                We may update these terms from time to time. Material changes will be communicated through our 
                official channels with reasonable notice. Continued use of libr after changes constitutes acceptance 
                of the updated terms.
              </p>
            </section>

            <section className="bg-card/30 rounded-xl p-6 border border-border">
              <h3 className="text-xl font-semibold text-foreground mb-4">Contact Information</h3>
              <p className="text-muted-foreground mb-4">
                For questions about these terms or to report violations, contact us at:
              </p>
              <div className="space-y-2 text-muted-foreground">
                <p><strong className="text-foreground">Email:</strong> libr.forum@gmail.com</p>
                <p><strong className="text-foreground">GitHub:</strong> https://github.com/devlup-labs/Libr</p>
                <p><strong className="text-foreground">Issue Tracker:</strong> Report violations through our GitHub issues</p>
              </div>
            </section>
          </div>
        </motion.div>
      </div>
      <Footer />
    </div>
  );
};

export default TermsAndConditions;
