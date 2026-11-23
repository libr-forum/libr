import React from 'react';
import { View, Text, StyleSheet, TouchableOpacity, StatusBar } from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';
import { Colors as C } from '@/constants/theme';
import { useRouter } from 'expo-router';
import { ChevronLeft, Shield } from 'lucide-react-native';

export default function ModerationScreen() {
    const router = useRouter();

    return (
        <SafeAreaView style={styles.container} edges={['top', 'left', 'right']}>
            <StatusBar barStyle="light-content" backgroundColor={C.dark.background} />

            {/* Header */}
            <View style={styles.header}>
                <TouchableOpacity onPress={() => router.back()} style={styles.backBtn}>
                    <ChevronLeft size={28} color={C.dark.text} />
                </TouchableOpacity>
                <View style={styles.headerTitleWrap}>
                    <Shield size={20} color={C.dark.tint} />
                    <Text style={styles.headerTitle}>Moderation Logs</Text>
                </View>
                <View style={{ width: 40 }} />
            </View>

            <View style={styles.centered}>
                <Shield size={64} color={C.dark.border} style={{ marginBottom: 20 }} />
                <Text style={styles.title}>Logs coming soon</Text>
                <Text style={styles.subtitle}>Detailed moderation logs and decision history will be viewable here in a future update.</Text>
            </View>
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
        fontWeight: '800',
        color: C.dark.text,
    },
    centered: {
        flex: 1,
        justifyContent: 'center',
        alignItems: 'center',
        padding: 40,
        textAlign: 'center',
    },
    title: {
        color: C.dark.text,
        fontSize: 20,
        fontWeight: '700',
        marginBottom: 10,
    },
    subtitle: {
        color: C.dark.icon,
        fontSize: 14,
        textAlign: 'center',
        lineHeight: 20,
    },
});
