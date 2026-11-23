import React, { useEffect, useRef } from 'react';
import { Animated, StyleSheet, TouchableOpacity, View } from 'react-native';

interface AnimatedHamburgerProps {
    isOpen: boolean;
    onPress: () => void;
    color?: string;
}

export function AnimatedHamburger({ isOpen, onPress, color = '#ededed' }: AnimatedHamburgerProps) {
    const animation = useRef(new Animated.Value(0)).current;

    useEffect(() => {
        Animated.timing(animation, {
            toValue: isOpen ? 1 : 0,
            duration: 300,
            useNativeDriver: true,
        }).start();
    }, [isOpen]);

    const topBarRotate = animation.interpolate({
        inputRange: [0, 1],
        outputRange: ['0deg', '45deg'],
    });

    const topBarY = animation.interpolate({
        inputRange: [0, 1],
        outputRange: [0, 8],
    });

    const middleBarOpacity = animation.interpolate({
        inputRange: [0, 0.5, 1],
        outputRange: [1, 0, 0],
    });

    const bottomBarRotate = animation.interpolate({
        inputRange: [0, 1],
        outputRange: ['0deg', '-45deg'],
    });

    const bottomBarY = animation.interpolate({
        inputRange: [0, 1],
        outputRange: [0, -8],
    });

    return (
        <TouchableOpacity onPress={onPress} activeOpacity={0.7} style={styles.container}>
            <Animated.View
                style={[
                    styles.bar,
                    { backgroundColor: color },
                    { transform: [{ translateY: topBarY }, { rotate: topBarRotate }] },
                ]}
            />
            <Animated.View
                style={[
                    styles.bar,
                    { backgroundColor: color, marginVertical: 5 },
                    { opacity: middleBarOpacity },
                ]}
            />
            <Animated.View
                style={[
                    styles.bar,
                    { backgroundColor: color },
                    { transform: [{ translateY: bottomBarY }, { rotate: bottomBarRotate }] },
                ]}
            />
        </TouchableOpacity>
    );
}

const styles = StyleSheet.create({
    container: {
        width: 30,
        height: 30,
        justifyContent: 'center',
        alignItems: 'center',
    },
    bar: {
        width: 20,
        height: 2,
        borderRadius: 1,
    },
});
