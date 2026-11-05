import React, { useState, useEffect, useCallback, useRef } from 'react';
import { StyleSheet, TextInput, TouchableOpacity, ActivityIndicator, FlatList, View, Text, Alert, RefreshControl, KeyboardAvoidingView, Platform, StatusBar, Modal, ScrollView, Animated } from 'react-native';
import { SafeAreaView, useSafeAreaInsets } from 'react-native-safe-area-context';
import { Image } from 'expo-image';
import { SvgXml } from 'react-native-svg';
import { useRouter } from 'expo-router';
import LibrCore, { LibrCoreEvents, RetMsgCert, MsgCert } from '@/modules/LibrCore';
import { useAppStore } from '@/store/useAppStore';
import { useSidebar } from './_layout';
import { Menu, Plus, Check, Clock, Shield, Trash2, Flag, X, AlertTriangle, Hash, RotateCcw } from 'lucide-react-native';
import { ConnectingScreen } from '@/components/ConnectingScreen';
import { AnimatedHamburger } from '@/components/AnimatedHamburger';
import { Colors, Fonts, getAppColors } from '@/constants/theme';

const LIBR_LOGO_SVG = `<svg id="Layer_1" data-name="Layer 1" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 1000 1000"><path d="M572.9,63.28c13-7.31,26.55-12.91,41.24-15.66,5.72-1.13,11.52-2.12,17.33-1.8,7.17.4,11.45,4.14,11.24,11.85a13.2,13.2,0,0,1-.21,3.89c-8.42,24.79-9.91,51.6-22.67,75.2-5.44,10.05-12.23,20.33-8.73,33,.83,7.94-5.21,10.86-10.49,14.35a60.61,60.61,0,0,0-21.82,22.27c-17.65,21.75-35.95,42.93-51.74,66.18C496.86,317,471.54,363.9,452.78,414.3c-3.16,8.49-5.46,17.44-11.18,24.82-4.2,2.85-5,7.58-6,11.9-2.42,10.56-8.56,18.89-15.21,26.94a27.59,27.59,0,0,1-17.21,10.19c-11.87,2-20.7,10.68-31.08,16-10.9,5.54-19.61,5.19-28.46-2.65-13-19.2-20.8-40.79-24.94-63.15-12.6-67.94-1.47-132,34.78-191.26,4-6.48,4.95-14.67,11.85-19.74-6.69,19.05-8.79,38.75-7.55,58.52,2.13,34.18,6.59,67.93,27.13,97.1,2.07,2.94,3.9,7.4,7.83,5.71,4.17-1.79,1.43-6.11.86-9.48-.88-5.09-3.43-9.91-2.62-15.29,6.35.16,12,2.9,18.89,2.55,12.88-.67,23.26-4.36,32-13.8,3.29-3.54,8.11-5.45,12.53-7.61A47.77,47.77,0,0,0,477,323.13c6.56-12.56,17.61-22.4,25.2-34.75,10.73-17.43,22.34-34.26,35.9-49.93,9.39-10.85,17.27-23.52,13.91-40.93-3.5-18.14-.15-36.24,11.6-52.18,8-10.81,12.63-23.58,15.48-36.81.95-4.4,3.71-7.53,6.42-10.82,5.83-7.07,11.65-14,13.38-23.59s1-10.73-9.19-9.47C584,65.36,578.15,68,572.9,63.28Z" style="fill:#00ffd0"/><path d="M528.62,485.8c-17.64,9.58-37.15,13.8-56.1,19.69-3.24,1-8.28.06-9,5.54,4.75,2.83,9.47-.11,14.2-.15,8.77-.08,12.56,4.24,10.56,12.83a24.27,24.27,0,0,1-3.1,8.17c-13.84,20.5-22.32,43.89-35.15,64.92-2.21,3.63-3.77,7.55-5.69,11.3-2,4-4.81,6.27-9.23,2.8-1.49-2.6-1.4-5.37-.92-8.19,2.16-11.4.71-22.91.94-34.37A305.89,305.89,0,0,1,443.46,499c6.75-26.25,12.8-52.66,21.44-78.45,15.75-47,35.57-92,60.73-134.72,4-6.76,7.68-13.69,11.81-20.35,12.45-18.65,24.82-37.37,40.34-53.72,9.91-13.7,24-23.55,34-37.22,2.1-2.68,3.81-5.75,7.06-7.34,5.44-2.49,10.93-8.68,16.93-3.9s1.9,11.7.29,17.44c-5.45,19.49-13.15,38.3-17.09,58.29-2.1,10.63-9.94,16.46-20.37,18.82-2.43.55-5.26-.09-7.13,2.25-3.91,1.26-8.11,2-11.62,4-1.94,1.1.93,1.13,1.52,1.72,11.8,4.57,15.46,10.85,13.41,23.32-1.11,6.74-2.07,13.72-5.77,19.53-9.51,14.91-13.13,31.74-16.35,48.65-2.11,11.1-7.19,20.71-12.84,30.19l-54.37,26.81c.23.66.47,1.32.71,2l20.37-5.05c3.41,1.72,6.24-.75,9.33-1.32,11.3-2.08,18.62,4.46,15.7,15.43-3.94,14.78-8.76,29.31-10.81,44.53C539.69,477.49,531.15,479.34,528.62,485.8Z" style="fill:#00ffd0"/></svg>`;

// ── Design tokens ────────────────────────────────────────────────────────────
const C = {
  bg: Colors.dark.background,
  card: Colors.dark.primary,
  border: Colors.dark.border,
  teal: Colors.dark.tint,
  tealDim: 'rgba(31,164,169,0.15)',
  cyan: Colors.dark.accent,
  text: Colors.dark.text,
  muted: Colors.dark.muted,
  green: Colors.dark.green,
  red: Colors.dark.red,
  amber: Colors.dark.amber,
  debugBg: Colors.dark.debugBg,
} as const;

const ALIAS_COLORS = ['#1fa4a9', '#9b6fd4', '#60b3f0', '#00fcdf', '#f0a060', '#e87c7c'];
function aliasColor(pubkey: string): string {
  let n = 0;
  for (let i = 0; i < pubkey.length; i++) n = (n * 31 + pubkey.charCodeAt(i)) & 0xffff;
  return ALIAS_COLORS[n % ALIAS_COLORS.length];
}

function getAvatarUri(svg: string | null): string | null {
  if (!svg) return null;
  if (svg.startsWith('<svg')) {
    return `data:image/svg+xml;utf8,${encodeURIComponent(svg)}`;
  }
  // Assume base64
  if (!svg.startsWith('data:')) {
    return `data:image/svg+xml;base64,${svg}`;
  }
  return svg;
}

// ── MessageCard ──────────────────────────────────────────────────────────────



// ── MessageImage & Gallery ──────────────────────────────────────────────────

function MessageImage({ src }: { src: string }) {
  const [modalVisible, setModalVisible] = useState(false);

  return (
    <>
      <TouchableOpacity onPress={() => setModalVisible(true)} activeOpacity={0.9}>
        <Image
          source={src}
          style={styles.messageImage}
          contentFit="cover"
          transition={200}
        />
      </TouchableOpacity>

      <Modal visible={modalVisible} transparent animationType="fade" onRequestClose={() => setModalVisible(false)}>
        <TouchableOpacity style={styles.lightboxBackdrop} activeOpacity={1} onPress={() => setModalVisible(false)}>
          <Image
            source={src}
            style={styles.lightboxImage}
            contentFit="contain"
          />
          <TouchableOpacity style={styles.lightboxClose} onPress={() => setModalVisible(false)}>
            <X size={28} color="#fff" />
          </TouchableOpacity>
        </TouchableOpacity>
      </Modal>
    </>
  );
}

function MessageCard({ cert, myPublicKey, onLongPress, isIncognito }: { cert: RetMsgCert, myPublicKey: string, onLongPress: (c: RetMsgCert) => void, isIncognito: boolean }) {
  const isMine = cert.public_key === myPublicKey;
  const cardColors = getAppColors(isIncognito);
  const [alias, setAlias] = useState<string>('…');
  const [avatarSvg, setAvatarSvg] = useState<string | null>(null);

  useEffect(() => {
    let cancelled = false;
    (async () => {
      try {
        const a = await LibrCore.generateAlias(cert.public_key);
        if (!cancelled) setAlias(a);
        const av = await LibrCore.generateAvatar(cert.public_key);
        if (!cancelled) setAvatarSvg(av);
      } catch { /* ignore */ }
    })();
    return () => { cancelled = true; };
  }, [cert.public_key]);

  const ts = new Date(cert.msg.ts * 1000).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });

  let rawContent = cert.msg.content;
  const matchHead = rawContent.match(/<HEAD>(.*?)<\/HEAD>/s);
  const matchBody = rawContent.match(/<BODY>(.*?)<\/BODY>/s);
  const title = matchHead ? matchHead[1].trim() : null;
  rawContent = matchBody ? matchBody[1].trim() : rawContent;

  // Extract images
  const imgSrcs: string[] = [];
  const imgRegex = /<img[^>]+src="([^">]+)"/g;
  let match;
  while ((match = imgRegex.exec(rawContent)) !== null) {
    imgSrcs.push(match[1]);
  }

  // Final text: remove all HTML tags
  const plainContent = rawContent
    .replace(/<\/?[^>]+(>|$)/g, '')
    .replace(/&amp;/g, '&')
    .replace(/&lt;/g, '<')
    .replace(/&gt;/g, '>')
    .replace(/&quot;/g, '"')
    .replace(/&#39;/g, "'")
    .trim();

  const color = aliasColor(cert.public_key);

  return (
    <TouchableOpacity
      onLongPress={() => onLongPress(cert)}
      activeOpacity={0.85}
    >
      <View style={[
        styles.card,
        cert.sign.startsWith('temp-') && { opacity: 0.7 },
        { overflow: 'hidden', backgroundColor: cardColors.primary, borderColor: cardColors.border }
      ]}>
        {isMine && <View style={{
          position: 'absolute',
          left: 0,
          top: 0,
          bottom: 0,
          width: 4,
          backgroundColor: cardColors.tint,
          borderTopLeftRadius: 16,
          borderBottomLeftRadius: 16
        }} />}
        <View style={styles.cardAvatar}>
          {avatarSvg ? (
            <Image source={getAvatarUri(avatarSvg)} style={styles.avatarImage} contentFit="cover" />
          ) : (
            <View style={[styles.avatarFallback, { backgroundColor: color + '33' }]}>
              <Text style={[styles.avatarInitial, { color }]}>{alias.charAt(0).toUpperCase()}</Text>
            </View>
          )}
        </View>
        <View style={styles.cardBody}>
          <View style={styles.cardMeta}>
            <Text style={styles.aliasText} numberOfLines={1}>{alias}</Text>
          </View>
          {title && <Text style={styles.titleText}>{title}</Text>}
          {plainContent.length > 0 && <Text style={styles.contentText}>{plainContent}</Text>}

          {imgSrcs.length > 0 && (
            <View style={styles.imageGallery}>
              {imgSrcs.map((src, i) => (
                <MessageImage key={i} src={src} />
              ))}
            </View>
          )}

          <Text style={styles.timeText}>{ts}</Text>
        </View>
      </View>
    </TouchableOpacity>
  );
}

// ── ModSignatureItem ──────────────────────────────────────────────────────────

function ModSignatureItem({ modPub }: { modPub: string }) {
  const [alias, setAlias] = useState('…');
  useEffect(() => {
    LibrCore.generateAlias(modPub).then(setAlias).catch(() => { });
  }, [modPub]);

  return (
    <View style={styles.modItem}>
      <Text style={styles.modAlias} numberOfLines={1}>{alias}</Text>
      <Check size={14} color={C.green} />
      <Text style={styles.modSignLabel}>sign</Text>
    </View>
  );
}

// ── MessageDetailModal ────────────────────────────────────────────────────────

interface MessageDetailModalProps {
  visible: boolean;
  onClose: () => void;
  cert: RetMsgCert | null;
  isMine: boolean;
  onDelete: (c: RetMsgCert) => void;
  onReport: (c: RetMsgCert) => void;
  onRetry: (c: RetMsgCert) => void;
}

function MessageDetailModal({ visible, onClose, cert, isMine, onDelete, onReport, onRetry }: MessageDetailModalProps) {
  const { state } = useAppStore();
  if (!cert) return null;
  const fullTs = new Date(cert.msg.ts * 1000).toLocaleString([], { dateStyle: 'medium', timeStyle: 'medium' });
  const isReported = state.reportedSigns.has(cert.sign);
  const reportingDisabled = isReported || state.isIncognito;

  const isPending = cert.sign.startsWith('temp-');
  const isTimeout = (Date.now() / 1000 - cert.msg.ts) > 15;
  const hasAttachments = /<img/i.test(cert.msg.content);
  const canRetry = isMine && isPending && !hasAttachments && isTimeout;

  return (
    <Modal visible={visible} transparent animationType="fade" onRequestClose={onClose}>
      <TouchableOpacity style={styles.modalBackdrop} activeOpacity={1} onPress={onClose}>
        <View style={styles.modalContent} onStartShouldSetResponder={() => true}>
          <View style={styles.modalHeader}>
            <Clock size={16} color={C.muted} />
            <Text style={styles.modalTimeLabel}>Time</Text>
            <Text style={styles.modalTimeValue}>{fullTs}</Text>
          </View>

          <View style={styles.statusRow}>
            {isReported ? (
              <>
                <Flag size={16} color={C.red} />
                <Text style={[styles.statusLabel, { color: C.red }]}>Reported</Text>
              </>
            ) : (
              <>
                <Check size={16} color={C.green} />
                <Text style={styles.statusLabel}>Approved</Text>
              </>
            )}
          </View>

          <ScrollView style={styles.modList} bounces={false}>
            {cert.mod_certs?.map((mc, i) => (
              <ModSignatureItem key={i} modPub={mc.public_key} />
            ))}
          </ScrollView>

          <View style={styles.modalActions}>
            {isMine && (
              <>
                {canRetry && (
                  <TouchableOpacity 
                    style={styles.actionBtn} 
                    onPress={() => { onClose(); onRetry(cert); }}
                  >
                    <RotateCcw size={18} color={C.teal} />
                    <Text style={[styles.actionBtnText, { color: C.teal }]}>Retry Send</Text>
                  </TouchableOpacity>
                )}
                <TouchableOpacity style={styles.actionBtn} onPress={() => { onClose(); onDelete(cert); }}>
                  <Trash2 size={18} color={C.red} />
                  <Text style={[styles.actionBtnText, { color: C.red }]}>Delete</Text>
                </TouchableOpacity>
              </>
            )}
            {!isMine && (
              <TouchableOpacity
                style={[styles.actionBtn, reportingDisabled && { opacity: 0.5 }]}
                disabled={reportingDisabled}
                onPress={() => { onClose(); onReport(cert); }}
              >
                <Flag size={18} color={C.red} />
                <Text style={[styles.actionBtnText, { color: C.red }]}>
                  {isReported ? 'Reported' : state.isIncognito ? 'Blocked In Incognito' : 'Report'}
                </Text>
              </TouchableOpacity>
            )}
          </View>
        </View>
      </TouchableOpacity>
    </Modal>
  );
}

// ── ReportReasonModal ────────────────────────────────────────────────────────

function ReportReasonModal({
  visible,
  onClose,
  onSubmit
}: {
  visible: boolean;
  onClose: () => void;
  onSubmit: (reason: string) => void
}) {
  const [reason, setReason] = useState('');
  const [selectedChip, setSelectedChip] = useState<string | null>(null);

  const chips = ['Spam', 'Inappropriate', 'Harassment', 'False Identity'];

  const handlePressChip = (chip: string) => {
    setSelectedChip(chip);
    setReason(chip);
  };

  const handleCustomChange = (text: string) => {
    setReason(text);
    if (chips.includes(text)) {
      setSelectedChip(text);
    } else {
      setSelectedChip(null);
    }
  };

  return (
    <Modal visible={visible} transparent animationType="slide" onRequestClose={onClose}>
      <KeyboardAvoidingView
        behavior={Platform.OS === 'ios' ? 'padding' : 'height'}
        style={{ flex: 1 }}
      >
        <TouchableOpacity style={styles.modalBackdrop} activeOpacity={1} onPress={onClose}>
          <TouchableOpacity style={styles.reportModalContainer} activeOpacity={1}>
            <View style={styles.modalHeader}>
              <View style={styles.modalHeaderTitle}>
                <Flag size={20} color={C.red} fill={C.red} />
                <Text style={styles.modalTitle}>Report Message</Text>
              </View>
              <TouchableOpacity onPress={onClose}>
                <X size={24} color={C.muted} />
              </TouchableOpacity>
            </View>

            <Text style={styles.modalSubtitle}>Why are you reporting this message?</Text>

            <View style={styles.chipContainer}>
              {chips.map(chip => (
                <TouchableOpacity
                  key={chip}
                  onPress={() => handlePressChip(chip)}
                  style={[
                    styles.reasonChip,
                    selectedChip === chip && styles.reasonChipSelected
                  ]}
                >
                  <Text style={[
                    styles.reasonChipText,
                    selectedChip === chip && styles.reasonChipTextSelected
                  ]}>{chip}</Text>
                </TouchableOpacity>
              ))}
            </View>

            <TextInput
              style={styles.customReasonInput}
              placeholder="Or enter a custom reason..."
              placeholderTextColor={C.muted}
              value={selectedChip ? '' : reason}
              onChangeText={handleCustomChange}
              multiline
            />

            <TouchableOpacity
              style={[styles.submitReportBtn, !reason.trim() && { opacity: 0.5 }]}
              disabled={!reason.trim()}
              onPress={() => {
                onSubmit(reason.trim());
                setReason('');
                setSelectedChip(null);
              }}
            >
              <Text style={styles.submitReportText}>Submit Report</Text>
            </TouchableOpacity>
          </TouchableOpacity>
        </TouchableOpacity>
      </KeyboardAvoidingView>
    </Modal>
  );
}

// ── ActionSuccessModal ────────────────────────────────────────────────────────

function ActionSuccessModal({
  visible,
  title,
  message,
  icon,
  onClose
}: {
  visible: boolean;
  title: string;
  message: string;
  icon?: 'check' | 'clock' | 'info';
  onClose: () => void;
}) {
  const isClock = icon === 'clock';
  const isInfo = icon === 'info';
  const iconColor = isClock ? C.amber : isInfo ? C.teal : C.green;
  const iconBg = isClock ? C.amber + '20' : isInfo ? C.teal + '20' : C.green + '20';

  return (
    <Modal visible={visible} transparent animationType="fade" onRequestClose={onClose}>
      <TouchableOpacity style={styles.modalBackdrop} activeOpacity={1} onPress={onClose}>
        <View style={styles.successModalContainer} onStartShouldSetResponder={() => true}>
          <View style={[styles.successIconWrap, { backgroundColor: iconBg }]}>
            {isClock ? (
              <Clock size={32} color={iconColor} />
            ) : isInfo ? (
              <Flag size={28} color={iconColor} />
            ) : (
              <Check size={32} color={iconColor} />
            )}
          </View>
          <Text style={styles.successTitle}>{title}</Text>
          <Text style={styles.successMessage}>{message}</Text>
          <TouchableOpacity style={styles.successBtn} onPress={onClose}>
            <Text style={styles.successBtnText}>Dismiss</Text>
          </TouchableOpacity>
        </View>
      </TouchableOpacity>
    </Modal>
  );
}

// ── ActionConfirmModal ────────────────────────────────────────────────────────

function ActionConfirmModal({
  visible,
  title,
  message,
  confirmText,
  cancelText,
  isDestructive,
  onConfirm,
  onCancel
}: {
  visible: boolean;
  title: string;
  message: string;
  confirmText: string;
  cancelText: string;
  isDestructive: boolean;
  onConfirm: () => void;
  onCancel: () => void;
}) {
  return (
    <Modal visible={visible} transparent animationType="fade" onRequestClose={onCancel}>
      <TouchableOpacity style={styles.modalBackdrop} activeOpacity={1} onPress={onCancel}>
        <View style={styles.successModalContainer} onStartShouldSetResponder={() => true}>
          <View style={[styles.successIconWrap, isDestructive && { backgroundColor: C.red + '20' }]}>
            {isDestructive ? <AlertTriangle size={32} color={C.red} /> : <AlertTriangle size={32} color={C.amber} />}
          </View>
          <Text style={styles.successTitle}>{title}</Text>
          <Text style={styles.successMessage}>{message}</Text>
          <View style={{ flexDirection: 'row', gap: 12, width: '100%' }}>
            <TouchableOpacity style={[styles.successBtn, { flex: 1, backgroundColor: 'transparent', borderWidth: 1, borderColor: C.border }]} onPress={onCancel}>
              <Text style={[styles.successBtnText, { color: C.text }]}>{cancelText}</Text>
            </TouchableOpacity>
            <TouchableOpacity style={[styles.successBtn, { flex: 1, backgroundColor: isDestructive ? C.red : C.teal }]} onPress={onConfirm}>
              <Text style={styles.successBtnText}>{confirmText}</Text>
            </TouchableOpacity>
          </View>
        </View>
      </TouchableOpacity>
    </Modal>
  );
}

// ── Main screen ────────────────────────────────────────────────────────────

export default function ChatScreen() {
  const router = useRouter();
  const { state, setPublicKey, setPeerId, setConnectionStatus, setModerator, setError, setMessages, setFetching, removeMessage, addReportedSign, showNotice, hideNotice } = useAppStore();
  const colors = getAppColors(state.isIncognito);
  const { toggleSidebar, isOpen } = useSidebar();
  const [input, setInput] = useState('');
  const [sending, setSending] = useState(false);
  const [debugLog, setDebugLog] = useState<string[]>([]);
  const [selectedMessage, setSelectedMessage] = useState<RetMsgCert | null>(null);
  const [detailsVisible, setDetailsVisible] = useState(false);
  const [reportModalVisible, setReportModalVisible] = useState(false);
  const [reportMessageCert, setReportMessageCert] = useState<RetMsgCert | null>(null);
  const [successModalConfig, setSuccessModalConfig] = useState({ visible: false, title: '', message: '' });
  const [confirmModalConfig, setConfirmModalConfig] = useState({ visible: false, title: '', message: '', onConfirm: () => {}, isDestructive: true });
  const fetchIntervalRef = useRef<ReturnType<typeof setInterval> | null>(null);
  const fabScale = useRef(new Animated.Value(1)).current;
  const insets = useSafeAreaInsets();

  const log = useCallback((msg: string) => {
    const line = `[${new Date().toISOString().slice(11, 23)}] ${msg}`;
    console.log('[LibrInit]', line);
    setDebugLog(prev => [...prev.slice(-30), line]);
  }, []);

  // ── Initialisation ───────────────────────────────────────────────────────

  const initialize = useCallback(async () => {
    if (state.connectionStatus === 'initializing' || state.connectionStatus === 'connected') return;
    setConnectionStatus('initializing');
    setError(null);
    setDebugLog([]);

    try {
      // 1. Init keys + discovery server
      log('Step 1: initApp…');
      let pubKey = '';
      try {
        if (typeof (LibrCore as any).initApp === 'function') {
          pubKey = await (LibrCore as any).initApp('');
          log(`Step 1 OK: pubKey=${pubKey ? pubKey.slice(0, 16) + '…' : '(empty)'}`);
        } else {
          log('Step 1 SKIP: initApp not in AAR');
        }
      } catch (e: any) {
        log(`Step 1 ERROR: ${e?.message}`);
      }
      if (pubKey) setPublicKey(pubKey);

      // 2. Get relay addresses
      log('Step 2: getRelayAddresses…');
      let relayAddrs: string[] = [];
      try {
        if (typeof (LibrCore as any).getRelayAddresses === 'function') {
          const relayJSON: string = await (LibrCore as any).getRelayAddresses();
          log(`Step 2 raw: ${relayJSON?.slice(0, 80)}`);
          if (relayJSON && !relayJSON.startsWith('error:')) {
            relayAddrs = JSON.parse(relayJSON);
          }
          log(`Step 2 OK: ${relayAddrs.length} relay(s)`);
        } else {
          log('Step 2 SKIP: getRelayAddresses not in AAR');
        }
      } catch (e: any) {
        log(`Step 2 ERROR: ${e?.message}`);
      }

      // 3. Start libp2p node
      log(`Step 3: initNode with ${relayAddrs.length} relay(s)…`);
      const nodeResult: string = await LibrCore.initNode(JSON.stringify(relayAddrs));
      log(`Step 3 result: ${nodeResult}`);
      if (nodeResult !== 'success' && nodeResult !== 'already_initialized') {
        throw new Error(`Node init failed: ${nodeResult}`);
      }

      // 4. Get peer ID
      log('Step 4: getPeerID…');
      const id = await LibrCore.getPeerID();
      log(`Step 4 OK: ${id?.slice(0, 20)}…`);
      setPeerId(id);
      setConnectionStatus('connected');
      log('Connected!');

      // Check moderator
      const mod = await LibrCore.amIMod();
      setModerator(mod);

      // 5. Initial message fetch — delayed so state settles first
      setTimeout(() => doFetch(), 500);
    } catch (err: any) {
      const msg = err?.message ?? String(err) ?? 'Initialization failed';
      log(`FATAL: ${msg}`);
      setConnectionStatus('error');
      setError(msg);
    }
  }, [state.connectionStatus, log]);

  // ── Message fetching ─────────────────────────────────────────────────────

  const doFetch = useCallback(async () => {
    if (state.isFetching) return;
    if (typeof (LibrCore as any).fetchMessages !== 'function') return;
    setFetching(true);
    try {
      const raw: string = await (LibrCore as any).fetchMessages();
      if (raw && !raw.startsWith('error:')) {
        const certs: RetMsgCert[] = JSON.parse(raw) ?? [];
        setMessages(certs);
      }
    } catch { /* ignore */ } finally {
      setFetching(false);
    }
  }, [state.isFetching]);

  // ── Lifecycle ────────────────────────────────────────────────────────────

  useEffect(() => {
    // Listen for push messages from libp2p
    const sub = LibrCoreEvents?.addListener('onLibrMessage', () => doFetch());

    // Auto connect on launch
    initialize();

    return () => {
      sub?.remove();
    };
  }, [initialize]);

  // ── Send ─────────────────────────────────────────────────────────────────

  const handleSend = useCallback(async () => {
    const text = input.trim();
    if (!text || sending || state.connectionStatus !== 'connected') return;

    setSending(true);
    try {
      if (typeof (LibrCore as any).sendTextMessage !== 'function') {
        Alert.alert('Not supported', 'Please rebuild the Go bridge and reinstall the app.');
        return;
      }
      const raw: string = await (LibrCore as any).sendTextMessage(text);
      const result = JSON.parse(raw);
      const status: string = result.status ?? '';
      if (status === 'sent' || status.startsWith('sent:')) {
        const dbStatus = status.includes(':') ? status.slice(status.indexOf(':') + 1) : '';
        setInput('');
        setTimeout(() => doFetch(), 1500);
        if (dbStatus.startsWith('db_error:')) {
          log(`Sent but DB error: ${dbStatus}`);
        }
      } else {
        Alert.alert('Send failed', status || 'Unknown error');
      }
    } catch (err: any) {
      Alert.alert('Error', err?.message ?? 'Unknown error');
    } finally {
      setSending(false);
    }
  }, [input, sending, state.connectionStatus]);

  // ── Report ───────────────────────────────────────────────────────────────

  const handleReport = useCallback(async (cert: RetMsgCert) => {
    if (state.isIncognito) {
      Alert.alert('Reporting Disabled', 'Turn off incognito to report messages.');
      return;
    }
    setReportMessageCert(cert);
    setDetailsVisible(false);
    setTimeout(() => setReportModalVisible(true), 300); // Small delay for detail modal to close
  }, [state.isIncognito]);

  const submitReport = async (reason: string) => {
    if (!reportMessageCert) return;
    setReportModalVisible(false);

    if (state.isIncognito) {
      Alert.alert('Reporting Disabled', 'Turn off incognito to report messages.');
      return;
    }

    if (typeof (LibrCore as any).reportMessage !== 'function') {
      Alert.alert('Not supported', 'Please rebuild the Go bridge.');
      return;
    }

    try {
      const msgCert: MsgCert = {
        public_key: reportMessageCert.public_key,
        msg: reportMessageCert.msg,
        mod_certs: reportMessageCert.mod_certs,
        sign: reportMessageCert.sign,
        reason: reason,
        type: 'report'
      };
      const result: string = await (LibrCore as any).reportMessage(JSON.stringify(msgCert), reason);
      if (result === 'ok') {
        addReportedSign(reportMessageCert.sign, reportMessageCert);
        showNotice('Reported', 'Message has been flagged for review.', 'check');
      } else {
        Alert.alert('Report failed', result);
      }
    } catch (err: any) {
      Alert.alert('Error', err?.message ?? 'Unknown error');
    }
  };


  const handleRetry = useCallback(async (cert: RetMsgCert) => {
    // 1. Remove the old temporary/failed message
    removeMessage(cert.sign);
    
    setSending(true);
    try {
      let raw: string;
      
      // OPTIMIZATION: If the message is already signed by mods, only retry the storage (DB) part.
      if (cert.mod_certs && cert.mod_certs.length > 0) {
        console.log('[Chat] Retrying storage only for signed message:', cert.sign);
        // Map RetMsgCert back to MsgCert for the Go backend
        const msgCert: MsgCert = {
          public_key: cert.public_key,
          msg: cert.msg,
          mod_certs: cert.mod_certs,
          sign: cert.sign
        };
        raw = await LibrCore.resendToDb(JSON.stringify(msgCert));
      } else {
        // FULL RETRY: Redo the moderation flow from scratch
        console.log('[Chat] Retrying full moderation flow for unsigned message');
        let rawContent = cert.msg.content;
        const matchBody = rawContent.match(/<BODY>(.*?)<\/BODY>/s);
        const plainText = (matchBody ? matchBody[1].trim() : rawContent).replace(/<\/?[^>]+(>|$)/g, '').trim();
        raw = await (LibrCore as any).sendTextMessage(plainText);
      }

      const result = JSON.parse(raw);
      const status: string = result.status ?? '';
      if (status === 'sent' || status.startsWith('sent:')) {
        setTimeout(() => doFetch(), 1500);
      } else {
        Alert.alert('Retry failed', status || 'Unknown error');
      }
    } catch (err: any) {
      Alert.alert('Error', err?.message ?? 'Unknown error');
    } finally {
      setSending(false);
    }
  }, [removeMessage, doFetch]);

  const handleDelete = useCallback(async (cert: RetMsgCert) => {
    if (typeof (LibrCore as any).deleteMessage !== 'function') {
      Alert.alert('Not supported', 'Please rebuild the Go bridge.');
      return;
    }
    setConfirmModalConfig({
      visible: true,
      title: 'Delete message',
      message: 'Are you sure? This will send a delete request to the network.',
      isDestructive: true,
      onConfirm: async () => {
        setConfirmModalConfig(prev => ({ ...prev, visible: false }));
        try {
          const msgCert: MsgCert = {
            public_key: cert.public_key,
            msg: cert.msg,
            mod_certs: cert.mod_certs,
            sign: cert.sign,
          };
          const result: string = await (LibrCore as any).deleteMessage(JSON.stringify(msgCert));
          if (result === 'ok') {
            removeMessage(cert.sign);
            setDetailsVisible(false);
            setSuccessModalConfig({ visible: true, title: 'Deleted', message: 'The message has been removed from the network.' });
          } else {
            Alert.alert('Delete failed', result);
          }
        } catch (err: any) {
          Alert.alert('Error', err?.message ?? 'Unknown error');
        }
      }
    });
  }, [doFetch]);

  // ── Status helpers ─────────────────────────────────────────────────────────

  const isConnected = state.connectionStatus === 'connected';
  const isInit = state.connectionStatus === 'initializing';
  const isIdle = state.connectionStatus === 'idle';
  const isError = state.connectionStatus === 'error';

  const statusDotColor =
    isConnected ? C.green :
      isError ? C.red :
        isInit ? C.amber : C.muted;

  const statusLabel =
    isIdle ? 'Not connected' :
      isInit ? 'Connecting…' :
        isConnected ? 'Connected' :
          'Error';

  const showDebug = (isInit || isError) && debugLog.length > 0;

  const animateFab = (toValue: number) => {
    Animated.spring(fabScale, {
      toValue,
      useNativeDriver: true,
      tension: 100,
      friction: 5,
    }).start();
  };

  // ── Render ────────────────────────────────────────────────────────────────

  return (
    <SafeAreaView style={[styles.root, { backgroundColor: colors.background }]} edges={['top', 'left', 'right']}>
      <StatusBar barStyle="light-content" translucent={false} backgroundColor={colors.background} />

      {/* Header */}
      <View style={[styles.header, { marginTop: 6, backgroundColor: colors.primary }]}>
        <View style={styles.headerLeft}>
          <View style={styles.hamburgerWrap}>
            <AnimatedHamburger isOpen={isOpen} onPress={toggleSidebar} color={C.text} />
          </View>
          <View style={[styles.logoWrap, { backgroundColor: state.isIncognito ? 'rgba(155,111,212,0.18)' : C.tealDim }]}>
            <Hash size={18} color={colors.tint} />
          </View>
          <Text style={[styles.channelName, { color: colors.text }]}>libr</Text>
        </View>

      </View>



      {/* Debug log */}
      {showDebug && (
        <View style={styles.debugPanel}>
          <Text style={styles.debugTitle}>Init log</Text>
          {debugLog.map((line, i) => (
            <Text key={i} style={styles.debugLine}>{line}</Text>
          ))}
        </View>
      )}

      {/* Initial Connection Animation Overlay */}
      {!isConnected && !isError && (
        <ConnectingScreen status={statusLabel} logLines={debugLog} />
      )}

      {/* Message list */}
      <FlatList
        data={state.messages}
        keyExtractor={(item) => item.sign}
        renderItem={({ item }) => (
                  <MessageCard
            cert={item}
            myPublicKey={state.publicKey}
            isIncognito={state.isIncognito}
            onLongPress={(c) => {
              setSelectedMessage(c);
              setDetailsVisible(true);
            }}
          />
        )}
        refreshControl={
          <RefreshControl
            refreshing={state.isFetching}
            onRefresh={doFetch}
            tintColor={colors.tint}
            colors={[colors.tint]}
          />
        }
        contentContainerStyle={styles.messageList}
        ListEmptyComponent={
          <Text style={styles.emptyText}>
            {isConnected
              ? 'No messages yet. Be the first to post!'
              : isError
                ? `Connection error: ${state.lastError ?? 'unknown'}`
                : 'Tap Connect to join the network.'}
          </Text>
        }
      />

      {/* Details Modal */}
      <MessageDetailModal
        visible={detailsVisible}
        onClose={() => setDetailsVisible(false)}
        cert={selectedMessage}
        isMine={selectedMessage?.public_key === state.publicKey}
        onDelete={handleDelete}
        onReport={handleReport}
        onRetry={handleRetry}
      />

      <ReportReasonModal
        visible={reportModalVisible}
        onClose={() => setReportModalVisible(false)}
        onSubmit={submitReport}
      />

      <ActionSuccessModal
        visible={!!state.noticeModal?.visible}
        title={state.noticeModal?.title ?? ''}
        message={state.noticeModal?.message ?? ''}
        icon={state.noticeModal?.icon}
        onClose={hideNotice}
      />

      <ActionConfirmModal
        visible={confirmModalConfig.visible}
        title={confirmModalConfig.title}
        message={confirmModalConfig.message}
        confirmText="Delete"
        cancelText="Cancel"
        isDestructive={confirmModalConfig.isDestructive}
        onConfirm={confirmModalConfig.onConfirm}
        onCancel={() => setConfirmModalConfig(prev => ({ ...prev, visible: false }))}
      />

      {/* Floating Action Button */}
      <Animated.View style={{
        position: 'absolute',
        bottom: 24,
        right: 24,
        transform: [{ scale: fabScale }]
      }}>
        <TouchableOpacity
          style={[styles.fab, { backgroundColor: state.isIncognito ? '#4a2f69' : '#123f48', borderColor: state.isIncognito ? 'rgba(155,111,212,0.32)' : 'rgba(31,164,169,0.28)' }]}
          activeOpacity={0.8}
          onPressIn={() => animateFab(0.9)}
          onPressOut={() => animateFab(1)}
          onPress={() => {
            if (!isConnected) {
              Alert.alert("Not connected", "Please wait for connection to create a message.");
              return;
            }
            router.push('/modal');
          }}
        >
          <Plus size={22} color={colors.text} strokeWidth={2.5} />
          <Text style={styles.fabText}>Create</Text>
        </TouchableOpacity>
      </Animated.View>
    </SafeAreaView>
  );
}

// ── Styles ─────────────────────────────────────────────────────────────────

const styles = StyleSheet.create({
  root: { flex: 1, backgroundColor: C.bg },

  // Header
  header: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'space-between',
    marginHorizontal: 12,
    paddingHorizontal: 20,
    paddingVertical: 10,
    backgroundColor: C.card,
    borderRadius: 24,
    shadowColor: '#000',
    shadowOffset: { width: 0, height: 8 },
    shadowOpacity: 0.18,
    shadowRadius: 14,
    elevation: 7,
  },
  headerLeft: { flexDirection: 'row', alignItems: 'center', gap: 2, marginLeft: -6 },
  hamburgerWrap: {
    width: 34,
    height: 34,
    borderRadius: 6,
    backgroundColor: 'transparent',
    justifyContent: 'center',
    alignItems: 'center',
    marginRight: 6,
  },
  logoWrap: {
    width: 28,
    height: 28,
    marginRight: 6,
    borderRadius: 8,
    backgroundColor: C.tealDim,
    alignItems: 'center',
    justifyContent: 'center',
  },
  channelName: { fontSize: 22, fontFamily: Fonts.bold, color: C.text },



  // Debug panel
  debugPanel: {
    maxHeight: 160,
    marginHorizontal: 12,
    marginTop: 8,
    padding: 10,
    backgroundColor: C.debugBg,
    borderRadius: 10,
    borderWidth: 1,
    borderColor: C.border,
  },
  debugTitle: { color: '#a0a0c0', fontSize: 11, fontFamily: Fonts.bold, marginBottom: 4 },
  debugLine: {
    color: '#c0ffc0',
    fontSize: 10,
    fontFamily: Fonts.mono,
    lineHeight: 16,
  },

  // Message list
  messageList: { padding: 12, paddingBottom: 24 },
  centered: { flex: 1, alignItems: 'center', justifyContent: 'center', gap: 14 },
  centeredText: { color: C.muted, fontSize: 14, fontFamily: Fonts.sans },
  emptyText: {
    textAlign: 'center',
    marginTop: 60,
    color: C.muted,
    fontSize: 14,
    fontFamily: Fonts.sans,
    paddingHorizontal: 32,
    lineHeight: 22,
  },

  // Message card
  card: {
    flexDirection: 'row',
    alignItems: 'flex-start',
    backgroundColor: C.card,
    borderRadius: 16,
    borderWidth: 1,
    borderColor: 'rgba(30,43,60,0.45)',
    paddingHorizontal: 14,
    paddingVertical: 12,
    marginBottom: 8,
    gap: 12,
  },
  cardAvatar: { width: 36, alignItems: 'center', paddingTop: 2 },
  avatarImage: { width: 36, height: 36, borderRadius: 18 },
  avatarFallback: {
    width: 36,
    height: 36,
    borderRadius: 18,
    alignItems: 'center',
    justifyContent: 'center',
  },
  avatarInitial: { fontSize: 15, fontFamily: Fonts.medium },

  cardBody: { flex: 1, gap: 4 },
  cardMeta: { flexDirection: 'row', alignItems: 'center', gap: 8 },
  aliasText: { fontSize: 13, fontFamily: Fonts.medium, color: '#ffffff', flexShrink: 1 },
  titleText: { fontSize: 16, fontFamily: Fonts.medium, color: C.text, marginBottom: 2 },
  contentText: { color: C.text, fontSize: 15, fontFamily: Fonts.sans, lineHeight: 22 },
  timeText: { color: C.muted, fontSize: 11, fontFamily: Fonts.sans, alignSelf: 'flex-end', marginTop: 4 },

  // Floating Action Button
  fab: {
    flexDirection: 'row',
    alignItems: 'center',
    backgroundColor: '#123f48',
    borderWidth: 1,
    borderColor: 'rgba(31,164,169,0.28)',
    paddingHorizontal: 18,
    paddingVertical: 11,
    borderRadius: 18,
    gap: 8,
  },
  fabText: {
    color: C.text,
    fontSize: 14,
    fontFamily: Fonts.bold,
  },

  // Modal Details
  modalBackdrop: {
    flex: 1,
    backgroundColor: 'rgba(0,0,0,0.7)',
    justifyContent: 'center',
    alignItems: 'center',
    padding: 30,
  },
  modalContent: {
    width: '100%',
    backgroundColor: C.card,
    borderRadius: 16,
    padding: 20,
    borderWidth: 1,
    borderColor: C.border,
    shadowColor: '#000',
    shadowOffset: { width: 0, height: 10 },
    shadowOpacity: 0.5,
    shadowRadius: 15,
    elevation: 20,
  },
  modalHeader: {
    flexDirection: 'row',
    alignItems: 'center',
    marginBottom: 16,
    gap: 8,
  },
  modalTimeLabel: {
    color: C.muted,
    fontSize: 14,
    fontFamily: Fonts.medium,
  },
  modalTimeValue: {
    color: C.text,
    fontSize: 14,
    fontFamily: Fonts.sans,
    flex: 1,
    textAlign: 'right',
  },
  statusRow: {
    flexDirection: 'row',
    alignItems: 'center',
    marginBottom: 16,
    gap: 8,
  },
  statusLabel: {
    color: C.text,
    fontSize: 16,
    fontFamily: Fonts.medium,
  },
  modList: {
    maxHeight: 200,
    backgroundColor: 'rgba(0,0,0,0.2)',
    borderRadius: 8,
    padding: 10,
    marginBottom: 20,
  },
  modItem: {
    flexDirection: 'row',
    alignItems: 'center',
    marginBottom: 8,
    paddingVertical: 6,
    paddingHorizontal: 8,
    backgroundColor: C.bg,
    borderRadius: 6,
    gap: 10,
  },
  modAlias: {
    flex: 1,
    color: C.text,
    fontSize: 14,
    fontFamily: Fonts.bold,
  },
  modSignLabel: {
    color: C.muted,
    fontSize: 12,
    fontFamily: Fonts.sans,
  },
  modalActions: {
    borderTopWidth: 1,
    borderTopColor: C.border,
    paddingTop: 16,
  },
  actionBtn: {
    flexDirection: 'row',
    alignItems: 'center',
    paddingVertical: 8,
    gap: 10,
  },
  actionBtnText: {
    fontSize: 14,
    fontFamily: Fonts.bold,
  },

  // Image Support
  imageGallery: {
    flexDirection: 'row',
    flexWrap: 'wrap',
    gap: 8,
    marginTop: 8,
  },
  messageImage: {
    width: 200,
    height: 150,
    borderRadius: 12,
    backgroundColor: C.border,
  },
  lightboxBackdrop: {
    flex: 1,
    backgroundColor: 'rgba(0,0,0,0.95)',
    justifyContent: 'center',
    alignItems: 'center',
  },
  lightboxImage: {
    width: '100%',
    height: '100%',
  },
  lightboxClose: {
    position: 'absolute',
    top: 50,
    right: 20,
    padding: 10,
    backgroundColor: 'rgba(0,0,0,0.5)',
    borderRadius: 25,
  },
  reportModalContainer: {
    backgroundColor: C.card,
    borderRadius: 24,
    padding: 24,
    paddingBottom: Platform.OS === 'ios' ? 40 : 24,
    marginBottom: 24,
    marginHorizontal: 16,
    borderWidth: 1,
    borderColor: C.border,
    shadowColor: '#000',
    shadowOffset: { width: 0, height: 10 },
    shadowOpacity: 0.5,
    shadowRadius: 20,
    elevation: 10,
  },
  modalHeaderTitle: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: 10,
  },
  modalTitle: {
    fontSize: 20,
    fontFamily: Fonts.bold,
    color: C.text,
  },
  modalSubtitle: {
    fontSize: 14,
    color: C.muted,
    fontFamily: Fonts.sans,
    marginBottom: 20,
  },
  chipContainer: {
    flexDirection: 'row',
    flexWrap: 'wrap',
    gap: 10,
    marginBottom: 20,
  },
  reasonChip: {
    paddingHorizontal: 16,
    paddingVertical: 8,
    borderRadius: 20,
    backgroundColor: 'rgba(255,255,255,0.05)',
    borderWidth: 1,
    borderColor: C.border,
  },
  reasonChipSelected: {
    backgroundColor: 'rgba(239, 68, 68, 0.2)',
    borderColor: C.red,
  },
  reasonChipText: {
    color: C.text,
    fontSize: 14,
    fontFamily: Fonts.medium,
  },
  reasonChipTextSelected: {
    color: C.red,
  },
  customReasonInput: {
    backgroundColor: 'rgba(0,0,0,0.3)',
    borderRadius: 12,
    padding: 16,
    color: C.text,
    fontSize: 15,
    fontFamily: Fonts.sans,
    minHeight: 80,
    textAlignVertical: 'top',
    borderWidth: 1,
    borderColor: C.border,
    marginBottom: 24,
  },
  submitReportBtn: {
    backgroundColor: C.red,
    borderRadius: 14,
    paddingVertical: 16,
    alignItems: 'center',
    shadowColor: C.red,
    shadowOffset: { width: 0, height: 4 },
    shadowOpacity: 0.3,
    shadowRadius: 8,
    elevation: 4,
  },
  submitReportText: {
    color: '#fff',
    fontSize: 16,
    fontFamily: Fonts.bold,
  },
  successModalContainer: {
    width: 280,
    backgroundColor: C.card,
    borderRadius: 24,
    padding: 24,
    alignItems: 'center',
    borderWidth: 1,
    borderColor: C.border,
    shadowColor: '#000',
    shadowOffset: { width: 0, height: 10 },
    shadowOpacity: 0.5,
    shadowRadius: 15,
    elevation: 20,
  },
  successIconWrap: {
    width: 64,
    height: 64,
    borderRadius: 32,
    backgroundColor: C.green + '20',
    alignItems: 'center',
    justifyContent: 'center',
    marginBottom: 16,
  },
  successTitle: {
    fontSize: 20,
    fontFamily: Fonts.bold,
    color: C.text,
    marginBottom: 8,
    textAlign: 'center',
  },
  successMessage: {
    fontSize: 15,
    color: C.muted,
    fontFamily: Fonts.sans,
    textAlign: 'center',
    marginBottom: 24,
    lineHeight: 22,
  },
  successBtn: {
    backgroundColor: C.teal,
    paddingHorizontal: 24,
    paddingVertical: 12,
    borderRadius: 12,
    width: '100%',
    alignItems: 'center',
  },
  successBtnText: {
    color: '#fff',
    fontSize: 16,
    fontFamily: Fonts.bold,
    textAlign: 'center'
  }
});


