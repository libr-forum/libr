import React, { useEffect, useState } from 'react';
import { motion } from 'framer-motion';
import { useAppStore } from '../store/useAppStore';
import { apiService } from '../services/api';
import {
  Shield,
  Wrench,
} from 'lucide-react';

interface Thresholds {
  [key: string]: number;
}

export const ModConfig: React.FC = () => {
  const { user } = useAppStore();
  const [forbiddenWords, setForbiddenWords] = useState('');
  const [thresholds, setThresholds] = useState<Thresholds>({});
  const [isSaving, setIsSaving] = useState(false);

  useEffect(() => {
    loadConfig();
  }, []);

  const loadConfig = async () => {
    const config = await apiService.GetModConfig();
    setForbiddenWords(config.forbidden.join('\n'));

    const parsedThresholds: Thresholds = {};
    config.thresholds.split(',').forEach(pair => {
      const [key, val] = pair.split(':');
      parsedThresholds[key] = parseFloat(val);
    });
    setThresholds(parsedThresholds);
  };

  const handleSliderChange = (key: string, value: number) => {
    setThresholds(prev => ({
      ...prev,
      [key]: Math.max(0, Math.min(1, value))
    }));
  };

  const handleInputChange = (key: string, value: string) => {
    const num = parseFloat(value);
    if (!isNaN(num)) {
      handleSliderChange(key, num);
    }
  };

  const saveConfig = async () => {
    setIsSaving(true);

    const thresholdString = Object.entries(thresholds)
      .map(([key, val]) => `${key}:${val.toFixed(2)}`)
      .join(',');

    await apiService.SaveModConfig({
      forbidden: forbiddenWords.split('\n').map(w => w.trim()).filter(Boolean),
      thresholds: thresholdString
    });

    setIsSaving(false);
  };

  if (user?.role !== 'moderator') {
    return (
      <div className="flex-1 flex items-center justify-center bg-libr-primary">
        <motion.div
          initial={{ opacity: 0, scale: 0.9 }}
          animate={{ opacity: 1, scale: 1 }}
          className="text-center"
        >
          <div className="w-20 h-20 bg-red-500/20 rounded-2xl flex items-center justify-center mx-auto mb-4">
            <Shield className="w-10 h-10 text-red-500" />
          </div>
          <h2 className="text-xl font-semibold text-foreground mb-2">
            Access Denied
          </h2>
          <p className="text-muted-foreground">
            You need moderator privileges to access this page
          </p>
        </motion.div>
      </div>
    );
  }

  return (
    <div className="flex-1 flex flex-col bg-libr-primary h-screen">
      <div className="flex-1 overflow-y-auto">
        <div className="p-6 pb-24">
          <motion.div
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            className="max-w-6xl mx-auto"
          >
            <div className="mb-8">
              <div className="flex items-center space-x-3 mb-4">
                <div className="w-12 h-12 bg-libr-accent1/20 rounded-xl flex items-center justify-center">
                  <Wrench className="w-6 h-6 text-libr-accent2" />
                </div>
                <div>
                  <h1 className="text-2xl font-bold text-foreground">Moderation Config</h1>
                  <p className="text-muted-foreground">
                    Configure thresholds and forbidden words for moderation
                  </p>
                </div>
              </div>

              {/* Forbidden Words */}
              <div className="mb-8">
                <label className="block text-sm font-medium text-muted-foreground mb-2">
                  Forbidden Words (one per line)
                </label>
                <textarea
                  rows={6}
                  className="w-full p-3 bg-muted/30 border border-border/50 rounded-lg text-sm text-foreground"
                  value={forbiddenWords}
                  onChange={e => setForbiddenWords(e.target.value)}
                />
              </div>

              {/* Threshold Controls */}
              <div className="grid grid-cols-1 sm:grid-cols-2 gap-6">
                {Object.keys(thresholds).map((key) => (
                  <div key={key} className="space-y-2">
                    <label className="text-sm font-medium text-foreground">{key}</label>
                    <input
                      type="range"
                      min={0}
                      max={1}
                      step={0.01}
                      value={thresholds[key]}
                      onChange={e => handleSliderChange(key, parseFloat(e.target.value))}
                      className="w-full h-2 accent-libr-accent1 appearance-none bg-libr-accent1/20 rounded-full"
                    />
                    <input
                      type="number"
                      step={0.01}
                      min={0}
                      max={1}
                      value={thresholds[key]}
                      onChange={e => handleInputChange(key, e.target.value)}
                      className="w-full px-2 py-1 border border-border/50 rounded-lg bg-background text-sm text-foreground"
                    />
                  </div>
                ))}
              </div>

              {/* Save Button */}
              <div className="mt-8 flex justify-end">
                <motion.button
                  onClick={async () => {
                    await saveConfig();
                    // Wait a tiny bit to let backend update
                    setTimeout(() => {
                        loadConfig();
                    }, 300); // 300ms delay
                  }}

                  whileHover={{ scale: 1.03 }}
                  whileTap={{ scale: 0.97 }}
                  className="flex items-center space-x-2 px-5 py-2.5 rounded-lg bg-libr-accent1 text-white shadow-md hover:bg-libr-accent1/80 transition disabled:opacity-50 disabled:pointer-events-none"
                  disabled={isSaving}
                >
                  <span>{isSaving ? 'Saving...' : 'Save'}</span>
                </motion.button>
              </div>
            </div>
          </motion.div>
        </div>
      </div>
    </div>
  );
};
