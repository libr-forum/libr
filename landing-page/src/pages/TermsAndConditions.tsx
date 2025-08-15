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
              className="inline-flex items-center text-libr-secondary hover:text-foreground transition-colors mb-6"
            >
              <ArrowLeft className="w-4 h-4 mr-2" />
              Back to Home
            </Link>
            
            <div className="flex items-center mb-4 text-libr-secondary">
              <Scale className="w-8 h-8 mr-3" />
              <h1 className="text-4xl font-bold">
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
              <FileText className="w-6 h-6 text-libr-secondary mr-3" />
              <h2 className="text-2xl font-semibold text-foreground">Terms and Conditions for libr</h2>
            </div>
            <div className="space-y-2 text-muted-foreground">
              <p><strong className="text-foreground">Effective Date:</strong> August 15, 2025</p>
              <p><strong className="text-foreground">Last Updated:</strong> August 15, 2025</p>
            </div>
            <p className="text-muted-foreground leading-relaxed mt-4">
              Welcome to libr, a decentralized, censorship-resilient public forum with community-driven moderation. 
              These Terms and Conditions govern your use of the Platform, including all features, services, and tools provided through libr.
              By accessing or using libr, you agree to be bound by these Terms. If you do not agree, do not use the Platform.
            </p>
          </div>

          <div className="space-y-8">
            <section className="bg-card/30 rounded-xl p-6 border border-border">
              <div className="flex items-center mb-4">
                <Users className="w-6 h-6 text-libr-secondary mr-3" />
                <h3 className="text-xl font-semibold text-foreground">The libr Framework & Key Definitions</h3>
              </div>
              <div className="space-y-4 text-muted-foreground">
                <p>
                  libr operates as a decentralized network composed of independent participants ("Nodes"). The Platform is not centrally owned or controlled. 
                  Due to this decentralized nature, content may be stored across multiple independent servers globally, and no single entity can unilaterally 
                  alter or remove lawful content once it has been validated and stored.
                </p>
                <div>
                  <h4 className="font-medium text-foreground mb-2">Key Participants in the libr Ecosystem:</h4>
                  <ul className="space-y-2">
                    <li className="flex items-start">
                      <div className="w-2 h-2 bg-libr-secondary rounded-full mt-2 mr-3 flex-shrink-0"></div>
                      <div>
                        <strong className="text-foreground">Clients:</strong> Users who interact with the network to create, send, and read messages.
                      </div>
                    </li>
                    <li className="flex items-start">
                      <div className="w-2 h-2 bg-libr-secondary rounded-full mt-2 mr-3 flex-shrink-0"></div>
                      <div>
                        <strong className="text-foreground">Database Nodes (DBs):</strong> Participants who contribute storage and network resources to store messages, ensuring data availability and resilience.
                      </div>
                    </li>
                    <li className="flex items-start">
                      <div className="w-2 h-2 bg-libr-secondary rounded-full mt-2 mr-3 flex-shrink-0"></div>
                      <div>
                        <strong className="text-foreground">Moderators (Mods):</strong> Participants elected by a Community to evaluate and moderate content based on that Community's established guidelines.
                      </div>
                    </li>
                  </ul>
                </div>
              </div>
            </section>

            <section className="bg-card/30 rounded-xl p-6 border border-border">
              <div className="flex items-center mb-4">
                <Shield className="w-6 h-6 text-libr-secondary mr-3" />
                <h3 className="text-xl font-semibold text-foreground">User Eligibility and Account Security</h3>
              </div>
              <div className="space-y-4 text-muted-foreground">
                <div>
                  <h4 className="font-medium text-foreground mb-2">To use libr, you must:</h4>
                  <ul className="space-y-2">
                    <li className="flex items-start">
                      <div className="w-2 h-2 bg-libr-secondary rounded-full mt-2 mr-3 flex-shrink-0"></div>
                      <span>Comply with all applicable laws in your jurisdiction</span>
                    </li>
                    <li className="flex items-start">
                      <div className="w-2 h-2 bg-libr-secondary rounded-full mt-2 mr-3 flex-shrink-0"></div>
                      <span>Not be barred from receiving services under any applicable laws</span>
                    </li>
                    <li className="flex items-start">
                      <div className="w-2 h-2 bg-libr-secondary rounded-full mt-2 mr-3 flex-shrink-0"></div>
                      <span>Be at least 18 years of age or the age of legal majority in your jurisdiction</span>
                    </li>
                  </ul>
                </div>
                <div className="bg-amber-100/20 border border-amber-500/30 rounded-lg p-4">
                  <h4 className="font-medium text-amber-200 mb-2 flex items-center">
                    <AlertTriangle className="w-4 h-4 mr-2" />
                    Account Security
                  </h4>
                  <p className="text-amber-100 text-sm">
                    You interact with libr using cryptographic keys. You are solely responsible for maintaining the confidentiality and security of your private keys. 
                    All actions taken using your keys are considered your own. <strong>Losing your private keys will result in the permanent loss of access to your identity on the network.</strong>
                  </p>
                </div>
              </div>
            </section>

            <section className="bg-card/30 rounded-xl p-6 border border-border">
              <div className="flex items-center mb-4">
                <Gavel className="w-6 h-6 text-libr-secondary mr-3" />
                <h3 className="text-xl font-semibold text-foreground">User Responsibilities & Acceptable Use</h3>
              </div>
              <div className="space-y-4 text-muted-foreground">
                <div>
                  <h4 className="font-medium text-foreground mb-2">When using libr, you agree to:</h4>
                  <ul className="space-y-2">
                    <li className="flex items-start">
                      <div className="w-2 h-2 bg-libr-secondary rounded-full mt-2 mr-3 flex-shrink-0"></div>
                      <span>Post content that complies with all applicable laws and the moderation guidelines of the Community you are participating in</span>
                    </li>
                    <li className="flex items-start">
                      <div className="w-2 h-2 bg-libr-secondary rounded-full mt-2 mr-3 flex-shrink-0"></div>
                      <span>Accept that community-elected Moderators may approve, reject, or flag your content in accordance with their established rules</span>
                    </li>
                    <li className="flex items-start">
                      <div className="w-2 h-2 bg-libr-secondary rounded-full mt-2 mr-3 flex-shrink-0"></div>
                      <span>Understand that due to the network's design, content that is successfully stored may remain accessible indefinitely</span>
                    </li>
                    <li className="flex items-start">
                      <div className="w-2 h-2 bg-libr-secondary rounded-full mt-2 mr-3 flex-shrink-0"></div>
                      <span>Not attempt to circumvent, disable, or otherwise interfere with security-related features of the platform</span>
                    </li>
                    <li className="flex items-start">
                      <div className="w-2 h-2 bg-libr-secondary rounded-full mt-2 mr-3 flex-shrink-0"></div>
                      <span>Not engage in any activity that could disrupt or interfere with the proper functioning of the libr protocol</span>
                    </li>
                  </ul>
                </div>
                <div>
                  <h4 className="font-medium text-foreground mb-2">Prohibited Activities</h4>
                  <p>You agree not to use libr to:</p>
                  <ul className="mt-2 space-y-1">
                    <li>• Distribute illegal content or engage in illegal activities</li>
                    <li>• Harass, threaten, or harm other users</li>
                    <li>• Distribute malware, viruses, or other harmful code</li>
                    <li>• Violate intellectual property rights</li>
                    <li>• Attempt to disrupt network operations or security</li>
                    <li>• Impersonate others or provide false information</li>
                    <li>• Engage in spam, scams, or fraudulent activities</li>
                    <li>• Manipulate consensus mechanisms or governance processes</li>
                  </ul>
                </div>
              </div>
            </section>

            <section className="bg-gradient-to-r from-blue-500/10 to-purple-500/10 rounded-xl p-6 border border-blue-500/20">
              <div className="flex items-center mb-4">
                <Users className="w-6 h-6 text-blue-400 mr-3" />
                <h3 className="text-xl font-semibold text-foreground">Community Governance and Moderation</h3>
              </div>
              <div className="space-y-4 text-muted-foreground">
                <div>
                  <h4 className="font-medium text-foreground mb-2">Community-Led Moderation</h4>
                  <p>
                    Content moderation is performed by Moderators elected by each Community, not by the creators of the libr framework. 
                    Moderation decisions are based on the guidelines established by that Community and are recorded transparently 
                    (e.g., via cryptographically verifiable logs).
                  </p>
                </div>
                <div>
                  <h4 className="font-medium text-foreground mb-2">Dispute Resolution</h4>
                  <p>
                    If you believe your content was unfairly moderated, you may use the built-in ticketing system to raise a dispute. 
                    The dispute will be reviewed by the Community's Moderators according to their governance rules. The creators of the 
                    libr framework have no authority to overturn the decisions of a Community's Moderators.
                  </p>
                </div>
                <div>
                  <h4 className="font-medium text-foreground mb-2">Byzantine Fault Tolerance</h4>
                  <p>
                    The libr protocol employs Byzantine Consistent Broadcast to ensure that moderation decisions are reached through consensus, 
                    even if some moderators act maliciously or fail to respond.
                  </p>
                </div>
              </div>
            </section>

            <section className="bg-card/30 rounded-xl p-6 border border-border">
              <h3 className="text-xl font-semibold text-foreground mb-4">Content Ownership and Licensing</h3>
              <div className="space-y-4 text-muted-foreground">
                <div>
                  <h4 className="font-medium text-foreground mb-2">Ownership</h4>
                  <p>You retain full ownership of the content you create and submit to the libr network ("User Content").</p>
                </div>
                <div>
                  <h4 className="font-medium text-foreground mb-2">License to the Network</h4>
                  <p>
                    By submitting User Content, you grant all participants in the libr network (specifically, the DBs and Clients within a Community) 
                    a license to your content under the terms of the <strong className="text-foreground">Apache License, Version 2.0</strong>. 
                    This license is worldwide, non-exclusive, royalty-free, perpetual, and irrevocable.
                  </p>
                  <p>
                    You can review the full license at{' '}
                    <a href="http://www.apache.org/licenses/LICENSE-2.0" className="text-libr-accent1 hover:text-libr-accent2 underline">
                      http://www.apache.org/licenses/LICENSE-2.0
                    </a>
                  </p>
                </div>
                <div>
                  <h4 className="font-medium text-foreground mb-2">Content Responsibility</h4>
                  <p>
                    You are solely responsible for your User Content and the consequences of posting it. The creators of the libr framework 
                    do not endorse any User Content or any opinion, recommendation, or advice expressed therein.
                  </p>
                </div>
              </div>
            </section>

            <section className="bg-card/30 rounded-xl p-6 border border-border">
              <h3 className="text-xl font-semibold text-foreground mb-4">Network Participation and Incentivization</h3>
              <div className="space-y-4 text-muted-foreground">
                <p>
                  libr includes mechanisms for incentivizing Database Nodes and Moderators through cryptographically verifiable 
                  Proof of Service and Proof of Storage. Participation in these systems is voluntary and subject to:
                </p>
                <ul className="space-y-2">
                  <li className="flex items-start">
                    <div className="w-2 h-2 bg-libr-secondary rounded-full mt-2 mr-3 flex-shrink-0"></div>
                    <span>Maintaining service quality and availability standards</span>
                  </li>
                  <li className="flex items-start">
                    <div className="w-2 h-2 bg-libr-secondary rounded-full mt-2 mr-3 flex-shrink-0"></div>
                    <span>Adhering to community governance decisions</span>
                  </li>
                  <li className="flex items-start">
                    <div className="w-2 h-2 bg-libr-secondary rounded-full mt-2 mr-3 flex-shrink-0"></div>
                    <span>Providing verifiable proofs of legitimate service provision</span>
                  </li>
                  <li className="flex items-start">
                    <div className="w-2 h-2 bg-libr-secondary rounded-full mt-2 mr-3 flex-shrink-0"></div>
                    <span>Compliance with the consensus protocols and state transaction validation</span>
                  </li>
                </ul>
              </div>
            </section>

            <section className="bg-gradient-to-r from-yellow-500/10 to-orange-500/10 rounded-xl p-6 border border-yellow-500/20">
              <div className="flex items-center mb-4">
                <AlertTriangle className="w-6 h-6 text-yellow-500 mr-3" />
                <h3 className="text-xl font-semibold text-foreground">Intellectual Property Rights</h3>
              </div>
              <div className="space-y-4 text-muted-foreground">
                <p>
                  All rights, title, and interest in and to the libr framework itself (including its source code, logos, and branding) 
                  are and will remain the exclusive property of libr and its licensors. The libr protocol source code is licensed under 
                  the <strong className="text-foreground">Apache License, Version 2.0</strong>.
                </p>
                <p>
                  You are prohibited from using libr's trademarks, logos, or branding without our prior written consent.
                </p>
              </div>
            </section>

            <section className="bg-card/30 rounded-xl p-6 border border-border">
              <h3 className="text-xl font-semibold text-foreground mb-4">Disclaimers and Limitation of Liability</h3>
              <div className="space-y-4 text-muted-foreground">
                <div className="bg-slate-100/10 border border-slate-500/20 rounded-lg p-4">
                  <h4 className="font-medium text-foreground mb-2">"AS IS" Service</h4>
                  <p className="text-sm">
                    The libr framework is provided "AS IS" and "AS AVAILABLE" without warranties of any kind, express or implied. 
                    We do not guarantee uninterrupted service, permanent data storage, or immunity from security breaches.
                  </p>
                </div>
                <div>
                  <h4 className="font-medium text-foreground mb-2">Risk of Use</h4>
                  <p>
                    You use libr at your own risk, including the risk of encountering content you may find offensive, harmful, or inaccurate. 
                    Due to the decentralized nature, content moderation may vary between communities.
                  </p>
                </div>
                <div>
                  <h4 className="font-medium text-foreground mb-2">Limitation of Liability</h4>
                  <p>
                    To the fullest extent permitted by law, no libr contributor, node operator, or developer shall be liable for any direct or 
                    indirect damages arising from your use or inability to use the Platform. This includes, but is not limited to, data loss, 
                    content exposure, disputes arising from moderation decisions, or network failures.
                  </p>
                </div>
                <div>
                  <h4 className="font-medium text-foreground mb-2">Third-Party Components</h4>
                  <p>
                    libr integrates with various third-party services (including blockchain networks, NLP services, and distributed storage systems). 
                    We are not responsible for the availability, functionality, or security of these external services.
                  </p>
                </div>
              </div>
            </section>

            <section className="bg-card/30 rounded-xl p-6 border border-border">
              <h3 className="text-xl font-semibold text-foreground mb-4">Privacy and Data Handling</h3>
              <div className="space-y-4 text-muted-foreground">
                <p>
                  Due to libr's decentralized architecture, your data may be stored across multiple independent Database Nodes. 
                  While the protocol employs cryptographic security measures, you should be aware that:
                </p>
                <ul className="space-y-2">
                  <li className="flex items-start">
                    <div className="w-2 h-2 bg-libr-secondary rounded-full mt-2 mr-3 flex-shrink-0"></div>
                    <span>Message metadata (such as timestamps) may be publicly visible</span>
                  </li>
                  <li className="flex items-start">
                    <div className="w-2 h-2 bg-libr-secondary rounded-full mt-2 mr-3 flex-shrink-0"></div>
                    <span>Content approved by moderators becomes part of the distributed network</span>
                  </li>
                  <li className="flex items-start">
                    <div className="w-2 h-2 bg-libr-secondary rounded-full mt-2 mr-3 flex-shrink-0"></div>
                    <span>Private key management is entirely your responsibility</span>
                  </li>
                  <li className="flex items-start">
                    <div className="w-2 h-2 bg-libr-secondary rounded-full mt-2 mr-3 flex-shrink-0"></div>
                    <span>The protocol is designed for public forum use, not private communications</span>
                  </li>
                </ul>
              </div>
            </section>

            <section className="bg-card/30 rounded-xl p-6 border border-border">
              <h3 className="text-xl font-semibold text-foreground mb-4">Governing Law</h3>
              <p className="text-muted-foreground">
                These Terms shall be governed by and construed in accordance with the laws of India, 
                without regard to its conflict of law principles.
              </p>
            </section>

            <section className="bg-card/30 rounded-xl p-6 border border-border">
              <h3 className="text-xl font-semibold text-foreground mb-4">Changes to Terms</h3>
              <p className="text-muted-foreground">
                We may update these Terms periodically to reflect changes in the protocol, legal requirements, or community governance decisions. 
                Any updates will be communicated via official community channels and posted on our website. Your continued use of libr after 
                the changes are posted constitutes your acceptance of the revised Terms.
              </p>
            </section>

            <section className="bg-card/30 rounded-xl p-6 border border-border">
              <h3 className="text-xl font-semibold text-foreground mb-4">Contact and Governance</h3>
              <p className="text-muted-foreground mb-4">
                As libr is decentralized, there is no single corporate operator. For governance-related queries, appeals, or issues, 
                please engage with the relevant Community's official channels, as specified in the libr protocol documentation.
              </p>
              <div className="space-y-2 text-muted-foreground">
                <p>
                  <strong className="text-foreground">Email: </strong>
                  <a href="mailto:libr.forum@gmail.com" target="_blank" className="text-libr-accent1 hover:text-libr-accent2 hover:underline">
                    libr.forum@gmail.com
                  </a>
                </p>
                <p>
                  <strong className="text-foreground">GitHub: </strong>
                  <a href="https://github.com/libr-forum/libr" target="_blank" className="text-libr-accent1 hover:text-libr-accent2 hover:underline">
                    https://github.com/libr-forum/libr
                  </a>
                </p>
                <p>
                  <strong className="text-foreground">Website: </strong>
                  <a href="https://libr-ashen.vercel.app/" target="_blank" className="text-libr-accent1 hover:text-libr-accent2 hover:underline">
                    https://libr-ashen.vercel.app/
                  </a>
                </p>
                <p>
                  <strong className="text-foreground">Apache 2.0 License: </strong>
                  <a href="http://www.apache.org/licenses/LICENSE-2.0" target="_blank" className="text-libr-accent1 hover:text-libr-accent2 hover:underline">
                    View Full License Text
                  </a>
                </p>
                <p>
                  <strong className="text-foreground">Privacy Policy: </strong>
                  <a href="/privacy-policy" className="text-libr-accent1 hover:text-libr-accent2 hover:underline">
                    View our Privacy Policy
                  </a>
                </p>
                <p>
                  <strong className="text-foreground">EULA: </strong>
                  <a href="/eula" className="text-libr-accent1 hover:text-libr-accent2 hover:underline">
                    View our End User License Agreement
                  </a>
                </p>
              </div>
            </section>

            <section className="bg-card/30 rounded-xl p-6 border border-border border-t-2 border-t-libr-accent1">
              <p className="text-sm text-muted-foreground text-center">
                By using libr, you acknowledge that you have read, understood, and agree to be bound by these Terms and Conditions.
              </p>
            </section>

          </div>
        </motion.div>
      </div>
      <Footer />
    </div>
  );
};

export default TermsAndConditions;
