import React, { useEffect, useState, useCallback } from 'react';
import { View, Text, StyleSheet, FlatList, TouchableOpacity, Alert, RefreshControl, ActivityIndicator, StatusBar } from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';
import { Colors as C, Fonts } from '@/constants/theme';
import LibrCore from '@/modules/LibrCore';
import { useAppStore } from '@/store/useAppStore';
import { useRouter } from 'expo-router';
import { ChevronLeft, Flag, Check, X, ShieldAlert, Clock, AlertTriangle } from 'lucide-react-native';
import { Image } from 'expo-image';

interface ReportCert {
    msgcert: {
        public_key: string;
        msg: {
            content: string;
            ts: number;
        };
        sign: string;
        reason?: string;
    };
    repmod_certs: any[];
    mode: string;
}

export default function QueueScreen() {
    const router = useRouter();
    const { state } = useAppStore();
    const [reports, setReports] = useState<ReportCert[]>([]);
    const [loading, setLoading] = useState(false);
    const [refreshing, setRefreshing] = useState(false);

    const fetchReports = useCallback(async (isRefresh = false) => {
        if (isRefresh) setRefreshing(true);
        else setLoading(true);

        try {
            const result = await LibrCore.fetchReports();
            if (result.startsWith('error:')) {
                Alert.alert('Error', result);
            } else {
                const parsed = JSON.parse(result);
                setReports(parsed);
            }
        } catch (err: any) {
            Alert.alert('Error', err?.message || 'Failed to fetch reports');
        } finally {
            setLoading(false);
            setRefreshing(false);
        }
    }, []);

    useEffect(() => {
        fetchReports();
    }, [fetchReports]);

    const handleModerate = async (report: ReportCert, action: 'approve' | 'reject') => {
        try {
            const result = await LibrCore.moderateMessage(JSON.stringify(report.msgcert), action);
            if (result === 'ok') {
                Alert.alert('Success', `Message ${action === 'approve' ? 'approved' : 'rejected'}.`);
                // Remove from local list
                setReports(prev => prev.filter(r => r.msgcert.sign !== report.msgcert.sign));
            } else {
                Alert.alert('Error', result);
            }
        } catch (err: any) {
            Alert.alert('Error', err?.message || 'Moderation failed');
        }
    };

    const renderReport = ({ item }: { item: ReportCert }) => {
        const date = new Date(item.msgcert.msg.ts * 1000).toLocaleString([], { dateStyle: 'short', timeStyle: 'short' });


        // Extract content and images
        let rawContent = item.msgcert.msg.content;
        const matchBody = rawContent.match(/<BODY>(.*?)<\/BODY>/s);
        if (matchBody) rawContent = matchBody[1].trim();

        const imgSrcs: string[] = [];
        const imgRegex = /<img[^>]+src="([^">]+)"/g;
        let match;
        while ((match = imgRegex.exec(rawContent)) !== null) {
            imgSrcs.push(match[1]);
        }

        const plainContent = rawContent.replace(/<\/?[^>]+(>|$)/g, '').replace(/&amp;/g, '&').replace(/&lt;/g, '<').replace(/&gt;/g, '>').replace(/&quot;/g, '"').replace(/&#39;/g, "'").trim();

        return (
            <View style={styles.card}>
                <View style={styles.cardHeader}>
                    <Clock size={14} color={C.dark.icon} />
                    <Text style={styles.cardTime}>{date}</Text>
                    <View style={styles.reasonBadge}>
                        <Flag size={12} color="#fff" />
                        <Text style={styles.reasonText}>{item.msgcert.reason || 'No reason'}</Text>
                    </View>
                </View>

                <View style={styles.contentBox}>
                    {plainContent ? (
                        <Text style={styles.contentText}>{plainContent}</Text>
                    ) : null}

                    {imgSrcs.length > 0 && (
                        <View style={styles.imageGallery}>
                            {imgSrcs.map((src, i) => (
                                <Image
                                    key={i}
                                    source={src}
                                    style={styles.queueImage}
                                    contentFit="cover"
                                />
                            ))}
                        </View>
                    )}
                </View>

                <View style={styles.cardFooter}>
                    <Text style={styles.pubKey} numberOfLines={1}>
                        From: {item.msgcert.public_key.slice(0, 12)}...
                    </Text>

                    <View style={styles.actions}>
                        <TouchableOpacity
                            style={[styles.actionBtn, styles.rejectBtn]}
                            onPress={() => Alert.alert('Confirm', 'Reject and delete this message?', [
                                { text: 'Cancel', style: 'cancel' },
                                { text: 'Reject', style: 'destructive', onPress: () => handleModerate(item, 'reject') }
                            ])}
                        >
                            <X size={18} color="#fff" />
                        </TouchableOpacity>

                        <TouchableOpacity
                            style={[styles.actionBtn, styles.approveBtn]}
                            onPress={() => handleModerate(item, 'approve')}
                        >
                            <Check size={18} color="#fff" />
                        </TouchableOpacity>
                    </View>
                </View>
            </View>
        );
    };

    return (
        <SafeAreaView style={styles.container} edges={['top', 'left', 'right']}>
            <StatusBar barStyle="light-content" backgroundColor={C.dark.background} />

            {/* Header */}
            <View style={styles.header}>
                <TouchableOpacity onPress={() => router.back()} style={styles.backBtn}>
                    <ChevronLeft size={28} color={C.dark.text} />
                </TouchableOpacity>
                <View style={styles.headerTitleWrap}>
                    <ShieldAlert size={20} color={C.dark.tint} />
                    <Text style={styles.headerTitle}>Message Reports</Text>
                </View>
                <View style={{ width: 40 }} />
            </View>

            {loading ? (
                <View style={styles.centered}>
                    <ActivityIndicator size="large" color={C.dark.tint} />
                    <Text style={styles.loadingText}>Fetching reports from DHT...</Text>
                </View>
            ) : (
                <FlatList
                    data={reports}
                    keyExtractor={(item) => item.msgcert.sign}
                    renderItem={renderReport}
                    contentContainerStyle={styles.listContent}
                    ListEmptyComponent={
                        <View style={styles.emptyContainer}>
                            <AlertTriangle size={48} color={C.dark.border} />
                            <Text style={styles.emptyText}>No pending reports found.</Text>
                        </View>
                    }
                    refreshControl={
                        <RefreshControl
                            refreshing={refreshing}
                            onRefresh={() => fetchReports(true)}
                            tintColor={C.dark.tint}
                        />
                    }
                />
            )}
        </SafeAreaView>
    );
}

const styles = StyleSheet.create({
    container: {
        flex: 1,
        backgroundColor: C.dark.background,
    },
    header: {
        flexDirection: 'row',
        alignItems: 'center',
        justifyContent: 'space-between',
        paddingHorizontal: 16,
        paddingBottom: 16,
        backgroundColor: C.dark.background,
        borderBottomWidth: 1,
        borderBottomColor: C.dark.border,
    },
    backBtn: {
        width: 40,
        height: 40,
        justifyContent: 'center',
    },
    headerTitleWrap: {
        flexDirection: 'row',
        alignItems: 'center',
        gap: 8,
    },
    headerTitle: {
        fontSize: 18,
        fontFamily: Fonts.bold,
        color: C.dark.text,
    },
    listContent: {
        padding: 16,
        paddingBottom: 40,
    },
    card: {
        backgroundColor: C.dark.primary,
        borderRadius: 16,
        padding: 16,
        marginBottom: 16,
        borderWidth: 1,
        borderColor: C.dark.border,
        shadowColor: '#000',
        shadowOffset: { width: 0, height: 4 },
        shadowOpacity: 0.2,
        shadowRadius: 8,
        elevation: 4,
    },
    cardHeader: {
        flexDirection: 'row',
        alignItems: 'center',
        marginBottom: 12,
        gap: 6,
    },
    cardTime: {
        color: C.dark.icon,
        fontSize: 12,
        fontFamily: Fonts.medium,
        flex: 1,
    },
    reasonBadge: {
        flexDirection: 'row',
        alignItems: 'center',
        backgroundColor: C.dark.tint,
        paddingHorizontal: 8,
        paddingVertical: 4,
        borderRadius: 8,
        gap: 4,
    },
    reasonText: {
        color: '#fff',
        fontSize: 11,
        fontFamily: Fonts.bold,
    },
    contentBox: {
        backgroundColor: 'transparent',
        padding: 12,
        borderRadius: 12,
        marginBottom: 16,
    },
    contentText: {
        color: C.dark.text,
        fontSize: 15,
        fontFamily: Fonts.sans,
        lineHeight: 22,
    },
    cardFooter: {
        flexDirection: 'row',
        alignItems: 'center',
        justifyContent: 'space-between',
    },
    pubKey: {
        color: C.dark.icon,
        fontSize: 11,
        fontFamily: Fonts.mono,
    },
    actions: {
        flexDirection: 'row',
        gap: 12,
    },
    actionBtn: {
        width: 44,
        height: 44,
        borderRadius: 22,
        justifyContent: 'center',
        alignItems: 'center',
        elevation: 2,
    },
    approveBtn: {
        backgroundColor: '#10b981',
    },
    rejectBtn: {
        backgroundColor: '#ef4444',
    },
    centered: {
        flex: 1,
        justifyContent: 'center',
        alignItems: 'center',
        gap: 16,
    },
    loadingText: {
        color: C.dark.icon,
        fontSize: 14,
        fontFamily: Fonts.sans,
    },
    emptyContainer: {
        marginTop: 100,
        alignItems: 'center',
        gap: 16,
        opacity: 0.5,
    },
    emptyText: {
        color: C.dark.text,
        fontSize: 16,
        fontFamily: Fonts.medium,
    },
    imageGallery: {
        flexDirection: 'row',
        flexWrap: 'wrap',
        gap: 8,
        marginTop: 10,
    },
    queueImage: {
        width: 100,
        height: 100,
        borderRadius: 8,
        backgroundColor: '#1a2235',
        borderWidth: 1,
        borderColor: C.dark.border,
    },
});
