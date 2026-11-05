import React, { useEffect, useState } from 'react';
import { motion } from 'framer-motion';
import { toast } from 'sonner';
import { useAppStore } from '../store/useAppStore';
import { apiService } from '../services/api';
import { Sidebar } from '@/components/layout/Sidebar';
import {
  Shield,
  Wrench,
  Sparkles,
  Cpu,
  Key,
  CheckCircle2,
  AlertCircle,
  Eye,
  EyeOff,
  RefreshCw,
  ExternalLink,
} from 'lucide-react';
import {
  AlertDialog,
  AlertDialogTrigger,
  AlertDialogContent,
  AlertDialogHeader,
  AlertDialogTitle,
  AlertDialogFooter,
  AlertDialogCancel,
  AlertDialogAction,
} from '@/components/ui/alert-dialog';

interface Thresholds {
  [key: string]: number;
}

const PROVIDER_PRESETS: {
  [key: string]: {
    label: string;
    icon: string;
    defaultModel: string;
    keyName: string;
    placeholder: string;
    docUrl: string;
    description: string;
  };
} = {
  groq: {
    label: 'Groq Cloud',
    icon: '⚡',
    defaultModel: 'llama-3.3-70b-versatile',
    keyName: 'GROQ_API_KEY',
    placeholder: 'gsk_...',
    docUrl: 'https://console.groq.com/keys',
    description: 'Ultra-fast Llama 3 inference directly via Groq API',
  },
  gemini: {
    label: 'Google Gemini',
    icon: '✨',
    defaultModel: 'gemini-2.0-flash',
    keyName: 'GEMINI_API_KEY',
    placeholder: 'AIzaSy...',
    docUrl: 'https://aistudio.google.com/app/apikey',
    description: 'High-speed moderation via Gemini 2.0 Flash',
  },
  openai: {
    label: 'OpenAI',
    icon: '🌐',
    defaultModel: 'gpt-4o-mini',
    keyName: 'OPENAI_API_KEY',
    placeholder: 'sk-proj-...',
    docUrl: 'https://platform.openai.com/api-keys',
    description: 'Standard OpenAI classification with gpt-4o-mini',
  },
  deepseek: {
    label: 'DeepSeek',
    icon: '🐋',
    defaultModel: 'deepseek-chat',
    keyName: 'DEEPSEEK_API_KEY',
    placeholder: 'sk-...',
    docUrl: 'https://platform.deepseek.com/api_keys',
    description: 'DeepSeek-V3 chat completions API',
  },
  custom: {
    label: 'Custom Model',
    icon: '⚙️',
    defaultModel: 'llama-3.3-70b-versatile',
    keyName: 'LLM_API_KEY',
    placeholder: 'your_api_key_here',
    docUrl: 'https://console.groq.com/keys',
    description: 'Any OpenAI-compatible direct endpoint',
  },
};

export const ModConfig: React.FC = () => {
  const { user } = useAppStore();
  const [forbiddenWords, setForbiddenWords] = useState('');
  const [thresholds, setThresholds] = useState<Thresholds>({});
  const [imageAutoMod, setImageAutoMod] = useState(true);
  const [isSaving, setIsSaving] = useState(false);

  // LLM Provider States (Direct Cloud - Zero Proxy)
  const [llmEnabled, setLlmEnabled] = useState(false);
  const [llmProvider, setLlmProvider] = useState('groq');
  const [llmModel, setLlmModel] = useState('llama-3.3-70b-versatile');
  const [providerApiKey, setProviderApiKey] = useState('');
  const [savedKeys, setSavedKeys] = useState<{ [key: string]: string }>({});
  const [showApiKey, setShowApiKey] = useState(false);
  const [isTesting, setIsTesting] = useState(false);
  const [testResult, setTestResult] = useState<{ success: boolean; message: string } | null>(null);

  // Legacy Google NLP Dialog
  const [googleApiKey, setGoogleApiKey] = useState('');
  const [isSavingKey, setIsSavingKey] = useState(false);

  useEffect(() => {
    loadConfig();
  }, []);

  const loadConfig = async () => {
    try {
      const config = await apiService.GetModConfig();
      if (config.forbidden) {
        setForbiddenWords(config.forbidden.join('\n'));
      }
      if (config.image_auto_mod !== undefined) {
        setImageAutoMod(config.image_auto_mod);
      }
      if (config.llm_enabled !== undefined) {
        setLlmEnabled(config.llm_enabled);
      }
      if (config.llm_provider) {
        setLlmProvider(config.llm_provider);
      }
      if (config.llm_model) {
        setLlmModel(config.llm_model);
      }

      if (config.thresholds) {
        const parsedThresholds: Thresholds = {};
        config.thresholds.split(',').forEach((pair) => {
          const [key, val] = pair.split(':');
          if (key && val) {
            parsedThresholds[key] = parseFloat(val);
          }
        });
        setThresholds(parsedThresholds);
      }

      const keys = await apiService.GetModKeys();
      if (keys) {
        setSavedKeys(keys);
        if (keys['GOOGLE_NLP_API_KEY']) {
          setGoogleApiKey(keys['GOOGLE_NLP_API_KEY']);
        }
      }
    } catch (error) {
      console.error('Failed to load moderation config:', error);
      toast.error('Failed to load configuration');
    }
  };

  const handleProviderSelect = (key: string) => {
    setLlmProvider(key);
    const preset = PROVIDER_PRESETS[key];
    if (preset) {
      setLlmModel(preset.defaultModel);
      setProviderApiKey('');
      setTestResult(null);
    }
  };

  const handleSliderChange = (key: string, value: number) => {
    setThresholds((prev) => ({
      ...prev,
      [key]: Math.max(0, Math.min(1, value)),
    }));
  };

  const handleInputChange = (key: string, value: string) => {
    const num = parseFloat(value);
    if (!isNaN(num)) {
      handleSliderChange(key, num);
    }
  };

  const handleTestConnection = async () => {
    setIsTesting(true);
    setTestResult(null);
    try {
      const result = await apiService.TestLLMConnection(
        llmProvider,
        llmModel,
        providerApiKey
      );
      toast.success(result);
      setTestResult({ success: true, message: result });
    } catch (error: any) {
      console.error('Test connection error:', error);
      const msg = error?.message || String(error);
      toast.error(msg);
      setTestResult({ success: false, message: msg });
    } finally {
      setIsTesting(false);
    }
  };

  const saveConfig = async () => {
    setIsSaving(true);
    try {
      const thresholdString = Object.entries(thresholds)
        .map(([key, val]) => `${key}:${val.toFixed(2)}`)
        .join(',');

      await apiService.SaveModConfig({
        forbidden: forbiddenWords
          .split('\n')
          .map((w) => w.trim())
          .filter(Boolean),
        thresholds: thresholdString,
        image_auto_mod: imageAutoMod,
        llm_enabled: llmEnabled,
        llm_provider: llmProvider,
        llm_model: llmModel,
      });

      // If an API key was entered, persist it to modkeys.json
      if (providerApiKey.trim()) {
        const preset = PROVIDER_PRESETS[llmProvider];
        const keyName = preset ? preset.keyName : 'LLM_API_KEY';
        await apiService.SaveModKey(keyName, providerApiKey.trim());
        toast.success(`Saved API key for ${preset?.label || llmProvider}`);
        setProviderApiKey('');
        // Reload keys
        const updatedKeys = await apiService.GetModKeys();
        setSavedKeys(updatedKeys || {});
      }

      toast.success('Moderation settings saved successfully');
    } catch (error) {
      console.error('Failed to save config:', error);
      toast.error('Failed to save configuration');
    } finally {
      setIsSaving(false);
    }
  };

  const saveLegacyGoogleApiKey = async () => {
    setIsSavingKey(true);
    try {
      await apiService.SaveGoogleApiKey(googleApiKey);
      toast.success('Google API key saved');
      const updatedKeys = await apiService.GetModKeys();
      setSavedKeys(updatedKeys || {});
    } catch (error) {
      console.error('Failed to save Google API key:', error);
      toast.error('Failed to save Google API key');
    } finally {
      setIsSavingKey(false);
    }
  };

  if (user?.role !== 'moderator') {
    return (
      <div className="flex flex-row">
        <div className="w-[19.4%]">
          <Sidebar />
        </div>
        <div className="flex-1 flex items-center justify-center bg-libr-primary m-80">
          <motion.div
            initial={{ opacity: 0, scale: 0.9 }}
            animate={{ opacity: 1, scale: 1 }}
            className="text-center"
          >
            <div className="w-20 h-20 bg-red-500/20 rounded-2xl flex items-center justify-center mx-auto mb-4">
              <Shield className="w-10 h-10 text-red-500" />
            </div>
            <h2 className="text-xl font-semibold text-foreground mb-2">Access Denied</h2>
            <p className="text-muted-foreground">
              You need moderator privileges to access this page
            </p>
          </motion.div>
        </div>
      </div>
    );
  }

  const activePreset = PROVIDER_PRESETS[llmProvider] || PROVIDER_PRESETS.custom;
  const currentSavedKey = savedKeys[activePreset.keyName];

  return (
    <div className="flex flex-row">
      <div className="w-[19.4%]">
        <Sidebar />
      </div>
      <div className="flex-1 flex flex-col h-screen">
        <div className="flex-1 overflow-y-auto">
          <div className="pt-6 pb-24">
            <motion.div
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              className="max-w-6xl mx-auto px-6"
            >
              {/* Top Header */}
              <div className="mb-8 flex items-center justify-between border-b border-border/40 pb-6">
                <div className="flex items-center space-x-3">
                  <div className="w-12 h-12 bg-libr-accent1/20 rounded-xl flex items-center justify-center">
                    <Wrench className="w-6 h-6 text-libr-accent2" />
                  </div>
                  <div>
                    <h1 className="text-2xl font-bold text-foreground">Moderation Configuration</h1>
                    <p className="text-muted-foreground text-sm">
                      Configure automated LLM classification, direct cloud provider keys, and safety thresholds
                    </p>
                  </div>
                </div>

                <div className="flex items-center space-x-3">
                  <AlertDialog>
                    <AlertDialogTrigger asChild>
                      <button className="px-3.5 py-2 text-xs font-medium rounded-lg bg-muted/40 text-muted-foreground hover:text-foreground hover:bg-muted/60 border border-border/50 transition">
                        Legacy Google NLP Key
                      </button>
                    </AlertDialogTrigger>
                    <AlertDialogContent>
                      <AlertDialogHeader>
                        <AlertDialogTitle>Enter Google Cloud NLP Key</AlertDialogTitle>
                      </AlertDialogHeader>
                      <input
                        placeholder="Google Cloud API Key"
                        value={googleApiKey}
                        onChange={(e) => setGoogleApiKey(e.target.value)}
                        className="w-full mt-4 px-3 py-2 text-sm rounded-md bg-muted/30 border border-border/50 text-foreground"
                      />
                      <AlertDialogFooter className="mt-4">
                        <AlertDialogCancel>Cancel</AlertDialogCancel>
                        <AlertDialogAction
                          className="bg-libr-accent1/20 libr-button hover:bg-muted/30 text-foreground"
                          onClick={saveLegacyGoogleApiKey}
                          disabled={isSavingKey}
                        >
                          {isSavingKey ? 'Saving...' : 'Save'}
                        </AlertDialogAction>
                      </AlertDialogFooter>
                    </AlertDialogContent>
                  </AlertDialog>

                  <button
                    onClick={saveConfig}
                    disabled={isSaving}
                    className="flex items-center space-x-2 px-5 py-2 text-sm font-semibold rounded-lg bg-libr-accent1 text-black hover:opacity-90 transition disabled:opacity-50"
                  >
                    <span>{isSaving ? 'Saving...' : 'Save All Settings'}</span>
                  </button>
                </div>
              </div>

              {/* ── SECTION 1: Direct Cloud LLM Moderation (Zero Proxy) ───────── */}
              <div className="mb-8 p-6 rounded-2xl border border-libr-accent1/30 bg-libr-accent1/5 shadow-sm">
                <div className="flex items-center justify-between mb-6">
                  <div className="flex items-center space-x-3">
                    <div className="w-10 h-10 rounded-xl bg-libr-accent1/20 flex items-center justify-center text-libr-accent2">
                      <Sparkles className="w-5 h-5 text-libr-accent2" />
                    </div>
                    <div>
                      <h2 className="text-lg font-bold text-foreground flex items-center gap-2">
                        Direct Cloud LLM Moderation
                        <span
                          className={`text-xs px-2 py-0.5 rounded-full font-medium ${
                            llmEnabled
                              ? 'bg-green-500/20 text-green-400 border border-green-500/30'
                              : 'bg-muted text-muted-foreground'
                          }`}
                        >
                          {llmEnabled ? 'Active' : 'Disabled'}
                        </span>
                      </h2>
                      <p className="text-xs text-muted-foreground">
                        Native zero-proxy integration: call Groq, Gemini, OpenAI, or DeepSeek directly via API key with no background servers.
                      </p>
                    </div>
                  </div>

                  {/* LLM Master Toggle */}
                  <div className="flex items-center space-x-3">
                    <span className="text-xs font-medium text-muted-foreground">Enable LLM Moderation</span>
                    <button
                      role="switch"
                      aria-checked={llmEnabled}
                      onClick={() => setLlmEnabled((prev) => !prev)}
                      className={`relative inline-flex h-6 w-11 items-center rounded-full transition-colors focus:outline-none ${
                        llmEnabled ? 'bg-libr-accent1' : 'bg-muted'
                      }`}
                    >
                      <span
                        className={`inline-block h-4 w-4 transform rounded-full bg-white transition-transform ${
                          llmEnabled ? 'translate-x-6' : 'translate-x-1'
                        }`}
                      />
                    </button>
                  </div>
                </div>

                {/* Provider Selector Cards */}
                <div className="mb-6">
                  <label className="block text-xs font-semibold text-muted-foreground uppercase tracking-wider mb-3">
                    Select Cloud Provider
                  </label>
                  <div className="grid grid-cols-2 sm:grid-cols-5 gap-3">
                    {Object.entries(PROVIDER_PRESETS).map(([key, preset]) => {
                      const isSelected = llmProvider === key;
                      const hasKey = Boolean(savedKeys[preset.keyName]);
                      return (
                        <button
                          key={key}
                          type="button"
                          onClick={() => handleProviderSelect(key)}
                          className={`p-3.5 rounded-xl border text-left transition relative flex flex-col justify-between ${
                            isSelected
                              ? 'border-libr-accent1 bg-libr-accent1/20 text-foreground shadow-sm'
                              : 'border-border/60 bg-muted/20 text-muted-foreground hover:bg-muted/40 hover:text-foreground'
                          }`}
                        >
                          <div className="flex items-center justify-between mb-2">
                            <span className="text-xl">{preset.icon}</span>
                            {hasKey && (
                              <span className="w-2 h-2 rounded-full bg-green-500" title="API Key Configured" />
                            )}
                          </div>
                          <div>
                            <p className="text-sm font-semibold leading-tight">{preset.label}</p>
                            <p className="text-[11px] opacity-70 truncate mt-0.5">{preset.defaultModel}</p>
                          </div>
                        </button>
                      );
                    })}
                  </div>
                </div>

                {/* Model Configuration Form */}
                <div className="mb-5">
                  <div className="flex justify-between items-center mb-1.5">
                    <label className="text-xs font-medium text-foreground flex items-center gap-1.5">
                      <Cpu className="w-3.5 h-3.5 text-libr-accent2" />
                      Model Identifier
                    </label>
                    <span className="text-[11px] text-muted-foreground">Default: {activePreset.defaultModel}</span>
                  </div>
                  <input
                    type="text"
                    value={llmModel}
                    onChange={(e) => setLlmModel(e.target.value)}
                    placeholder={activePreset.defaultModel}
                    className="w-full px-3.5 py-2.5 bg-muted/40 border border-border/60 rounded-xl text-sm text-foreground focus:outline-none focus:border-libr-accent1 font-mono"
                  />
                  <p className="text-[11px] text-muted-foreground mt-1">{activePreset.description}</p>
                </div>

                {/* API Key Input */}
                <div className="mb-5">
                  <div className="flex justify-between items-center mb-1.5">
                    <label className="text-xs font-medium text-foreground flex items-center gap-1.5">
                      <Key className="w-3.5 h-3.5 text-libr-accent2" />
                      {activePreset.label} API Key ({activePreset.keyName})
                    </label>
                    <div className="flex items-center space-x-2">
                      {currentSavedKey && (
                        <span className="text-[11px] text-green-400 font-mono">
                          Configured: {currentSavedKey}
                        </span>
                      )}
                      <a
                        href={activePreset.docUrl}
                        target="_blank"
                        rel="noreferrer"
                        className="text-[11px] text-libr-accent2 hover:underline flex items-center gap-1"
                      >
                        Get Key <ExternalLink className="w-2.5 h-2.5" />
                      </a>
                    </div>
                  </div>

                  <div className="relative flex items-center">
                    <input
                      type={showApiKey ? 'text' : 'password'}
                      value={providerApiKey}
                      onChange={(e) => setProviderApiKey(e.target.value)}
                      placeholder={currentSavedKey ? 'Leave blank to keep existing key, or paste new key' : activePreset.placeholder}
                      className="w-full pl-3.5 pr-20 py-2.5 bg-muted/40 border border-border/60 rounded-xl text-sm text-foreground focus:outline-none focus:border-libr-accent1 font-mono"
                    />
                    <div className="absolute right-2 flex items-center space-x-1">
                      <button
                        type="button"
                        onClick={() => setShowApiKey((prev) => !prev)}
                        className="p-1.5 text-muted-foreground hover:text-foreground rounded-lg"
                        title={showApiKey ? 'Hide' : 'Show'}
                      >
                        {showApiKey ? <EyeOff className="w-4 h-4" /> : <Eye className="w-4 h-4" />}
                      </button>
                    </div>
                  </div>
                </div>

                {/* Test Connection Action */}
                <div className="flex items-center justify-between pt-3 border-t border-border/40">
                  <div className="flex-1 mr-4">
                    {testResult && (
                      <div
                        className={`text-xs flex items-center gap-1.5 px-3 py-1.5 rounded-lg ${
                          testResult.success
                            ? 'bg-green-500/10 text-green-400 border border-green-500/20'
                            : 'bg-red-500/10 text-red-400 border border-red-500/20'
                        }`}
                      >
                        {testResult.success ? (
                          <CheckCircle2 className="w-3.5 h-3.5 flex-shrink-0" />
                        ) : (
                          <AlertCircle className="w-3.5 h-3.5 flex-shrink-0" />
                        )}
                        <span className="truncate">{testResult.message}</span>
                      </div>
                    )}
                  </div>

                  <button
                    type="button"
                    onClick={handleTestConnection}
                    disabled={isTesting}
                    className="px-4 py-2 text-xs font-semibold rounded-lg bg-muted/60 hover:bg-muted text-foreground border border-border/60 transition flex items-center gap-2 disabled:opacity-50"
                  >
                    <RefreshCw className={`w-3.5 h-3.5 ${isTesting ? 'animate-spin' : ''}`} />
                    {isTesting ? 'Testing Cloud Connection...' : 'Test Direct Connection'}
                  </button>
                </div>
              </div>

              {/* ── SECTION 2: Image Auto-Moderation ────────────────────────── */}
              <div className="mb-8 flex items-center justify-between p-5 border border-border/50 rounded-2xl bg-muted/20">
                <div>
                  <p className="text-sm font-semibold text-foreground">Image Auto-Moderation</p>
                  <p className="text-xs text-muted-foreground mt-1">
                    When enabled, images are automatically analyzed for adult, violent, and racy content. When disabled, all image messages are sent to the manual review queue.
                  </p>
                </div>
                <button
                  role="switch"
                  aria-checked={imageAutoMod}
                  onClick={() => setImageAutoMod((prev) => !prev)}
                  className={`relative inline-flex h-6 w-11 items-center rounded-full transition-colors focus:outline-none ${
                    imageAutoMod ? 'bg-libr-accent1' : 'bg-muted'
                  }`}
                >
                  <span
                    className={`inline-block h-4 w-4 transform rounded-full bg-white transition-transform ${
                      imageAutoMod ? 'translate-x-6' : 'translate-x-1'
                    }`}
                  />
                </button>
              </div>

              {/* ── SECTION 3: Forbidden Words ──────────────────────────────── */}
              <div className="mb-8 p-5 border border-border/50 rounded-2xl bg-muted/20">
                <label className="block text-sm font-semibold text-foreground mb-1">
                  Forbidden Words / Phrases
                </label>
                <p className="text-xs text-muted-foreground mb-3">
                  Enter words or phrases (one per line). Any message containing these will be automatically rejected.
                </p>
                <textarea
                  rows={5}
                  className="w-full p-3 bg-muted/40 border border-border/60 rounded-xl text-sm text-foreground focus:outline-none focus:border-libr-accent1 font-mono"
                  value={forbiddenWords}
                  onChange={(e) => setForbiddenWords(e.target.value)}
                  placeholder="scam&#10;phishing&#10;free crypto"
                />
              </div>

              {/* ── SECTION 4: Threshold Controls ────────────────────────────── */}
              <div className="mb-8 p-5 border border-border/50 rounded-2xl bg-muted/20">
                <div className="mb-4">
                  <h3 className="text-sm font-semibold text-foreground">Category Sensitivity Thresholds</h3>
                  <p className="text-xs text-muted-foreground">
                    Content confidence scores equal to or above these thresholds will be automatically rejected.
                  </p>
                </div>

                <div className="grid grid-cols-1 sm:grid-cols-2 gap-8">
                  {Object.keys(thresholds).map((key) => (
                    <div key={key} className="space-y-2 p-3 rounded-xl bg-background/50 border border-border/40">
                      <div className="flex justify-between items-center">
                        <label className="text-xs font-semibold text-foreground">{key}</label>
                        <span className="text-xs font-mono text-libr-accent2">
                          {(thresholds[key] ?? 0).toFixed(2)}
                        </span>
                      </div>
                      <input
                        type="range"
                        min={0}
                        max={1}
                        step={0.01}
                        value={thresholds[key] ?? 0.5}
                        onChange={(e) => handleSliderChange(key, parseFloat(e.target.value))}
                        className="w-full h-2 accent-libr-accent1 appearance-none bg-libr-accent1/20 rounded-full cursor-pointer"
                      />
                      <input
                        type="number"
                        step={0.01}
                        min={0}
                        max={1}
                        value={thresholds[key] ?? 0.5}
                        onChange={(e) => handleInputChange(key, e.target.value)}
                        className="w-full px-2 py-1 border border-border/50 rounded-lg bg-background text-xs font-mono text-foreground focus:outline-none focus:border-libr-accent1"
                      />
                    </div>
                  ))}
                </div>
              </div>

              {/* Bottom Sticky-like Save Bar */}
              <div className="flex justify-end pt-4 border-t border-border/40">
                <motion.button
                  onClick={saveConfig}
                  whileHover={{ scale: 1.02 }}
                  whileTap={{ scale: 0.98 }}
                  className="flex items-center space-x-2 px-6 py-3 rounded-xl bg-libr-accent1 text-black font-semibold hover:opacity-95 transition disabled:opacity-50"
                  disabled={isSaving}
                >
                  <Wrench className="w-4 h-4 text-black" />
                  <span>{isSaving ? 'Saving Changes...' : 'Save Moderation Configuration'}</span>
                </motion.button>
              </div>
            </motion.div>
          </div>
        </div>
      </div>
    </div>
  );
};
