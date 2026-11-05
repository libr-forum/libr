import React, { useEffect } from 'react';
import { View, Text, StyleSheet, Animated, Easing } from 'react-native';
import { SvgXml } from 'react-native-svg';
import { Colors, Fonts } from '@/constants/theme';

const LIBR_LOGO_SVG = `<svg id="Layer_1" data-name="Layer 1" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 1000 1000"><path d="M572.9,63.28c13-7.31,26.55-12.91,41.24-15.66,5.72-1.13,11.52-2.12,17.33-1.8,7.17.4,11.45,4.14,11.24,11.85a13.2,13.2,0,0,1-.21,3.89c-8.42,24.79-9.91,51.6-22.67,75.2-5.44,10.05-12.23,20.33-8.73,33,.83,7.94-5.21,10.86-10.49,14.35a60.61,60.61,0,0,0-21.82,22.27c-17.65,21.75-35.95,42.93-51.74,66.18C496.86,317,471.54,363.9,452.78,414.3c-3.16,8.49-5.46,17.44-11.18,24.82-4.2,2.85-5,7.58-6,11.9-2.42,10.56-8.56,18.89-15.21,26.94a27.59,27.59,0,0,1-17.21,10.19c-11.87,2-20.7,10.68-31.08,16-10.9,5.54-19.61,5.19-28.46-2.65-13-19.2-20.8-40.79-24.94-63.15-12.6-67.94-1.47-132,34.78-191.26,4-6.48,4.95-14.67,11.85-19.74-6.69,19.05-8.79,38.75-7.55,58.52,2.13,34.18,6.59,67.93,27.13,97.1,2.07,2.94,3.9,7.4,7.83,5.71,4.17-1.79,1.43-6.11.86-9.48-.88-5.09-3.43-9.91-2.62-15.29,6.35.16,12,2.9,18.89,2.55,12.88-.67,23.26-4.36,32-13.8,3.29-3.54,8.11-5.45,12.53-7.61A47.77,47.77,0,0,0,477,323.13c6.56-12.56,17.61-22.4,25.2-34.75,10.73-17.43,22.34-34.26,35.9-49.93,9.39-10.85,17.27-23.52,13.91-40.93-3.5-18.14-.15-36.24,11.6-52.18,8-10.81,12.63-23.58,15.48-36.81.95-4.4,3.71-7.53,6.42-10.82,5.83-7.07,11.65-14,13.38-23.59s1-10.73-9.19-9.47C584,65.36,578.15,68,572.9,63.28Z" style="fill:#90b1ce"/><path d="M528.62,485.8c-17.64,9.58-37.15,13.8-56.1,19.69-3.24,1-8.28.06-9,5.54,4.75,2.83,9.47-.11,14.2-.15,8.77-.08,12.56,4.24,10.56,12.83a24.27,24.27,0,0,1-3.1,8.17c-13.84,20.5-22.32,43.89-35.15,64.92-2.21,3.63-3.77,7.55-5.69,11.3-2,4-4.81,6.27-9.23,2.8-1.49-2.6-1.4-5.37-.92-8.19,2.16-11.4.71-22.91.94-34.37A305.89,305.89,0,0,1,443.46,499c6.75-26.25,12.8-52.66,21.44-78.45,15.75-47,35.57-92,60.73-134.72,4-6.76,7.68-13.69,11.81-20.35,12.45-18.65,24.82-37.37,40.34-53.72,9.91-13.7,24-23.55,34-37.22,2.1-2.68,3.81-5.75,7.06-7.34,5.44-2.49,10.93-8.68,16.93-3.9s1.9,11.7.29,17.44c-5.45,19.49-13.15,38.3-17.09,58.29-2.1,10.63-9.94,16.46-20.37,18.82-2.43.55-5.26-.09-7.13,2.25-3.91,1.26-8.11,2-11.62,4-1.94,1.1.93,1.13,1.52,1.72,11.8,4.57,15.46,10.85,13.41,23.32-1.11,6.74-2.07,13.72-5.77,19.53-9.51,14.91-13.13,31.74-16.35,48.65-2.11,11.1-7.19,20.71-12.84,30.19l-54.37,26.81c.23.66.47,1.32.71,2l20.37-5.05c3.41,1.72,6.24-.75,9.33-1.32,11.3-2.08,18.62,4.46,15.7,15.43-3.94,14.78-8.76,29.31-10.81,44.53C539.69,477.49,531.15,479.34,528.62,485.8Z" style="fill:#8ba7c6"/></svg>`;

export function ConnectingScreen({ status, logLines }: { status: string; logLines: string[] }) {
    const pulseAnim = new Animated.Value(1);

    useEffect(() => {
        Animated.loop(
            Animated.sequence([
                Animated.timing(pulseAnim, {
                    toValue: 1.1,
                    duration: 1000,
                    easing: Easing.inOut(Easing.ease),
                    useNativeDriver: true,
                }),
                Animated.timing(pulseAnim, {
                    toValue: 1,
                    duration: 1000,
                    easing: Easing.inOut(Easing.ease),
                    useNativeDriver: true,
                }),
            ])
        ).start();
    }, []);

    const latestLog = logLines.length > 0 ? logLines[logLines.length - 1] : 'Initializing...';

    return (
        <View style={styles.container}>
            <Animated.View style={{ transform: [{ scale: pulseAnim }], marginBottom: 40 }}>
                <SvgXml xml={LIBR_LOGO_SVG} width={120} height={120} />
            </Animated.View>

            <Text style={styles.title}>libr</Text>
            <Text style={styles.statusText}>{status}</Text>

            <View style={styles.logContainer}>
                <Text style={styles.logText} numberOfLines={1}>{latestLog}</Text>
            </View>
        </View>
    );
}

const styles = StyleSheet.create({
    container: {
        ...StyleSheet.absoluteFillObject,
        backgroundColor: Colors.dark.background,
        alignItems: 'center',
        justifyContent: 'center',
        zIndex: 1000,
    },
    title: {
        fontSize: 48,
        color: Colors.dark.text,
        fontFamily: Fonts.display,
        marginBottom: 8,
    },
    statusText: {
        fontSize: 16,
        color: Colors.dark.tint,
        fontFamily: Fonts.bold,
        textTransform: 'uppercase',
        letterSpacing: 2,
        marginBottom: 40,
    },
    logContainer: {
        position: 'absolute',
        bottom: 60,
        width: '80%',
        padding: 12,
        backgroundColor: 'rgba(255,255,255,0.05)',
        borderRadius: 8,
        borderWidth: 1,
        borderColor: 'rgba(255,255,255,0.1)',
    },
    logText: {
        color: Colors.dark.icon,
        fontSize: 12,
        fontFamily: Fonts.mono,
        textAlign: 'center',
    },
});
