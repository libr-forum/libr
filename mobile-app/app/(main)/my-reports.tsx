import React, { useEffect, useState, useRef, useCallback } from 'react';
import { StyleSheet, View, Text, FlatList, TouchableOpacity, StatusBar, Modal } from 'react-native';
import AsyncStorage from '@react-native-async-storage/async-storage';
import { SafeAreaView, useSafeAreaInsets } from 'react-native-safe-area-context';
import { useRouter } from 'expo-router';
import { ArrowLeft, Flag, CheckCircle, Clock, XCircle, Bell, Trash2 } from 'lucide-react-native';
import { useAppStore } from '@/store/useAppStore';
import LibrCore, { RetMsgCert } from '@/modules/LibrCore';
import { Image } from 'expo-image';

type QueueItem = RetMsgCert & {
    placeholder?: boolean;
};

const C = {
    bg: '#0a0f1c',
    card: '#0f1625',
    border: '#1e2b3c',
    teal: '#1fa4a9',
    green: '#00ffd0',
    red: '#ff4444',
    text: '#ffffff',
    muted: '#8b9bb4',
    amber: '#ffb300',
};

// Styled in-app notification banner (non-blocking overlay)
function NotificationBanner({ visible, type, message, onDismiss }: { visible: boolean; type: 'approved' | 'rejected'; message: string; onDismiss: () => void }) {
    if (!visible) return null;
    const isApproved = type === 'approved';
    return (
        <View
            pointerEvents="box-none"
            style={{
                position: 'absolute',
                top: 55,
                left: 16,
                right: 16,
                zIndex: 9999,
                elevation: 10,
            }}
        >
            <TouchableOpacity activeOpacity={0.9} onPress={onDismiss}>
                <View style={{
                    backgroundColor: C.card,
                    borderRadius: 16,
                    padding: 14,
                    borderWidth: 1,
                    borderColor: isApproved ? 'rgba(0, 255, 208, 0.4)' : 'rgba(255, 68, 68, 0.4)',
                    flexDirection: 'row',
                    alignItems: 'center',
                    gap: 12,
                    shadowColor: '#000',
                    shadowOffset: { width: 0, height: 4 },
                    shadowOpacity: 0.35,
                    shadowRadius: 8,
                    elevation: 10,
                }}>
                    <View style={{
                        width: 36,
                        height: 36,
                        borderRadius: 18,
                        backgroundColor: isApproved ? 'rgba(0, 255, 208, 0.15)' : 'rgba(255, 68, 68, 0.15)',
                        alignItems: 'center',
                        justifyContent: 'center',
                    }}>
                        {isApproved ? <CheckCircle size={18} color={C.green} /> : <XCircle size={18} color={C.red} />}
                    </View>
                    <View style={{ flex: 1 }}>
                        <Text style={{ color: isApproved ? C.green : C.red, fontWeight: '700', fontSize: 14 }}>
                            {isApproved ? 'Report Upheld' : 'Report Dismissed'}
                        </Text>
                        <Text style={{ color: C.muted, fontSize: 12, marginTop: 2 }}>{message}</Text>
                    </View>
                </View>
            </TouchableOpacity>
        </View>
    );
}

export default function ModerationQueueScreen() {
    const router = useRouter();
    const { state, removeReportedSign } = useAppStore();
    const insets = useSafeAreaInsets();
    const [queueMessages, setQueueMessages] = useState<QueueItem[]>([]);
    const [pendingStatus, setPendingStatus] = useState<Record<string, { total: number, approved: number, rejected: number, final_status?: string, type?: string }>>({});
    const [loading, setLoading] = useState(true);
    const [notification, setNotification] = useState<{ visible: boolean; type: 'approved' | 'rejected'; message: string }>({ visible: false, type: 'approved', message: '' });
    const [dismissedSigns, setDismissedSigns] = useState<Set<string>>(new Set());
    // Track finalized reports: sign -> 'approved' | 'rejected'
    const [resolvedSigns, setResolvedSigns] = useState<Record<string, 'approved' | 'rejected'>>({});

    // Cache reported messages so they persist even after deletion from feed
    const cachedMessagesRef = useRef<Map<string, RetMsgCert>>(new Map());
    const prevStatusRef = useRef<Record<string, { total: number, approved: number, rejected: number, final_status?: string, type?: string }>>({});
    const lastKnownStatsRef = useRef<Record<string, { total: number, approved: number, rejected: number, final_status?: string, type?: string }>>({});
    const notifiedRef = useRef<Set<string>>(new Set());
    const isInitialLoad = useRef<boolean>(true);

    // Load persisted resolved status, stats, dismissed signs, and previously notified signs
    useEffect(() => {
        AsyncStorage.getItem('@libr_resolved_signs').then((raw: string | null) => {
            if (raw) {
                try {
                    setResolvedSigns(JSON.parse(raw));
                } catch { }
            }
        }).catch(() => { });

        AsyncStorage.getItem('@libr_resolved_stats').then((raw: string | null) => {
            if (raw) {
                try {
                    const parsed = JSON.parse(raw);
                    lastKnownStatsRef.current = { ...lastKnownStatsRef.current, ...parsed };
                } catch { }
            }
        }).catch(() => { });

        AsyncStorage.getItem('@libr_dismissed_signs').then((raw: string | null) => {
            if (raw) {
                try {
                    const arr: string[] = JSON.parse(raw);
                    setDismissedSigns(new Set(arr));
                } catch { }
            }
        }).catch(() => { });

        AsyncStorage.getItem('@libr_notified_signs').then((raw: string | null) => {
            if (raw) {
                try {
                    const arr: string[] = JSON.parse(raw);
                    for (const s of arr) notifiedRef.current.add(s);
                } catch { }
            }
        }).catch(() => { });
    }, []);

    // Show a styled notification that auto-dismisses after 3.5 seconds
    const showNotification = useCallback((type: 'approved' | 'rejected', message: string) => {
        setNotification({ visible: true, type, message });
        setTimeout(() => setNotification(prev => ({ ...prev, visible: false })), 3500);
    }, []);

    // Delete a single completed report — remove from state + Go cache + AsyncStorage
    const deleteReport = useCallback((sign: string) => {
        setDismissedSigns(prev => {
            const next = new Set(prev).add(sign);
            AsyncStorage.setItem('@libr_dismissed_signs', JSON.stringify(Array.from(next))).catch(() => {});
            return next;
        });
        setQueueMessages(prev => prev.filter(m => m.sign !== sign));
        setResolvedSigns(r => {
            const next = { ...r };
            delete next[sign];
            AsyncStorage.setItem('@libr_resolved_signs', JSON.stringify(next)).catch(() => {});
            return next;
        });
        // Call Go core bridge to delete from pending and finalized cache
        if (typeof (LibrCore as any).dismissReport === 'function') {
            (LibrCore as any).dismissReport(sign).catch(() => {});
        }
        // Remove from app store & local AsyncStorage caches
        removeReportedSign(sign);
        cachedMessagesRef.current.delete(sign);
    }, [removeReportedSign]);

    // Delete all completed reports at once
    const deleteAllCompleted = useCallback(() => {
        const completedSigns: string[] = [];
        for (const item of queueMessages) {
            const pStatus = pendingStatus[item.sign];
            const isResolved = !!(
                pStatus?.final_status ||
                resolvedSigns[item.sign] ||
                item.deleted === '1' ||
                (state.messages.length > 0 && !state.messages.some(m => m.sign === item.sign) && !pStatus)
            );
            if (isResolved) {
                completedSigns.push(item.sign);
            }
        }

        if (completedSigns.length === 0) return;

        setDismissedSigns(prev => {
            const next = new Set(prev);
            for (const s of completedSigns) next.add(s);
            AsyncStorage.setItem('@libr_dismissed_signs', JSON.stringify(Array.from(next))).catch(() => {});
            return next;
        });

        setQueueMessages(prev => prev.filter(m => !completedSigns.includes(m.sign)));

        setResolvedSigns(r => {
            const next = { ...r };
            for (const s of completedSigns) delete next[s];
            AsyncStorage.setItem('@libr_resolved_signs', JSON.stringify(next)).catch(() => {});
            return next;
        });

        for (const s of completedSigns) {
            if (typeof (LibrCore as any).dismissReport === 'function') {
                (LibrCore as any).dismissReport(s).catch(() => {});
            }
            removeReportedSign(s);
            cachedMessagesRef.current.delete(s);
        }
    }, [queueMessages, pendingStatus, resolvedSigns, state.messages, removeReportedSign]);

    // Fetch messages and cache them
    useEffect(() => {
        let cancelled = false;

        const fetchDetails = async () => {
            const pendingSigns = Object.keys(pendingStatus);
            if (state.reportedSigns.size === 0 && pendingSigns.length === 0) {
                if (!cancelled) {
                    setQueueMessages([]);
                    setLoading(false);
                }
                return;
            }

            setLoading(true);
            try {
                // Hydrate from AsyncStorage on first load
                if (cachedMessagesRef.current.size === 0) {
                    try {
                        const stored = await AsyncStorage.getItem('@libr_reported_messages');
                        if (stored) {
                            const parsed: Record<string, RetMsgCert> = JSON.parse(stored);
                            for (const [sign, cert] of Object.entries(parsed)) {
                                cachedMessagesRef.current.set(sign, cert);
                            }
                        }
                    } catch { }
                }

                // Update cache with any messages from the feed
                for (const msg of state.messages) {
                    if (state.reportedSigns.has(msg.sign) || pendingSigns.includes(msg.sign)) {
                        cachedMessagesRef.current.set(msg.sign, msg);
                    }
                }

                // For any pending/reported signs not in cache, try to fetch them from DHT
                const targetSigns = Array.from(new Set([...pendingSigns, ...Array.from(state.reportedSigns)]));
                const missingSigns = targetSigns.filter(s => !cachedMessagesRef.current.has(s));
                for (const sign of missingSigns) {
                    try {
                        if (typeof (LibrCore as any).fetchMessageBySign === 'function') {
                            const raw = await (LibrCore as any).fetchMessageBySign(sign);
                            if (raw && !raw.startsWith('error:')) {
                                cachedMessagesRef.current.set(sign, JSON.parse(raw));
                            }
                        }
                    } catch {
                        // Ignore individual fetch errors
                    }
                }

                // Detect resolved reports from feed status
                const newResolved: Record<string, 'approved' | 'rejected'> = {};
                const hasFeedMessages = state.messages.length > 0;
                const liveMessageSigns = new Set(state.messages.map(m => m.sign));

                for (const sign of state.reportedSigns) {
                    if (resolvedSigns[sign]) continue;

                    const liveMsg = state.messages.find(m => m.sign === sign);
                    if (liveMsg && liveMsg.deleted === '1') {
                        newResolved[sign] = 'approved';
                    } else if (hasFeedMessages && !liveMessageSigns.has(sign) && !pendingStatus[sign]) {
                        // Feed is loaded, the reported message is absent from chat, and not in pending moderation:
                        // That means it was removed by moderators!
                        newResolved[sign] = 'approved';
                    }
                }
                if (Object.keys(newResolved).length > 0) {
                    setResolvedSigns(prev => {
                        const next = { ...prev, ...newResolved };
                        AsyncStorage.setItem('@libr_resolved_signs', JSON.stringify(next)).catch(() => {});
                        return next;
                    });
                }

                // Build display list from the pending moderation queue.
                const results: QueueItem[] = [];
                for (const sign of targetSigns) {
                    if (dismissedSigns.has(sign)) continue;
                    const cached = cachedMessagesRef.current.get(sign);
                    if (cached) {
                        results.push(cached);
                    } else {
                        results.push({
                            public_key: '',
                            sign,
                            msg: {
                                content: '',
                                ts: 0,
                            },
                            mod_certs: [],
                            deleted: '0',
                            placeholder: true,
                        });
                    }
                }

                if (!cancelled) setQueueMessages(results);
            } catch (err) {
                console.warn('Failed to fetch reported messages details:', err);
            } finally {
                if (!cancelled) setLoading(false);
            }
        };

        fetchDetails();
        return () => { cancelled = true; };
    }, [state.reportedSigns, state.messages, dismissedSigns, pendingStatus, resolvedSigns]);

    // Poll pendingStatus every 4 seconds
    useEffect(() => {
        const pollStatus = async () => {
            if (typeof (LibrCore as any).getPendingReports !== 'function') return;
            try {
                const rawStatus = await (LibrCore as any).getPendingReports();
                if (rawStatus && !rawStatus.startsWith('error:')) {
                    const newPending: Record<string, { total: number, approved: number, rejected: number, final_status?: string, type?: string }> = JSON.parse(rawStatus);
                    const prev = prevStatusRef.current;
                    const newResolved: Record<string, 'approved' | 'rejected'> = {};
                    const newStatsToPersist: Record<string, any> = {};

                    const isFirstPoll = isInitialLoad.current;

                    for (const sign in newPending) {
                        const item = newPending[sign];
                        lastKnownStatsRef.current[sign] = item;

                        // Check if Go core finalized this item
                        if (item.final_status === 'approved' || item.final_status === 'rejected') {
                            const fs = item.final_status as 'approved' | 'rejected';
                            newResolved[sign] = fs;
                            newStatsToPersist[sign] = item;

                            // On initial screen load, do NOT pop up banners for already-completed reports!
                            if (isFirstPoll) {
                                notifiedRef.current.add(sign);
                            } else if (!notifiedRef.current.has(sign)) {
                                notifiedRef.current.add(sign);
                                const isRep = state.reportedSigns.has(sign) || item.type === 'report';
                                showNotification(
                                    fs,
                                    fs === 'approved'
                                        ? (isRep ? 'Your report was upheld — the message has been removed.' : 'Your image was approved and published to chat.')
                                        : (isRep ? 'Your report was dismissed — the message will stay.' : 'Your image was rejected by moderators.')
                                );
                            }
                        } else if (item.total > 0) {
                            // Check quorum threshold
                            const threshold = Math.floor(item.total / 2);
                            if (item.approved > threshold) {
                                newResolved[sign] = 'approved';
                                newStatsToPersist[sign] = item;
                            } else if (item.rejected > threshold) {
                                newResolved[sign] = 'rejected';
                                newStatsToPersist[sign] = item;
                            }
                        }
                    }

                    // Also detect signs that disappeared from pendingStatus (if not in newPending)
                    if (Object.keys(prev).length > 0) {
                        for (const sign in prev) {
                            if (!(sign in newPending) && !resolvedSigns[sign]) {
                                const reportApproved = prev[sign].approved > prev[sign].rejected;
                                const finalStatus = reportApproved ? 'approved' : 'rejected';
                                newResolved[sign] = finalStatus;
                                newStatsToPersist[sign] = prev[sign];

                                if (isFirstPoll) {
                                    notifiedRef.current.add(sign);
                                } else if (!notifiedRef.current.has(sign)) {
                                    notifiedRef.current.add(sign);
                                    showNotification(finalStatus,
                                        reportApproved
                                            ? 'Your report was upheld — the message has been removed.'
                                            : 'Your report was dismissed — the message will stay.'
                                    );
                                }
                            }
                        }
                    }

                    if (isFirstPoll) {
                        isInitialLoad.current = false;
                        AsyncStorage.setItem('@libr_notified_signs', JSON.stringify(Array.from(notifiedRef.current))).catch(() => {});
                    } else if (Object.keys(newResolved).length > 0) {
                        AsyncStorage.setItem('@libr_notified_signs', JSON.stringify(Array.from(notifiedRef.current))).catch(() => {});
                    }

                    if (Object.keys(newResolved).length > 0) {
                        setResolvedSigns(r => {
                            const next = { ...r, ...newResolved };
                            AsyncStorage.setItem('@libr_resolved_signs', JSON.stringify(next)).catch(() => {});
                            return next;
                        });
                        AsyncStorage.getItem('@libr_resolved_stats').then(raw => {
                            const existing = raw ? JSON.parse(raw) : {};
                            AsyncStorage.setItem('@libr_resolved_stats', JSON.stringify({ ...existing, ...newStatsToPersist })).catch(() => {});
                        }).catch(() => {});
                    }

                    prevStatusRef.current = newPending;
                    setPendingStatus(newPending);
                }
            } catch { }
        };

        pollStatus();
        const intervalId = setInterval(pollStatus, 4000);
        return () => clearInterval(intervalId);
    }, [state.reportedSigns, showNotification, resolvedSigns]);

    const renderItem = ({ item }: { item: QueueItem }) => {
        const pStatus = pendingStatus[item.sign];
        const isReport = state.reportedSigns.has(item.sign) || pStatus?.type === 'report';

        // Resolution detection: check Go final_status, resolvedSigns map, item.deleted, or feed absence
        let resolved: 'approved' | 'rejected' | null =
            (pStatus?.final_status as 'approved' | 'rejected') ||
            resolvedSigns[item.sign] ||
            null;

        if (!resolved && item.deleted === '1') {
            resolved = 'approved';
        }
        if (!resolved && isReport && state.messages.length > 0 && !state.messages.some(m => m.sign === item.sign) && !pStatus) {
            resolved = 'approved';
        }
        if (!resolved && pStatus && pStatus.total > 0) {
            const threshold = Math.floor(pStatus.total / 2);
            if (pStatus.approved > threshold) resolved = 'approved';
            else if (pStatus.rejected > threshold) resolved = 'rejected';
        }

        const stats = pStatus || lastKnownStatsRef.current[item.sign];

        // Extract plain content
        let rawContent = item.msg.content;
        const matchBody = rawContent.match(/<BODY>(.*?)<\/BODY>/s);
        if (matchBody) rawContent = matchBody[1].trim();
        const plainContent = rawContent.replace(/<\/?[^>]+(>|$)/g, '').replace(/&amp;/g, '&').replace(/&lt;/g, '<').replace(/&gt;/g, '>').replace(/&quot;/g, '"').replace(/&#39;/g, "'").trim();

        // Extract images
        const imgSrcs: string[] = [];
        const imgRegex = /<img[^>]+src="([^">]+)"/g;
        let match;
        while ((match = imgRegex.exec(item.msg.content)) !== null) {
            imgSrcs.push(match[1]);
        }

        return (
            <View style={styles.card}>
                <View style={{ flexDirection: 'row', alignItems: 'flex-start' }}>
                    <Text style={styles.labelSpan}>Message: </Text>
                    {plainContent ? (
                        <Text style={styles.contentText} numberOfLines={10}>
                            {plainContent}
                        </Text>
                    ) : (
                        <Text style={[styles.contentText, { fontStyle: 'italic', color: C.muted }]}>
                            {item.placeholder ? `Pending moderation item\n${item.sign.slice(0, 18)}...` : '[No text]'}
                        </Text>
                    )}
                </View>

                {imgSrcs.length > 0 && (
                    <View style={styles.imageGallery}>
                        {imgSrcs.map((src, i) => (
                            <Image
                                key={i}
                                source={src}
                                style={styles.reportImage}
                                contentFit="cover"
                            />
                        ))}
                    </View>
                )}

                <View style={styles.statusRow}>
                    <View style={{ gap: 8, paddingBottom: 8, width: '100%' }}>
                        <View style={{ flexDirection: 'row', alignItems: 'center', justifyContent: 'space-between' }}>
                            <View style={[
                                styles.statusBadge,
                                resolved === 'approved' && styles.statusBadgeApproved,
                                resolved === 'rejected' && styles.statusBadgeRejected
                            ]}>
                                {resolved === 'approved' ? (
                                    <>
                                        <CheckCircle size={14} color={C.green} />
                                        <Text style={[styles.statusText, { color: C.green }]}>
                                            {isReport ? 'Report Upheld' : 'Approved'}
                                        </Text>
                                    </>
                                ) : resolved === 'rejected' ? (
                                    <>
                                        <XCircle size={14} color={C.red} />
                                        <Text style={[styles.statusText, { color: C.red }]}>
                                            {isReport ? 'Report Dismissed' : 'Rejected'}
                                        </Text>
                                    </>
                                ) : (
                                    <>
                                        <Clock size={14} color={C.amber} />
                                        <Text style={[styles.statusText, { color: C.amber }]}>
                                            {isReport ? 'Report Under Review' : 'Image Under Review'}
                                        </Text>
                                    </>
                                )}
                            </View>

                            {resolved && (
                                <TouchableOpacity
                                    onPress={() => deleteReport(item.sign)}
                                    style={styles.deleteBtn}
                                    activeOpacity={0.7}
                                    hitSlop={{ top: 8, bottom: 8, left: 8, right: 8 }}
                                >
                                    <Trash2 size={13} color={C.red} />
                                    <Text style={styles.deleteBtnText}>Delete</Text>
                                </TouchableOpacity>
                            )}
                        </View>

                        {/* Vote tally and resolution details */}
                        {resolved ? (
                            <View style={styles.resolutionContainer}>
                                <Text style={styles.resolutionText}>
                                    {resolved === 'approved'
                                        ? (isReport ? 'Moderators upheld this report. Message removed from chat.' : 'Moderators approved this image. Published to chat.')
                                        : (isReport ? 'Moderators dismissed this report. Message retained in chat.' : 'Moderators rejected this image.')
                                    }
                                </Text>
                                {stats && (stats.approved > 0 || stats.rejected > 0 || stats.total > 0) && (
                                    <Text style={styles.tallySubtext}>
                                        Votes: {stats.approved} {isReport ? 'Upheld' : 'Approved'} • {stats.rejected} {isReport ? 'Dismissed' : 'Rejected'}
                                    </Text>
                                )}
                            </View>
                        ) : (
                            stats && stats.total > 0 ? (
                                <View style={{ flexDirection: 'row', gap: 12, paddingHorizontal: 10 }}>
                                    <Text style={{ fontSize: 13, color: C.green, fontWeight: '600' }}>
                                        {isReport ? 'Upheld: ' : 'Approved: '}{stats.approved}
                                    </Text>
                                    <Text style={{ fontSize: 13, color: C.red, fontWeight: '600' }}>
                                        {isReport ? 'Dismissed: ' : 'Rejected: '}{stats.rejected}
                                    </Text>
                                    <Text style={{ fontSize: 13, color: C.amber, fontWeight: '600' }}>
                                        Waiting: {Math.max(0, stats.total - stats.approved - stats.rejected)}
                                    </Text>
                                </View>
                            ) : (
                                <Text style={{ fontSize: 12, color: C.muted, paddingHorizontal: 10 }}>Awaiting moderator responses...</Text>
                            )
                        )}
                    </View>
                </View>
            </View>
        );
    };

    const hasCompletedReports = queueMessages.some(m => {
        const pStatus = pendingStatus[m.sign];
        const res = (pStatus?.final_status as 'approved' | 'rejected') || resolvedSigns[m.sign] || m.deleted === '1';
        return !!res;
    });

    return (
        <SafeAreaView style={styles.safeArea} edges={['left', 'right', 'bottom']}>
            <StatusBar barStyle="light-content" backgroundColor={C.bg} />
            {/* Header */}
            <View style={[styles.header, { paddingTop: Math.max(insets.top, 10) }]}>
                <TouchableOpacity style={styles.backBtn} onPress={() => router.back()}>
                    <ArrowLeft size={24} color={C.text} />
                </TouchableOpacity>
                <Text style={styles.title}>Review Status</Text>
                {hasCompletedReports ? (
                    <TouchableOpacity
                        style={styles.clearHeaderBtn}
                        onPress={deleteAllCompleted}
                        activeOpacity={0.7}
                    >
                        <Trash2 size={15} color={C.red} />
                        <Text style={styles.clearHeaderText}>Clear All</Text>
                    </TouchableOpacity>
                ) : (
                    <View style={{ width: 40 }} />
                )}
            </View>

            {loading ? (
                <View style={styles.centerContainer}>
                    <Text style={styles.emptyText}>Loading...</Text>
                </View>
            ) : queueMessages.length === 0 ? (
                <View style={styles.centerContainer}>
                    <Bell size={48} color={C.border} style={{ marginBottom: 16 }} />
                    <Text style={styles.emptyText}>No reports or submissions right now.</Text>
                </View>
            ) : (
                <FlatList
                    data={queueMessages}
                    keyExtractor={(item) => item.sign}
                    renderItem={renderItem}
                    contentContainerStyle={styles.listContainer}
                />
            )}

            <NotificationBanner
                visible={notification.visible}
                type={notification.type}
                message={notification.message}
                onDismiss={() => setNotification(prev => ({ ...prev, visible: false }))}
            />
        </SafeAreaView>
    );
}

const styles = StyleSheet.create({
    safeArea: {
        flex: 1,
        backgroundColor: C.bg,
    },
    header: {
        flexDirection: 'row',
        alignItems: 'center',
        justifyContent: 'space-between',
        paddingHorizontal: 16,
        paddingVertical: 12,
        borderBottomWidth: 1,
        borderBottomColor: C.border,
    },
    backBtn: {
        padding: 8,
    },
    title: {
        fontSize: 18,
        fontWeight: '700',
        color: C.text,
    },
    centerContainer: {
        flex: 1,
        justifyContent: 'center',
        alignItems: 'center',
        padding: 20,
    },
    emptyText: {
        color: C.muted,
        fontSize: 16,
        textAlign: 'center',
    },
    listContainer: {
        padding: 16,
        gap: 16,
    },
    card: {
        backgroundColor: C.card,
        borderRadius: 20,
        padding: 20,
        borderWidth: 1,
        borderColor: C.border,
        shadowColor: '#000',
        shadowOffset: { width: 0, height: 2 },
        shadowOpacity: 0.1,
        shadowRadius: 4,
        elevation: 2,
    },
    labelSpan: {
        color: C.muted,
        fontWeight: '700',
        fontSize: 14,
        marginRight: 4,
    },
    contentText: {
        color: C.text,
        fontSize: 14,
        lineHeight: 20,
        flex: 1,
    },
    statusRow: {
        flexDirection: 'row',
        justifyContent: 'space-between',
        marginTop: 16,
        paddingTop: 16,
        borderTopWidth: 1,
        borderTopColor: C.border + '50',
    },
    statusBadge: {
        flexDirection: 'row',
        alignItems: 'center',
        gap: 6,
        backgroundColor: C.bg,
        paddingHorizontal: 10,
        paddingVertical: 4,
        borderRadius: 12,
    },
    statusBadgeApproved: {
        backgroundColor: 'rgba(0, 255, 208, 0.12)',
        borderWidth: 1,
        borderColor: 'rgba(0, 255, 208, 0.3)',
    },
    statusBadgeRejected: {
        backgroundColor: 'rgba(255, 68, 68, 0.12)',
        borderWidth: 1,
        borderColor: 'rgba(255, 68, 68, 0.3)',
    },
    deleteBtn: {
        flexDirection: 'row',
        alignItems: 'center',
        gap: 5,
        paddingHorizontal: 10,
        paddingVertical: 5,
        borderRadius: 8,
        backgroundColor: 'rgba(255, 68, 68, 0.12)',
        borderWidth: 1,
        borderColor: 'rgba(255, 68, 68, 0.25)',
    },
    deleteBtnText: {
        fontSize: 12,
        color: C.red,
        fontWeight: '600',
    },
    clearHeaderBtn: {
        flexDirection: 'row',
        alignItems: 'center',
        gap: 4,
        paddingHorizontal: 10,
        paddingVertical: 6,
        borderRadius: 8,
        backgroundColor: 'rgba(255, 68, 68, 0.12)',
        borderWidth: 1,
        borderColor: 'rgba(255, 68, 68, 0.2)',
    },
    clearHeaderText: {
        fontSize: 12,
        color: C.red,
        fontWeight: '600',
    },
    resolutionContainer: {
        paddingHorizontal: 6,
        paddingVertical: 2,
        gap: 3,
    },
    resolutionText: {
        fontSize: 13,
        color: C.text,
        fontWeight: '500',
    },
    tallySubtext: {
        fontSize: 12,
        color: C.muted,
    },
    statusText: {
        fontSize: 12,
        fontWeight: '600',
    },
    actionTakenContainer: {
        flexDirection: 'column',
        alignItems: 'flex-start',
        backgroundColor: C.bg,
        borderRadius: 12,
        overflow: 'hidden',
        paddingHorizontal: 10,
        paddingVertical: 6,
    },
    modDivider: {
        height: 1,
        width: '100%',
        backgroundColor: C.border,
    },
    modList: {
        paddingHorizontal: 10,
        paddingVertical: 6,
        flexDirection: 'row',
        alignItems: 'center',
        gap: 8,
    },
    modListTitle: {
        fontSize: 11,
        color: C.muted,
        fontWeight: '500',
    },
    modAvatarsContainer: {
        flexDirection: 'row',
        alignItems: 'center',
    },
    modAvatar: {
        width: 20,
        height: 20,
        borderRadius: 10,
        justifyContent: 'center',
        alignItems: 'center',
        marginLeft: -6,
        borderWidth: 1.5,
        borderColor: C.bg,
    },
    modAvatarText: {
        color: '#fff',
        fontSize: 9,
        fontWeight: 'bold',
    },
    modCountText: {
        fontSize: 11,
        color: C.muted,
        marginLeft: 8,
        fontWeight: '600',
    },
    imageGallery: {
        flexDirection: 'row',
        flexWrap: 'wrap',
        gap: 8,
        marginTop: 12,
    },
    reportImage: {
        width: 100,
        height: 100,
        borderRadius: 12,
        backgroundColor: '#1a2235',
        borderWidth: 1,
        borderColor: C.border,
    },
});
