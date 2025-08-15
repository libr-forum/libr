import React, { useEffect } from 'react';
import { motion } from 'framer-motion';
import { Link } from 'react-router-dom';
import { ArrowLeft, FileCheck, Download, AlertCircle, CheckCircle, XCircle, Settings } from 'lucide-react';
import { Header } from '../components/LandingPageSections';
import { Footer } from '../components/LandingPageExtended';

interface EULAProps {
  isDarkMode?: boolean;
  toggleTheme?: () => void;
}

const EULA: React.FC<EULAProps> = ({ isDarkMode = true, toggleTheme = () => {} }) => {
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
              className="inline-flex items-center text-libr-foreground hover:text-libr-secondary transition-colors mb-6"
            >
              <ArrowLeft className="w-4 h-4 mr-2" />
              Back to Home
            </Link>
            
            <div className="flex items-center mb-4 text-libr-foreground">
              <FileCheck className="w-8 h-8 mr-3" />
              <h1 className="text-4xl font-bold">
                End User License Agreement
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
              <Download className="w-6 h-6 text-libr-secondary mr-3" />
              <h2 className="text-2xl font-semibold text-foreground">Agreement Overview</h2>
            </div>
            <p className="text-muted-foreground leading-relaxed mb-4">
              This End User License Agreement (EULA) is a legally binding contract between you and the libr project maintainers. 
              By downloading, installing, accessing, or using the libr software, documentation, or related materials, you agree to be bound by these terms.
            </p>
            <p className="text-muted-foreground leading-relaxed">
              libr is open-source software distributed under the Apache License 2.0, which provides you with extensive rights to use, 
              modify, and distribute the software. If you do not agree to these terms, you may not use the software.
            </p>
          </div>

          <div className="space-y-8">
            <section className="bg-card/30 rounded-xl p-6 border border-border">
              <div className="flex items-center mb-4">
                <CheckCircle className="w-6 h-6 text-green-500 mr-3" />
                <h3 className="text-xl font-semibold text-foreground">License Grant</h3>
              </div>
              <div className="space-y-4 text-muted-foreground">
                <p>
                  Subject to your compliance with this Agreement and the Apache License 2.0, we grant you a perpetual, 
                  worldwide, non-exclusive, no-charge, royalty-free, irrevocable license to use libr software.
                </p>
                <div>
                  <h4 className="font-medium text-foreground mb-2">Scope of Use & Permitted Activities:</h4>
                  <ul className="space-y-2">
                    <li className="flex items-start">
                      <CheckCircle className="w-4 h-4 text-green-500 mt-1 mr-3 flex-shrink-0" />
                      <span>Download, install, and run the software on your devices for personal or commercial use</span>
                    </li>
                    <li className="flex items-start">
                      <CheckCircle className="w-4 h-4 text-green-500 mt-1 mr-3 flex-shrink-0" />
                      <span>Participate in, moderate, and operate decentralized, censorship-resilient communities</span>
                    </li>
                    <li className="flex items-start">
                      <CheckCircle className="w-4 h-4 text-green-500 mt-1 mr-3 flex-shrink-0" />
                      <span>Run client, moderator, database, and relay nodes in compliance with community governance</span>
                    </li>
                    <li className="flex items-start">
                      <CheckCircle className="w-4 h-4 text-green-500 mt-1 mr-3 flex-shrink-0" />
                      <span>Modify the source code for your own purposes</span>
                    </li>
                    <li className="flex items-start">
                      <CheckCircle className="w-4 h-4 text-green-500 mt-1 mr-3 flex-shrink-0" />
                      <span>Distribute original or modified versions (subject to Apache 2.0 requirements)</span>
                    </li>
                    <li className="flex items-start">
                      <CheckCircle className="w-4 h-4 text-green-500 mt-1 mr-3 flex-shrink-0" />
                      <span>Contribute code, documentation, or translations under the project's open-source policy</span>
                    </li>
                  </ul>
                </div>
              </div>
            </section>

            <section className="bg-card/30 rounded-xl p-6 border border-border">
              <div className="flex items-center mb-4">
                <XCircle className="w-6 h-6 text-red-500 mr-3" />
                <h3 className="text-xl font-semibold text-foreground">Restrictions and Prohibited Uses</h3>
              </div>
              <div className="space-y-4 text-muted-foreground">
                <div>
                  <h4 className="font-medium text-foreground mb-2">You Must Not:</h4>
                  <ul className="space-y-2">
                    <li className="flex items-start">
                      <XCircle className="w-4 h-4 text-red-500 mt-1 mr-3 flex-shrink-0" />
                      <span>Remove or alter copyright notices, trademarks, or license information</span>
                    </li>
                    <li className="flex items-start">
                      <XCircle className="w-4 h-4 text-red-500 mt-1 mr-3 flex-shrink-0" />
                      <span>Use the software for unlawful, harmful, or abusive activities including harassment, spam, or malware distribution</span>
                    </li>
                    <li className="flex items-start">
                      <XCircle className="w-4 h-4 text-red-500 mt-1 mr-3 flex-shrink-0" />
                      <span>Operate nodes in ways that violate network stability or flood relays with excessive requests</span>
                    </li>
                    <li className="flex items-start">
                      <XCircle className="w-4 h-4 text-red-500 mt-1 mr-3 flex-shrink-0" />
                      <span>Violate privacy rights of other users or applicable data protection laws</span>
                    </li>
                    <li className="flex items-start">
                      <XCircle className="w-4 h-4 text-red-500 mt-1 mr-3 flex-shrink-0" />
                      <span>Use the libr name or logo without explicit permission for commercial purposes</span>
                    </li>
                    <li className="flex items-start">
                      <XCircle className="w-4 h-4 text-red-500 mt-1 mr-3 flex-shrink-0" />
                      <span>Distribute modified versions without proper attribution as required by Apache 2.0</span>
                    </li>
                  </ul>
                </div>
              </div>
            </section>

            <section className="bg-gradient-to-r from-blue-500/10 to-purple-500/10 rounded-xl p-6 border border-blue-500/20">
              <div className="flex items-center mb-4">
                <Settings className="w-6 h-6 text-libr-secondary mr-3" />
                <h3 className="text-xl font-semibold text-foreground">Apache 2.0 License & Open Source Components</h3>
              </div>
              <div className="space-y-4 text-muted-foreground">
                <p>
                  libr is distributed under the Apache License 2.0, which provides you with broad rights to use, modify, and distribute the software.
                  The complete license text is available in the repository's LICENSE file.
                </p>
                <div>
                  <h4 className="font-medium text-foreground mb-2">Key Apache 2.0 Provisions:</h4>
                  <ul className="space-y-2">
                    <li className="flex items-start">
                      <div className="w-2 h-2 bg-libr-secondary rounded-full mt-2 mr-3 flex-shrink-0"></div>
                      <span><strong className="text-foreground">Patent Grant:</strong> Contributors grant you patent rights for their contributions</span>
                    </li>
                    <li className="flex items-start">
                      <div className="w-2 h-2 bg-libr-secondary rounded-full mt-2 mr-3 flex-shrink-0"></div>
                      <span><strong className="text-foreground">Attribution:</strong> Modified versions must include NOTICE of changes</span>
                    </li>
                    <li className="flex items-start">
                      <div className="w-2 h-2 bg-libr-secondary rounded-full mt-2 mr-3 flex-shrink-0"></div>
                      <span><strong className="text-foreground">Trademark Protection:</strong> No license to use libr trademarks</span>
                    </li>
                    <li className="flex items-start">
                      <div className="w-2 h-2 bg-libr-secondary rounded-full mt-2 mr-3 flex-shrink-0"></div>
                      <span><strong className="text-foreground">Disclaimer:</strong> Software provided "AS IS" without warranties</span>
                    </li>
                  </ul>
                </div>
                <p>
                  Third-party libraries incorporated into libr are subject to their respective licenses. 
                  See the repository's NOTICE file and dependency documentation for complete details.
                </p>
              </div>
            </section>

            <section className="bg-card/30 rounded-xl p-6 border border-border">
              <div className="flex items-center mb-4">
                <AlertCircle className="w-6 h-6 text-libr-secondary mr-3" />
                <h3 className="text-xl font-semibold text-foreground">Intellectual Property Rights</h3>
              </div>
              <div className="space-y-4 text-muted-foreground">
                <p>
                  While the libr software source code is open-source under Apache 2.0, certain intellectual property rights are retained:
                </p>
                <ul className="space-y-3">
                  <li className="flex items-start">
                    <div className="w-2 h-2 bg-libr-secondary rounded-full mt-2 mr-3 flex-shrink-0"></div>
                    <span><strong className="text-foreground">Source Code:</strong> Licensed under Apache 2.0, freely usable and modifiable</span>
                  </li>
                  <li className="flex items-start">
                    <div className="w-2 h-2 bg-libr-secondary rounded-full mt-2 mr-3 flex-shrink-0"></div>
                    <span><strong className="text-foreground">Trademarks:</strong> The "libr" name and logo are trademarks, not covered by the software license</span>
                  </li>
                  <li className="flex items-start">
                    <div className="w-2 h-2 bg-libr-secondary rounded-full mt-2 mr-3 flex-shrink-0"></div>
                    <span><strong className="text-foreground">Documentation:</strong> Project documentation follows the same Apache 2.0 terms</span>
                  </li>
                  <li className="flex items-start">
                    <div className="w-2 h-2 bg-libr-secondary rounded-full mt-2 mr-3 flex-shrink-0"></div>
                    <span><strong className="text-foreground">Contributions:</strong> Your contributions are licensed under Apache 2.0 terms</span>
                  </li>
                </ul>
              </div>
            </section>

            <section className="bg-card/30 rounded-xl p-6 border border-border">
              <div className="flex items-center mb-4">
                <Settings className="w-6 h-6 text-libr-secondary mr-3" />
                <h3 className="text-xl font-semibold text-foreground">Community Governance & Moderation</h3>
              </div>
              <div className="space-y-4 text-muted-foreground">
                <p>
                  You acknowledge that libr enables decentralized, community-driven moderation. 
                  The libr project maintainers do not centrally control content or enforce community policies.
                </p>
                <ul className="space-y-3">
                  <li className="flex items-start">
                    <div className="w-2 h-2 bg-libr-secondary rounded-full mt-2 mr-3 flex-shrink-0"></div>
                    <span>Each community defines its own rules, governance, and moderation policies</span>
                  </li>
                  <li className="flex items-start">
                    <div className="w-2 h-2 bg-libr-secondary rounded-full mt-2 mr-3 flex-shrink-0"></div>
                    <span>As a node operator or participant, you agree to abide by policies of communities you join</span>
                  </li>
                  <li className="flex items-start">
                    <div className="w-2 h-2 bg-libr-secondary rounded-full mt-2 mr-3 flex-shrink-0"></div>
                    <span>Node operators are responsible for their own compliance with local laws and regulations</span>
                  </li>
                  <li className="flex items-start">
                    <div className="w-2 h-2 bg-libr-secondary rounded-full mt-2 mr-3 flex-shrink-0"></div>
                    <span>Content moderation decisions are made by community moderators, not libr maintainers</span>
                  </li>
                </ul>
              </div>
            </section>

            <section className="bg-card/30 rounded-xl p-6 border border-border">
              <div className="flex items-center mb-4">
                <CheckCircle className="w-6 h-6 text-libr-secondary mr-3" />
                <h3 className="text-xl font-semibold text-foreground">Privacy & Data Handling</h3>
              </div>
              <div className="space-y-4 text-muted-foreground">
                <p>
                  The libr software is designed with privacy in mind, but the decentralized nature has implications:
                </p>
                <ul className="space-y-3">
                  <li className="flex items-start">
                    <div className="w-2 h-2 bg-libr-secondary rounded-full mt-2 mr-3 flex-shrink-0"></div>
                    <span>Private keys and user credentials are stored locally on your device only</span>
                  </li>
                  <li className="flex items-start">
                    <div className="w-2 h-2 bg-libr-secondary rounded-full mt-2 mr-3 flex-shrink-0"></div>
                    <span>Messages and posts are published to independent relay servers with their own privacy policies</span>
                  </li>
                  <li className="flex items-start">
                    <div className="w-2 h-2 bg-libr-secondary rounded-full mt-2 mr-3 flex-shrink-0"></div>
                    <span>The software itself does not collect personal data for libr maintainers</span>
                  </li>
                  <li className="flex items-start">
                    <div className="w-2 h-2 bg-libr-secondary rounded-full mt-2 mr-3 flex-shrink-0"></div>
                    <span>Third-party relay operators may have their own data collection practices</span>
                  </li>
                </ul>
                <p className="mt-4">
                  For detailed privacy information, please refer to our separate Privacy Policy.
                </p>
              </div>
            </section>

            <section className="bg-card/30 rounded-xl p-6 border border-border">
              <div className="flex items-center mb-4">
                <AlertCircle className="w-6 h-6 text-orange-500 mr-3" />
                <h3 className="text-xl font-semibold text-foreground">Software Updates</h3>
              </div>
              <div className="space-y-4 text-muted-foreground">
                <p>
                  We may provide updates, patches, and new versions of the software. Updates may be delivered 
                  automatically or require manual installation, depending on your configuration.
                </p>
                <div>
                  <h4 className="font-medium text-foreground mb-2">Update Types:</h4>
                  <ul className="space-y-2">
                    <li className="flex items-start">
                      <div className="w-2 h-2 bg-orange-500 rounded-full mt-2 mr-3 flex-shrink-0"></div>
                      <span><strong className="text-foreground">Security Updates:</strong> Critical patches for security vulnerabilities</span>
                    </li>
                    <li className="flex items-start">
                      <div className="w-2 h-2 bg-orange-500 rounded-full mt-2 mr-3 flex-shrink-0"></div>
                      <span><strong className="text-foreground">Bug Fixes:</strong> Corrections for identified issues</span>
                    </li>
                    <li className="flex items-start">
                      <div className="w-2 h-2 bg-orange-500 rounded-full mt-2 mr-3 flex-shrink-0"></div>
                      <span><strong className="text-foreground">Feature Updates:</strong> New functionality and improvements</span>
                    </li>
                    <li className="flex items-start">
                      <div className="w-2 h-2 bg-orange-500 rounded-full mt-2 mr-3 flex-shrink-0"></div>
                      <span><strong className="text-foreground">Protocol Updates:</strong> Changes to the underlying libr protocol</span>
                    </li>
                  </ul>
                </div>
              </div>
            </section>

            <section className="bg-card/30 rounded-xl p-6 border border-border">
              <h3 className="text-xl font-semibold text-foreground mb-4">Data and Privacy</h3>
              <div className="space-y-4 text-muted-foreground">
                <p>
                  This EULA governs the software itself. Data collection, usage, and privacy are covered by our 
                  separate Privacy Policy. Please review both documents to understand your rights and obligations.
                </p>
                <p>
                  The software may collect telemetry data to improve functionality and identify issues. 
                  You can typically disable telemetry in the application settings.
                </p>
              </div>
            </section>

            <section className="bg-card/30 rounded-xl p-6 border border-border">
              <h3 className="text-xl font-semibold text-foreground mb-4">Disclaimer and Limitation of Liability</h3>
              <div className="space-y-4 text-muted-foreground">
                <p>
                  <strong className="text-foreground">Warranty Disclaimer:</strong> The software is provided "as is" 
                  without warranties of any kind. We do not guarantee that the software will be error-free or 
                  suitable for any particular purpose.
                </p>
                <p>
                  <strong className="text-foreground">Limitation of Liability:</strong> To the maximum extent permitted 
                  by law, we shall not be liable for any damages arising from your use of the software, 
                  including but not limited to data loss, business interruption, or other economic losses.
                </p>
              </div>
            </section>

            <section className="bg-card/30 rounded-xl p-6 border border-border">
              <h3 className="text-xl font-semibold text-foreground mb-4">Termination</h3>
              <div className="space-y-4 text-muted-foreground">
                <p>
                  This Agreement is effective until terminated. Due to the open-source nature of libr under Apache 2.0, 
                  your rights to use the software are perpetual unless you violate the license terms.
                </p>
                <p>
                  However, this EULA may be terminated if you breach any of its terms, particularly those related to:
                </p>
                <ul className="space-y-2 ml-4">
                  <li>• Unlawful use of the software</li>
                  <li>• Violation of network stability or community policies</li>
                  <li>• Improper use of libr trademarks</li>
                  <li>• Distribution without proper Apache 2.0 attribution</li>
                </ul>
                <p>
                  Upon termination of this EULA, your Apache 2.0 license rights continue, but you must cease any activities 
                  that violate this agreement and remove any unauthorized trademark usage.
                </p>
              </div>
            </section>

            <section className="bg-card/30 rounded-xl p-6 border border-border">
              <h3 className="text-xl font-semibold text-foreground mb-4">Governing Law & Dispute Resolution</h3>
              <div className="space-y-4 text-muted-foreground">
                <p>
                  This EULA is governed by the laws of India, where the libr project maintainers are primarily based, 
                  without regard to conflict of law principles.
                </p>
                <p>
                  Any disputes arising from this EULA will be resolved through:
                </p>
                <ul className="space-y-2 ml-4">
                  <li>• First, good faith negotiation between parties</li>
                  <li>• If unsuccessful, binding arbitration under Indian arbitration laws</li>
                  <li>• As a last resort, appropriate courts in India</li>
                </ul>
                <p>
                  For technical disputes related to the Apache 2.0 license itself, standard open-source dispute resolution 
                  mechanisms apply.
                </p>
              </div>
            </section>

            <div className="bg-gradient-to-r from-libr-accent1/10 to-libr-accent2/10 rounded-xl p-6 border border-libr-accent1/20">
              <h4 className="text-lg font-semibold text-foreground mb-3">Acknowledgment & Acceptance</h4>
              <p className="text-muted-foreground mb-3">
                By installing, copying, or using the libr software, you acknowledge that you have read, 
                understood, and agree to be bound by both:
              </p>
              <ul className="space-y-1 text-muted-foreground">
                <li>• This End User License Agreement (EULA)</li>
                <li>• The Apache License 2.0 under which the software is distributed</li>
              </ul>
              <p className="text-muted-foreground mt-3 text-sm">
                If you do not agree to these terms, you must not install, distribute, or use the software.
              </p>
            </div>

            <section className="bg-card/30 rounded-xl p-6 border border-border">
              <h3 className="text-xl font-semibold text-foreground mb-4">Contact and Governance</h3>
              <p className="text-muted-foreground mb-4">
                As libr is decentralized, there is no single corporate operator. For governance-related queries, appeals, or issues, 
                please engage with the relevant Community's official channels, as specified in the libr protocol documentation.
              </p>
              <div className="space-y-2 text-muted-foreground">
                <p><strong className="text-foreground">Email:</strong> <a href="mailto:libr.forum@gmail.com" target="_blank" className="text-libr-accent1 hover:text-libr-accent2 hover:underline">libr.forum@gmail.com</a></p>
                <p><strong className="text-foreground">GitHub:</strong> <a href="https://github.com/libr-forum/libr" target="_blank" className="text-libr-accent1 hover:text-libr-accent2 hover:underline">https://github.com/libr-forum/libr</a></p>
                <p><strong className="text-foreground">Website:</strong> <a href="https://libr-ashen.vercel.app/" target="_blank" className="text-libr-accent1 hover:text-libr-accent2 hover:underline">https://libr-ashen.vercel.app/</a></p>
                <p><strong className="text-foreground">Apache 2.0 License:</strong> <a href="http://www.apache.org/licenses/LICENSE-2.0" target="_blank" className="text-libr-accent1 hover:text-libr-accent2 hover:underline">View Full License Text</a></p>
                <p><strong className="text-foreground">Terms & Conditions:</strong> <a href="/terms-and-conditions" className="text-libr-accent1 hover:text-libr-accent2 hover:underline">View our Terms & Conditions</a></p>
                <p><strong className="text-foreground">Privacy Policy:</strong> <a href="/privacy-policy" className="text-libr-accent1 hover:text-libr-accent2 hover:underline">View our Privacy Policy</a></p>
              </div>
            </section>

          </div>
        </motion.div>
      </div>
      <Footer />
    </div>
  );
};

export default EULA;
