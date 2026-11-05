import React, { useEffect, useState } from 'react';
import { View, Text, StyleSheet, TouchableOpacity, Alert, StatusBar, Animated, Easing } from 'react-native';
import { useRef } from 'react';
import { SafeAreaView, useSafeAreaInsets } from 'react-native-safe-area-context';
import { Image } from 'expo-image';
import { Colors, Fonts, getAppColors } from '@/constants/theme';
import { useAppStore } from '@/store/useAppStore';
import LibrCore from '@/modules/LibrCore';
import { Shield, RefreshCcw, ArrowLeft, Ghost } from 'lucide-react-native';
import { useRouter } from 'expo-router';

function getAvatarUri(svg: string | null): string | null {
    if (!svg) return null;
    if (svg.startsWith('<svg')) {
        return `data:image/svg+xml;utf8,${encodeURIComponent(svg)}`;
    }
    if (!svg.startsWith('data:')) {
        return `data:image/svg+xml;base64,${svg}`;
    }
    return svg;
}

export default function ProfileScreen() {
    const insets = useSafeAreaInsets();
    const { state, setPublicKey, setIncognito } = useAppStore();
    const colors = getAppColors(state.isIncognito);
    const router = useRouter();
    const [loading, setLoading] = useState(false);
    const [incognitoLoading, setIncognitoLoading] = useState(false);
    const rotateAnim = useRef(new Animated.Value(0)).current;

    useEffect(() => {
        if (loading) {
            Animated.loop(
                Animated.timing(rotateAnim, {
                    toValue: 1,
                    duration: 1000,
                    easing: Easing.linear,
                    useNativeDriver: true,
                })
            ).start();
        } else {
            rotateAnim.setValue(0);
        }
    }, [loading]);

    const spin = rotateAnim.interpolate({
        inputRange: [0, 1],
        outputRange: ['0deg', '360deg'],
    });

    // Avatar and Alias generation
    const [alias, setAlias] = useState('…');
    const [avatarSvg, setAvatarSvg] = useState<string | null>(null);

    // Determine moderator status
    const [isModerator, setIsModerator] = useState(false);

    useEffect(() => {
        if (state.publicKey) {
            (async () => {
                try {
                    const a = await LibrCore.generateAlias(state.publicKey);
                    setAlias(a);
                    const av = await LibrCore.generateAvatar(state.publicKey);
                    setAvatarSvg(av);

                    const mod = await LibrCore.amIMod();
                    setIsModerator(mod);
                } catch { }
            })();
        }
    }, [state.publicKey]);

    useEffect(() => {
        LibrCore.isIncognitoEnabled()
            .then(setIncognito)
            .catch(() => setIncognito(false));
    }, [setIncognito]);

    const handleResetIdentity = () => {
        Alert.alert(
            'Reset Identity',
            'This will generate new cryptographic keys. Your alias and avatar will change, and your old identity will be lost. Continue?',
            [
                { text: 'Cancel', style: 'cancel' },
                {
                    text: 'Reset',
                    style: 'destructive',
                    onPress: async () => {
                        setLoading(true);
                        try {
                            const newKey = await LibrCore.regenKeys();
                            if (newKey && !newKey.startsWith('error:')) {
                                setPublicKey(newKey);
                                Alert.alert('Success', 'Identity has been reset.');
                            } else {
                                Alert.alert('Error', newKey || 'Failed to regenerate keys');
                            }
                        } catch (err: any) {
                            Alert.alert('Error', err?.message || 'Failed to reset identity');
                        } finally {
                            setLoading(false);
                        }
                    }
                }
            ]
        );
    };

    const handleIncognitoToggle = async () => {
        setIncognitoLoading(true);
        try {
            const nextKey = state.isIncognito
                ? await LibrCore.disableIncognito()
                : await LibrCore.enableIncognito();

            if (!nextKey) {
                Alert.alert('Incognito Unavailable', 'Incognito needs a rebuilt native app before it can switch identities on-device.');
                return;
            }

            if (nextKey.startsWith('error:')) {
                Alert.alert('Error', nextKey);
                return;
            }

            setPublicKey(nextKey);
            setIncognito(!state.isIncognito);
            setIsModerator(false);
        } catch (err: any) {
            Alert.alert('Error', err?.message || 'Failed to update incognito mode');
        } finally {
            setIncognitoLoading(false);
        }
    };

    return (
        <SafeAreaView style={[styles.container, { backgroundColor: colors.background }]} edges={['left', 'right', 'bottom']}>
            <StatusBar barStyle="light-content" backgroundColor={colors.background} />
            {/* Header */}
            <View style={[styles.header, { paddingTop: Math.max(insets.top, 10), borderBottomColor: colors.border }]}>
                <TouchableOpacity onPress={() => router.back()} style={styles.backButton}>
                    <ArrowLeft size={24} color={colors.text} />
                </TouchableOpacity>
                <Text style={[styles.headerTitle, { color: colors.text }]}>Profile</Text>
                <View style={styles.statusContainer}>
                    <View style={[styles.statusDot, { backgroundColor: state.connectionStatus === 'connected' ? '#22c55e' : '#eab308' }]} />
                    <Text style={[styles.statusText, { color: state.connectionStatus === 'connected' ? '#22c55e' : '#eab308' }]}>
                        {state.connectionStatus === 'connected' ? 'Connected' : 'Connecting...'}
                    </Text>
                </View>
            </View>

            <View style={styles.content}>
                <View style={[styles.card, { backgroundColor: colors.primary, borderColor: colors.border }]}>
                    {/* Avatar */}
                    <View style={[styles.avatarContainer, { backgroundColor: colors.border, borderColor: colors.background }]}>
                        {avatarSvg ? (
                            <Image
                                source={getAvatarUri(avatarSvg)}
                                style={styles.avatar}
                                contentFit="cover"
                            />
                        ) : (
                            <View style={styles.avatarFallback} />
                        )}
                    </View>

                    {/* Alias */}
                    <Text style={[styles.aliasText, { color: colors.text }]}>{alias}</Text>

                    {/* Badge */}
                    {isModerator && (
                        <View style={styles.badge}>
                            <Shield size={14} color={colors.background} fill={colors.background} />
                            <Text style={styles.badgeText}>Moderator</Text>
                        </View>
                    )}

                    <View style={styles.spacer} />

                    <TouchableOpacity
                        style={[
                            styles.incognitoBtn,
                            { borderColor: colors.border },
                            state.isIncognito && styles.incognitoBtnActive,
                            incognitoLoading && styles.buttonDisabled
                        ]}
                        onPress={handleIncognitoToggle}
                        disabled={incognitoLoading || loading}
                        activeOpacity={0.85}
                    >
                        <Ghost size={18} color={state.isIncognito ? colors.background : colors.text} />
                        <View style={styles.actionTextWrap}>
                            <Text style={[styles.incognitoBtnText, { color: state.isIncognito ? colors.background : colors.text }, state.isIncognito && styles.incognitoBtnTextActive]}>
                                {incognitoLoading ? 'Switching...' : state.isIncognito ? 'Incognito On' : 'Go Incognito'}
                            </Text>
                            <Text style={[styles.incognitoHint, { color: state.isIncognito ? 'rgba(15,10,24,0.72)' : colors.muted }, state.isIncognito && styles.incognitoHintActive]}>
                                {state.isIncognito ? 'Using a temporary identity until you switch back.' : 'Temporarily swap to a stealth identity.'}
                            </Text>
                        </View>
                    </TouchableOpacity>

                    {/* Actions */}
                    <TouchableOpacity
                        style={[styles.resetBtn, loading && styles.buttonDisabled]}
                        onPress={handleResetIdentity}
                        disabled={loading || incognitoLoading}
                        activeOpacity={0.8}
                    >
                        <Animated.View style={{ transform: [{ rotate: spin }] }}>
                            <RefreshCcw size={20} color="#fff" />
                        </Animated.View>
                        <Text style={styles.resetBtnText}>{loading ? 'Resetting...' : 'Reset Identity'}</Text>
                    </TouchableOpacity>
                </View>
            </View>
        </SafeAreaView>
    );
}

const styles = StyleSheet.create({
    container: {
        flex: 1,
        backgroundColor: Colors.dark.background,
    },
    header: {
        flexDirection: 'row',
        alignItems: 'center',
        justifyContent: 'space-between',
        paddingHorizontal: 16,
        paddingBottom: 16,
        borderBottomWidth: 1,
        borderBottomColor: Colors.dark.border,
    },
    headerTitle: {
        fontSize: 20,
        color: Colors.dark.text,
        fontFamily: Fonts.display,
    },
    backButton: {
        padding: 8,
    },
    statusContainer: {
        flexDirection: 'row',
        alignItems: 'center',
        gap: 6,
    },
    statusDot: {
        width: 8,
        height: 8,
        borderRadius: 4,
    },
    statusText: {
        fontSize: 12,
        fontFamily: Fonts.medium,
    },
    content: {
        flex: 1,
        padding: 24,
        alignItems: 'center',
        justifyContent: 'center',
    },
    card: {
        width: '100%',
        backgroundColor: Colors.dark.primary,
        borderRadius: 24,
        padding: 32,
        alignItems: 'center',
        borderWidth: 1,
        borderColor: Colors.dark.border,
        shadowColor: '#000',
        shadowOffset: { width: 0, height: 8 },
        shadowOpacity: 0.3,
        shadowRadius: 12,
        elevation: 8,
    },
    avatarContainer: {
        width: 120,
        height: 120,
        borderRadius: 60,
        backgroundColor: Colors.dark.border,
        overflow: 'hidden',
        marginBottom: 24,
        borderWidth: 4,
        borderColor: Colors.dark.background,
    },
    avatar: {
        width: '100%',
        height: '100%',
    },
    avatarFallback: {
        flex: 1,
        backgroundColor: Colors.dark.icon,
    },
    aliasText: {
        fontSize: 24,
        color: Colors.dark.text,
        marginBottom: 12,
        textAlign: 'center',
        fontFamily: Fonts.display,
    },
    badge: {
        flexDirection: 'row',
        alignItems: 'center',
        backgroundColor: '#22c55e', // Vibrant green
        paddingHorizontal: 12,
        paddingVertical: 6,
        borderRadius: 16,
        gap: 6,
    },
    badgeText: {
        color: Colors.dark.background,
        fontSize: 13,
        textTransform: 'uppercase',
        fontFamily: Fonts.medium,
    },
    spacer: {
        height: 24,
    },
    incognitoBtn: {
        flexDirection: 'row',
        alignItems: 'center',
        backgroundColor: 'rgba(255,255,255,0.03)',
        borderWidth: 1,
        borderColor: Colors.dark.border,
        paddingHorizontal: 18,
        paddingVertical: 14,
        borderRadius: 16,
        width: '100%',
        justifyContent: 'flex-start',
        marginBottom: 12,
        gap: 10,
    },
    incognitoBtnActive: {
        backgroundColor: '#b49cff',
        borderColor: '#b49cff',
    },
    actionTextWrap: {
        flex: 1,
    },
    incognitoBtnText: {
        color: Colors.dark.text,
        fontSize: 16,
        fontFamily: Fonts.medium,
        marginBottom: 2,
    },
    incognitoBtnTextActive: {
        color: Colors.dark.background,
    },
    incognitoHint: {
        color: Colors.dark.muted,
        fontSize: 12,
        fontFamily: Fonts.sans,
    },
    incognitoHintActive: {
        color: 'rgba(10,15,28,0.72)',
    },
    resetBtn: {
        flexDirection: 'row',
        alignItems: 'center',
        backgroundColor: Colors.dark.red,
        paddingHorizontal: 20,
        paddingVertical: 14,
        borderRadius: 16,
        width: '100%',
        justifyContent: 'center',
        gap: 8,
        shadowColor: Colors.dark.red,
        shadowOffset: { width: 0, height: 4 },
        shadowOpacity: 0.24,
        shadowRadius: 8,
        elevation: 4,
    },
    buttonDisabled: {
        opacity: 0.7,
    },
    resetBtnText: {
        color: '#fff',
        fontSize: 16,
        fontFamily: Fonts.medium,
    }
});
