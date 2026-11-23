import React, { useEffect, useState } from 'react';
import { motion } from 'framer-motion';
import { Link } from 'react-router-dom';
import { ArrowLeft, Copy, CopyCheck } from 'lucide-react';
import { Header } from '../components/LandingPageSections';
import { Footer } from '../components/LandingPageExtended';
import { FaApple, FaLinux, FaWindows } from 'react-icons/fa';

interface InstallProps {
  isDarkMode?: boolean;
  toggleTheme?: () => void;
}

const CopyButton: React.FC<{ text: string }> = ({ text }) => {
  const [copied, setCopied] = useState(false);
  const handleCopy = async () => {
    await navigator.clipboard.writeText(text);
    setCopied(true);
    setTimeout(() => setCopied(false), 1200);
  };
  return (
    <button
      onClick={handleCopy}
      className="absolute top-2 right-2 text-libr-secondary px-2 py-1 rounded"
      aria-label="Copy code"
      type="button"
    >
      {copied ? <CopyCheck className='bg-gray-100 dark:bg-gray-800'/> : <Copy className='bg-gray-100 dark:bg-gray-800'/>}
    </button>
  );
};

const Install: React.FC<InstallProps> = ({ isDarkMode = true, toggleTheme = () => {} }) => {
  useEffect(() => {
    window.scrollTo(0, 0);
  }, []);

  const [selectedOS, setSelectedOS] = useState<'linux' | 'macos' | 'windows'>('windows');
  const [latestRelease, setLatestRelease] = useState<{ tag: string, date: string, url: string } | null>(null);

  const osOptions = [
    {
      key: 'windows',
      label: 'Windows',
      icon: <FaWindows />,
    },
    {
      key: 'linux',
      label: 'Linux',
      icon: <FaLinux />,
    },
    {
      key: 'macos',
      label: 'macOS',
      icon: <FaApple />,
    },
  ];

  useEffect(() => {
    fetch('https://api.github.com/repos/libr-forum/Libr/releases/latest')
      .then(res => res.json())
      .then(data => {
        setLatestRelease({
          tag: data.tag_name,
          date: new Date(data.published_at).toLocaleDateString(),
          url: data.assets?.[0]?.browser_download_url || data.html_url
        });
      });
  }, []);

  const getDownloadUrl = () => {
    if (!latestRelease) return "#";
    const assetNames: Record<typeof selectedOS, string> = {
      windows: "libr-windows-amd64.exe",
      linux: "libr-linux-amd64",
      macos: "libr-darwin-amd64.out",
    };
    return `https://github.com/libr-forum/libr/releases/download/${latestRelease.tag}/${assetNames[selectedOS]}`;
  };

  const renderInstallContent = () => {
    switch (selectedOS) {
      case 'linux':
        return (
          <div>
            <h3 className="text-lg font-bold mt-4 mb-2 flex items-center gap-2">
              Option 1: APT Repository (Ubuntu/Debian)
              <span className="text-xs bg-libr-secondary text-libr-primary px-2 py-1 rounded">
                Recommended
              </span>
            </h3>
            <div className="relative mb-4">
              <CopyButton text={
                `# Add GPG key for repository verification
wget -qO- https://libr-forum.github.io/libr-apt-repo/libr-repo-key.gpg | sudo gpg --dearmor -o /usr/share/keyrings/libr-repo-key.gpg

# Add APT repository to sources
echo "deb [signed-by=/usr/share/keyrings/libr-repo-key.gpg] https://libr-forum.github.io/libr-apt-repo/ ./" | sudo tee /etc/apt/sources.list.d/libr.list

# Update package index and install
sudo apt update
sudo apt install libr`
              } />
              <pre className="bg-gray-100 dark:bg-gray-800 p-3 rounded text-sm overflow-x-auto">
                <code>
                  # Add GPG key for repository verification{'\n'}
                  wget -qO- https://libr-forum.github.io/libr-apt-repo/libr-repo-key.gpg | sudo gpg --dearmor -o /usr/share/keyrings/libr-repo-key.gpg{'\n\n'}
                  # Add APT repository to sources{'\n'}
                  echo "deb [signed-by=/usr/share/keyrings/libr-repo-key.gpg] https://libr-forum.github.io/libr-apt-repo/ ./" | sudo tee /etc/apt/sources.list.d/libr.list{'\n\n'}
                  # Update package index and install{'\n'}
                  sudo apt update{'\n'}
                  sudo apt install libr
                </code>
              </pre>
            </div>

            <h3 className="text-lg font-bold mt-6 mb-2">Option 2: Direct Download - All Distributions</h3>
            <ul className="list-disc ml-6 mb-2">
              <li>
                <span className="font-semibold">Ubuntu/Debian (.deb package):</span>
                <div className="relative mt-2 mb-2">
                  <CopyButton text={
                    `wget https://github.com/libr-forum/Libr/releases/download/v1.0.0-beta/libr_1.0.0-beta_amd64.deb
sudo dpkg -i libr_1.0.0-beta_amd64.deb
sudo apt-get install -f`
                  } />
                  <pre className="bg-gray-100 dark:bg-gray-800 p-3 rounded text-sm overflow-x-auto">
                    <code>
                      wget https://github.com/libr-forum/Libr/releases/download/v1.0.0-beta/libr_1.0.0-beta_amd64.deb{'\n'}
                      sudo dpkg -i libr_1.0.0-beta_amd64.deb{'\n'}
                      sudo apt-get install -f
                    </code>
                  </pre>
                </div>
              </li>
              <li>
                <span className="font-semibold">Fedora/RHEL/CentOS (.rpm package):</span>
                <div className="relative mt-2 mb-2">
                  <CopyButton text={
                    `wget https://github.com/libr-forum/Libr/releases/download/v1.0.0-beta/libr-1.0.0-beta-1.x86_64.rpm
sudo dnf install ./libr-1.0.0-beta-1.x86_64.rpm
# or on older systems: 
sudo yum install ./libr-1.0.0-beta-1.x86_64.rpm`
                  } />
                  <pre className="bg-gray-100 dark:bg-gray-800 p-3 rounded text-sm overflow-x-auto">
                    <code>
                      wget https://github.com/libr-forum/Libr/releases/download/v1.0.0-beta/libr-1.0.0-beta-1.x86_64.rpm{'\n'}
                      sudo dnf install ./libr-1.0.0-beta-1.x86_64.rpm{'\n'}
                      # or on older systems:{' '}
                      sudo yum install ./libr-1.0.0-beta-1.x86_64.rpm
                    </code>
                  </pre>
                </div>
              </li>
              <li>
                <span className="font-semibold">Arch Linux (.pkg.tar.zst package):</span>
                <div className="relative mt-2 mb-2">
                  <CopyButton text={
                    `wget https://github.com/libr-forum/Libr/releases/download/v1.0.0-beta/libr-1.0.0-beta-1-x86_64.pkg.tar.zst
sudo pacman -U libr-1.0.0-beta-1-x86_64.pkg.tar.zst`
                  } />
                  <pre className="bg-gray-100 dark:bg-gray-800 p-3 rounded text-sm overflow-x-auto">
                    <code>
                      wget https://github.com/libr-forum/Libr/releases/download/v1.0.0-beta/libr-1.0.0-beta-1-x86_64.pkg.tar.zst{'\n'}
                      sudo pacman -U libr-1.0.0-beta-1-x86_64.pkg.tar.zst
                    </code>
                  </pre>
                </div>
              </li>
            </ul>

            <h3 className="text-lg font-bold mt-6 mb-2">Option 3: Binary Installation</h3>
            <div className="relative mb-4">
              <CopyButton text={
                `wget https://github.com/libr-forum/Libr/releases/download/v1.0.0-beta/libr-linux-amd64
chmod +x libr-linux-amd64
sudo mv libr-linux-amd64 /usr/local/bin/libr`
              } />
              <pre className="bg-gray-100 dark:bg-gray-800 p-3 rounded text-sm overflow-x-auto">
                <code>
                  wget https://github.com/libr-forum/Libr/releases/download/v1.0.0-beta/libr-linux-amd64{'\n'}
                  chmod +x libr-linux-amd64{'\n'}
                  sudo mv libr-linux-amd64 /usr/local/bin/libr
                </code>
              </pre>
            </div>

            <h3 className="text-lg font-bold mt-6 mb-2">Solving WebKit Library Issues</h3>
            <ul className="list-disc ml-6 mb-2">
              <li>
                <span className="font-semibold">Ubuntu 24.04 (Noble) and newer:</span>
                <div className="relative mt-2 mb-2">
                  <CopyButton text={
                    `sudo apt update
sudo apt install -y libwebkit2gtk-4.1-0 libjavascriptcoregtk-4.1-0
sudo ln -sf /usr/lib/x86_64-linux-gnu/libwebkit2gtk-4.1.so.0 /usr/lib/x86_64-linux-gnu/libwebkit2gtk-4.0.so.37
sudo ln -sf /usr/lib/x86_64-linux-gnu/libjavascriptcoregtk-4.1.so.0 /usr/lib/x86_64-linux-gnu/libjavascriptcoregtk-4.0.so.18`
                  } />
                  <pre className="bg-gray-100 dark:bg-gray-800 p-3 rounded text-sm overflow-x-auto">
                    <code>
                      sudo apt update{'\n'}
                      sudo apt install -y libwebkit2gtk-4.1-0 libjavascriptcoregtk-4.1-0{'\n'}
                      sudo ln -sf /usr/lib/x86_64-linux-gnu/libwebkit2gtk-4.1.so.0 /usr/lib/x86_64-linux-gnu/libwebkit2gtk-4.0.so.37{'\n'}
                      sudo ln -sf /usr/lib/x86_64-linux-gnu/libjavascriptcoregtk-4.1.so.0 /usr/lib/x86_64-linux-gnu/libjavascriptcoregtk-4.0.so.18
                    </code>
                  </pre>
                </div>
              </li>
              <li>
                <span className="font-semibold">Fedora 35+ and newer RHEL/CentOS:</span>
                <div className="relative mt-2 mb-2">
                  <CopyButton text={
                    `sudo dnf install webkit2gtk4.1-devel
sudo ln -sf /usr/lib64/libwebkit2gtk-4.1.so.0 /usr/lib64/libwebkit2gtk-4.0.so.37
sudo ln -sf /usr/lib64/libjavascriptcoregtk-4.1.so.0 /usr/lib64/libjavascriptcoregtk-4.0.so.18`
                  } />
                  <pre className="bg-gray-100 dark:bg-gray-800 p-3 rounded text-sm overflow-x-auto">
                    <code>
                      sudo dnf install webkit2gtk4.1-devel{'\n'}
                      sudo ln -sf /usr/lib64/libwebkit2gtk-4.1.so.0 /usr/lib64/libwebkit2gtk-4.0.so.37{'\n'}
                      sudo ln -sf /usr/lib64/libjavascriptcoregtk-4.1.so.0 /usr/lib64/libjavascriptcoregtk-4.0.so.18
                    </code>
                  </pre>
                </div>
              </li>
              <li>
                <span className="font-semibold">Arch Linux:</span>
                <div className="relative mt-2 mb-2">
                  <CopyButton text={
                    `sudo pacman -S webkit2gtk-4.1
sudo ln -sf /usr/lib/libwebkit2gtk-4.1.so.0 /usr/lib/libwebkit2gtk-4.0.so.37
sudo ln -sf /usr/lib/libjavascriptcoregtk-4.1.so.0 /usr/lib/libjavascriptcoregtk-4.0.so.18`
                  } />
                  <pre className="bg-gray-100 dark:bg-gray-800 p-3 rounded text-sm overflow-x-auto">
                    <code>
                      sudo pacman -S webkit2gtk-4.1{'\n'}
                      sudo ln -sf /usr/lib/libwebkit2gtk-4.1.so.0 /usr/lib/libwebkit2gtk-4.0.so.37{'\n'}
                      sudo ln -sf /usr/lib/libjavascriptcoregtk-4.1.so.0 /usr/lib/libjavascriptcoregtk-4.0.so.18
                    </code>
                  </pre>
                </div>
              </li>
              <li>
                <span className="font-semibold">Generic Linux (if above don't work):</span>
                <div className="relative mt-2 mb-2">
                  <CopyButton text={
                    `# Ubuntu/Debian:
sudo apt install libwebkit2gtk-4.0-dev
# Fedora/RHEL/CentOS:
sudo dnf install webkit2gtk3-devel
# OpenSUSE:
sudo zypper install webkit2gtk3-devel
# Arch Linux:
sudo pacman -S webkit2gtk`
                  } />
                  <pre className="bg-gray-100 dark:bg-gray-800 p-3 rounded text-sm overflow-x-auto">
                    <code>
                      # Ubuntu/Debian:{'\n'}
                      sudo apt install libwebkit2gtk-4.0-dev{'\n'}
                      # Fedora/RHEL/CentOS:{'\n'}
                      sudo dnf install webkit2gtk3-devel{'\n'}
                      # OpenSUSE:{'\n'}
                      sudo zypper install webkit2gtk3-devel{'\n'}
                      # Arch Linux:{'\n'}
                      sudo pacman -S webkit2gtk
                    </code>
                  </pre>
                </div>
              </li>
            </ul>

            <h3 className="text-lg font-bold mt-6 mb-2">Alternative: Automated Installation Script</h3>
            <div className="relative mb-4">
              <CopyButton text={
                `curl -fsSL https://raw.githubusercontent.com/libr-forum/Libr/main/scripts/install-libr.sh | bash

# Or inspect the script first (recommended for security)
wget https://raw.githubusercontent.com/libr-forum/Libr/main/scripts/install-libr.sh
chmod +x install-libr.sh
./install-libr.sh`
              } />
              <pre className="bg-gray-100 dark:bg-gray-800 p-3 rounded text-sm overflow-x-auto">
                <code>
                  curl -fsSL https://raw.githubusercontent.com/libr-forum/Libr/main/scripts/install-libr.sh | bash{'\n\n'}
                  # Or inspect the script first (recommended for security){'\n'}
                  wget https://raw.githubusercontent.com/libr-forum/Libr/main/scripts/install-libr.sh{'\n'}
                  chmod +x install-libr.sh{'\n'}
                  ./install-libr.sh
                </code>
              </pre>
            </div>
            <p>
              The script automatically detects your distribution and handles package installation and library dependencies.
            </p>
          </div>
        );
      case 'macos':
        return (
          <div>
            <ol className="list-decimal ml-6 mb-4">
              <li>
                Download the <strong>macOS release</strong> (<code>libr-darwin-amd64.out</code>) from <a href="https://github.com/libr-forum/Libr/releases" target="_blank" rel="noopener noreferrer" className="text-blue-600 dark:text-blue-400 underline">Releases</a>
              </li>
              <li>
                On first run, macOS may block the app. To fix this:
                <ul className="list-disc ml-6">
                  <li>Go to <strong>System Settings → Privacy &amp; Security</strong></li>
                  <li>Allow the app under the "Security" section</li>
                </ul>
              </li>
              <li>
                Make executable and run:
                <div className="relative mt-2 mb-2">
                  <CopyButton text={
                    `chmod +x ./libr-darwin-amd64.out
./libr-darwin-amd64.out`
                  } />
                  <pre className="bg-gray-100 dark:bg-gray-800 p-3 rounded text-sm overflow-x-auto">
                    <code>
                      chmod +x ./libr-darwin-amd64.out{'\n'}
                      ./libr-darwin-amd64.out
                    </code>
                  </pre>
                </div>
              </li>
            </ol>
          </div>
        );
      case 'windows':
        return (
          <div>
            <ol className="list-decimal ml-6 mb-4">
              <li>
                Download the latest <strong>Windows release</strong> (<code>libr-win-amd64.exe</code>) from <a href="https://github.com/libr-forum/Libr/releases" target="_blank" rel="noopener noreferrer" className="text-blue-600 dark:text-blue-400 underline">Releases</a>
              </li>
              <li>
                Double-click to run it
                <ul className="list-disc ml-6">
                  <li>If the app doesn't start, try <strong>right-click → Run as administrator</strong></li>
                </ul>
              </li>
            </ol>
          </div>
        );
      default:
        return null;
    }
  };

  return (
    <div className="min-h-screen bg-libr-primary/50 text-foreground">
      <Header isDark={isDarkMode} toggleTheme={toggleTheme} />
      <div className="container mx-auto px-4 pl-10 py-8 mt-20">
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
            <h1 className="text-4xl font-bold">
              Install libr
            </h1>
          </div>
        </motion.div>

        {/* Content */}
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.6, delay: 0.2 }}
          className="prose prose-slate dark:prose-invert max-w-none"
        >
          <div className="flex flex-col md:flex-row w-full items-center justify-center gap-y-4 md:gap-x-8 mb-8">
            {osOptions.map(os => (
              <button
                key={os.key}
                onClick={() => setSelectedOS(os.key as 'linux' | 'macos' | 'windows')}
                className={`flex flex-col bg-card w-full items-center p-4 rounded-lg transition-all
                  ${selectedOS === os.key ? 'border-b-4 border-libr-secondary' : 'border-b-4 border-transparent'}
                  hover:border-libr-secondary border-b-2`}
                type="button"
              >
                <div className="flex flex-row items-center gap-4 text-lg">
                  {os.icon}
                  {os.label}
                </div>
              </button>
            ))}
          </div>
          {latestRelease && (
            <div className="flex flex-col md:flex-row items-center justify-between bg-libr-secondary/10 border border-libr-secondary rounded-lg p-4 mb-8">
              <div className="flex flex-col md:flex-row items-center gap-2">
                <span className="font-semibold text-libr-secondary">Latest Version:</span>
                <span className="font-mono">{latestRelease.tag}</span>
                <span className="text-xs text-muted-foreground ml-2">Updated: {latestRelease.date}</span>
              </div>
              <a
                href={getDownloadUrl()}
                className="mt-2 md:mt-0 px-4 py-2 rounded libr-button text-white font-semibold"
                target="_blank"
                rel="noopener noreferrer"
                download
              >
                Direct Download
              </a>
            </div>
          )}
          <div className="bg-card p-6 rounded-lg shadow">
            {renderInstallContent()}
          </div>
        </motion.div>
      </div>
      <Footer />
    </div>
  );
};

export default Install;
