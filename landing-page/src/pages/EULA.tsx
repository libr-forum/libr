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
              className="inline-flex items-center text-libr-accent1 hover:text-libr-accent2 transition-colors mb-6"
            >
              <ArrowLeft className="w-4 h-4 mr-2" />
              Back to Home
            </Link>
            
            <div className="flex items-center mb-4">
              <FileCheck className="w-8 h-8 text-libr-accent1 mr-3" />
              <h1 className="text-4xl font-bold bg-gradient-to-r from-libr-accent1 to-libr-accent2 bg-clip-text text-transparent">
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
              <Download className="w-6 h-6 text-libr-accent1 mr-3" />
              <h2 className="text-2xl font-semibold text-foreground">Software License Agreement</h2>
            </div>
            <p className="text-muted-foreground leading-relaxed">
              This End User License Agreement (EULA) governs your use of libr client applications, 
              protocol implementations, and associated software components distributed by the libr project.
            </p>
          </div>

          <div className="space-y-8">
            <section className="bg-card/30 rounded-xl p-6 border border-border">
              <div className="flex items-center mb-4">
                <CheckCircle className="w-6 h-6 text-green-500 mr-3" />
                <h3 className="text-xl font-semibold text-foreground">Grant of License</h3>
              </div>
              <div className="space-y-4 text-muted-foreground">
                <p>
                  Subject to the terms of this EULA, we grant you a non-exclusive, non-transferable, 
                  revocable license to use the libr software applications for personal or commercial purposes.
                </p>
                <div>
                  <h4 className="font-medium text-foreground mb-2">You May:</h4>
                  <ul className="space-y-2">
                    <li className="flex items-start">
                      <CheckCircle className="w-4 h-4 text-green-500 mt-1 mr-3 flex-shrink-0" />
                      <span>Install and use the software on your devices</span>
                    </li>
                    <li className="flex items-start">
                      <CheckCircle className="w-4 h-4 text-green-500 mt-1 mr-3 flex-shrink-0" />
                      <span>Make backup copies for personal use</span>
                    </li>
                    <li className="flex items-start">
                      <CheckCircle className="w-4 h-4 text-green-500 mt-1 mr-3 flex-shrink-0" />
                      <span>Modify the software (where open-source licenses permit)</span>
                    </li>
                    <li className="flex items-start">
                      <CheckCircle className="w-4 h-4 text-green-500 mt-1 mr-3 flex-shrink-0" />
                      <span>Use the software for commercial purposes</span>
                    </li>
                    <li className="flex items-start">
                      <CheckCircle className="w-4 h-4 text-green-500 mt-1 mr-3 flex-shrink-0" />
                      <span>Distribute modified versions (under applicable open-source terms)</span>
                    </li>
                  </ul>
                </div>
              </div>
            </section>

            <section className="bg-card/30 rounded-xl p-6 border border-border">
              <div className="flex items-center mb-4">
                <XCircle className="w-6 h-6 text-red-500 mr-3" />
                <h3 className="text-xl font-semibold text-foreground">Restrictions</h3>
              </div>
              <div className="space-y-4 text-muted-foreground">
                <div>
                  <h4 className="font-medium text-foreground mb-2">You May Not:</h4>
                  <ul className="space-y-2">
                    <li className="flex items-start">
                      <XCircle className="w-4 h-4 text-red-500 mt-1 mr-3 flex-shrink-0" />
                      <span>Remove copyright notices or license information</span>
                    </li>
                    <li className="flex items-start">
                      <XCircle className="w-4 h-4 text-red-500 mt-1 mr-3 flex-shrink-0" />
                      <span>Reverse engineer proprietary components (where applicable)</span>
                    </li>
                    <li className="flex items-start">
                      <XCircle className="w-4 h-4 text-red-500 mt-1 mr-3 flex-shrink-0" />
                      <span>Use the software to violate laws or regulations</span>
                    </li>
                    <li className="flex items-start">
                      <XCircle className="w-4 h-4 text-red-500 mt-1 mr-3 flex-shrink-0" />
                      <span>Distribute malware or harmful modifications</span>
                    </li>
                    <li className="flex items-start">
                      <XCircle className="w-4 h-4 text-red-500 mt-1 mr-3 flex-shrink-0" />
                      <span>Claim ownership of the libr trademark or branding</span>
                    </li>
                  </ul>
                </div>
              </div>
            </section>

            <section className="bg-gradient-to-r from-blue-500/10 to-purple-500/10 rounded-xl p-6 border border-blue-500/20">
              <div className="flex items-center mb-4">
                <Settings className="w-6 h-6 text-blue-500 mr-3" />
                <h3 className="text-xl font-semibold text-foreground">Open Source Components</h3>
              </div>
              <div className="space-y-4 text-muted-foreground">
                <p>
                  libr incorporates various open-source components, each governed by their respective licenses. 
                  These may include but are not limited to:
                </p>
                <ul className="space-y-2">
                  <li className="flex items-start">
                    <div className="w-2 h-2 bg-blue-500 rounded-full mt-2 mr-3 flex-shrink-0"></div>
                    <span><strong className="text-foreground">MIT License</strong> components</span>
                  </li>
                  <li className="flex items-start">
                    <div className="w-2 h-2 bg-blue-500 rounded-full mt-2 mr-3 flex-shrink-0"></div>
                    <span><strong className="text-foreground">Apache 2.0</strong> licensed libraries</span>
                  </li>
                  <li className="flex items-start">
                    <div className="w-2 h-2 bg-blue-500 rounded-full mt-2 mr-3 flex-shrink-0"></div>
                    <span><strong className="text-foreground">BSD</strong> licensed software</span>
                  </li>
                  <li className="flex items-start">
                    <div className="w-2 h-2 bg-blue-500 rounded-full mt-2 mr-3 flex-shrink-0"></div>
                    <span><strong className="text-foreground">GPL</strong> components (where applicable)</span>
                  </li>
                </ul>
                <p>
                  The full list of dependencies and their licenses can be found in the software's documentation 
                  and source code repositories.
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
                      <div className="w-2 h-2 bg-libr-accent1 rounded-full mt-2 mr-3 flex-shrink-0"></div>
                      <span><strong className="text-foreground">Security Updates:</strong> Critical patches for security vulnerabilities</span>
                    </li>
                    <li className="flex items-start">
                      <div className="w-2 h-2 bg-libr-accent1 rounded-full mt-2 mr-3 flex-shrink-0"></div>
                      <span><strong className="text-foreground">Bug Fixes:</strong> Corrections for identified issues</span>
                    </li>
                    <li className="flex items-start">
                      <div className="w-2 h-2 bg-libr-accent1 rounded-full mt-2 mr-3 flex-shrink-0"></div>
                      <span><strong className="text-foreground">Feature Updates:</strong> New functionality and improvements</span>
                    </li>
                    <li className="flex items-start">
                      <div className="w-2 h-2 bg-libr-accent1 rounded-full mt-2 mr-3 flex-shrink-0"></div>
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
                  This license is effective until terminated. Your rights under this license will terminate 
                  automatically if you fail to comply with any of its terms.
                </p>
                <p>
                  Upon termination, you must cease all use of the software and destroy all copies in your possession. 
                  Provisions regarding disclaimers, limitations of liability, and governing law survive termination.
                </p>
              </div>
            </section>

            <section className="bg-card/30 rounded-xl p-6 border border-border">
              <h3 className="text-xl font-semibold text-foreground mb-4">Governing Law</h3>
              <p className="text-muted-foreground">
                This EULA is governed by the laws of the jurisdiction where the libr project is based, 
                without regard to conflict of law principles. Any disputes will be resolved in the 
                appropriate courts of that jurisdiction.
              </p>
            </section>

            <section className="bg-card/30 rounded-xl p-6 border border-border">
              <h3 className="text-xl font-semibold text-foreground mb-4">Contact Information</h3>
              <p className="text-muted-foreground mb-4">
                For questions about this EULA or licensing issues, contact us at:
              </p>
              <div className="space-y-2 text-muted-foreground">
                <p><strong className="text-foreground">Email:</strong> libr.forum@gmail.com</p>
                <p><strong className="text-foreground">GitHub:</strong> https://github.com/devlup-labs/Libr</p>
                <p><strong className="text-foreground">License Information:</strong> Available in software documentation</p>
              </div>
            </section>

            <div className="bg-gradient-to-r from-libr-accent1/10 to-libr-accent2/10 rounded-xl p-6 border border-libr-accent1/20">
              <p className="text-center text-muted-foreground">
                By installing, copying, or using the libr software, you acknowledge that you have read, 
                understood, and agree to be bound by the terms of this End User License Agreement.
              </p>
            </div>
          </div>
        </motion.div>
      </div>
      <Footer />
    </div>
  );
};

export default EULA;
