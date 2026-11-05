import React, { useState } from 'react';
import { View, Text, TextInput, TouchableOpacity, StyleSheet, KeyboardAvoidingView, Platform, ScrollView, Alert } from 'react-native';
import { Stack, useRouter } from 'expo-router';
import { X, Image as ImageIcon, Send, XCircle } from 'lucide-react-native';
import LibrCore, { RetMsgCert, SendResult } from '@/modules/LibrCore';
import { Colors, Fonts, getAppColors } from '@/constants/theme';
import { useAppStore } from '@/store/useAppStore';
import * as ImagePicker from 'expo-image-picker';
import * as ImageManipulator from 'expo-image-manipulator';
import { Image } from 'expo-image';
import PasteInput from '@tbvjaos510/react-native-paste-input';
import RichTextInput from '../components/RichTextInput';

// Use custom native input on Android to bypass Samsung Keyboard limitations
const AdaptiveInput = Platform.OS === 'android' ? RichTextInput : PasteInput;

export default function CreateMessageModal() {
  const router = useRouter();
  const { state, addMessage, removeMessage, showNotice } = useAppStore();
  const colors = getAppColors(state.isIncognito);
  const [title, setTitle] = useState('');
  const [content, setContent] = useState('');
  const [pendingImages, setPendingImages] = useState<string[]>([]); // base64 strings
  const [isProcessing, setIsProcessing] = useState(false);

  const processFile = async (file: { uri: string; fileSize?: number; base64?: string | null }) => {
    const { uri, fileSize, base64 } = file;

    // Enforce 1MB limit for all image types
    if (fileSize && fileSize > 1024 * 1024) {
      Alert.alert('Too Large', 'Images must be smaller than 1 MB to ensure reliable delivery.');
      return;
    }

    // Check if it's a GIF
    const isGif = uri.toLowerCase().endsWith('.gif');

    if (isGif) {
      if (base64) {
        setPendingImages((prev: string[]) => [...prev, `data:image/gif;base64,${base64}`]);
      } else {
        // If no base64 (common in paste), we might need to fetch it or use manipulation to get it
        // But manipulator strips animation. So we use fetch for GIFs if base64 is missing.
        try {
          const response = await fetch(uri);
          const blob = await response.blob();
          const reader = new FileReader();
          reader.onloadend = () => {
            setPendingImages((prev: string[]) => [...prev, reader.result as string]);
          };
          reader.readAsDataURL(blob);
        } catch (e) {
          console.error('Failed to convert pasted GIF to base64', e);
        }
      }
    } else {
      // 1. Resize and compress like desktop (max 800px)
      const manipResult = await ImageManipulator.manipulateAsync(
        uri,
        [{ resize: { width: 800 } }],
        { compress: 0.8, format: ImageManipulator.SaveFormat.JPEG, base64: true }
      );

      if (manipResult.base64) {
        setPendingImages((prev: string[]) => [...prev, `data:image/jpeg;base64,${manipResult.base64}`]);
      }
    }
  };

  const handlePickImage = async () => {
    try {
      const result = await ImagePicker.launchImageLibraryAsync({
        mediaTypes: ['images'],
        allowsEditing: false,
        quality: 1,
        base64: true,
      });

      if (!result.canceled && result.assets && result.assets.length > 0) {
        setIsProcessing(true);
        await processFile(result.assets[0]);
      }
    } catch (e) {
      Alert.alert('Error', 'Failed to pick image');
      console.error(e);
    } finally {
      setIsProcessing(false);
    }
  };

  const handlePaste = async (error: string | null, files: any[]) => {
    if (error) {
      console.error('Paste error:', error);
      return;
    }
    if (files && files.length > 0) {
      setIsProcessing(true);
      for (const file of files) {
        await processFile(file);
      }
      setIsProcessing(false);
    }
  };

  const handleMediaInserted = async (event: any) => {
    const { uri, fileSize } = event.nativeEvent;
    setIsProcessing(true);
    await processFile({ uri, fileSize, base64: null });
    setIsProcessing(false);
  };

  const removeImage = (index: number) => {
    setPendingImages((prev: string[]) => prev.filter((_: string, i: number) => i !== index));
  };

  const handleSend = async () => {
    const trimmedTitle = title.trim();
    const trimmedBody = content.trim();
    if (!trimmedBody && pendingImages.length === 0) return;

    const imagesHtml = pendingImages.map((src: string) => `<img src="${src}" />`).join('');
    const bodyWithImages = `<BODY>${trimmedBody}${imagesHtml}</BODY>`;

    const fullContent = trimmedTitle ? `<HEAD>${trimmedTitle}</HEAD>${bodyWithImages}` : bodyWithImages;

    // 1. Optimistic UI: Create a temporary certificate to show immediately
    const tempCert: RetMsgCert = {
      public_key: state.publicKey,
      sign: `temp-${Date.now()}-${Math.random().toString(36).substr(2, 9)}`,
      msg: {
        content: fullContent,
        ts: Math.floor(Date.now() / 1000),
      },
      mod_certs: [],
      deleted: '0'
    };

    // 2. Add message to feed local state
    addMessage(tempCert);

    // 3. Dismiss modal immediately
    if (router.canGoBack()) {
      router.back();
    } else {
      router.replace('/');
    }

    // 4. Background: Perform actual deliverya
    try {
      const raw: string = await LibrCore.sendTextMessage(fullContent);
      const result: SendResult = JSON.parse(raw);

      if (result.status === 'sent' || result.status?.startsWith('sent:') || result.status === 'pending_manual') {
        if (result.status === 'pending_manual') {
          showNotice('Under Review', 'Your message contains an image and has been submitted to moderators for review.', 'clock');
        }

        // 5. Replace temp with real data once confirmed
        removeMessage(tempCert.sign);
        const realCert: RetMsgCert = {
          public_key: state.publicKey,
          sign: result.sign,
          msg: {
            content: fullContent,
            ts: result.ts,
          },
          mod_certs: result.mod_certs || [],
          deleted: '0'
        };
        addMessage(realCert);
      } else {
        Alert.alert('Send Failed', result.status || 'Unknown error');
      }
    } catch (e) {
      console.error('[BackgroundSend]', e);
      // Optional: notify of failure or keep as "permanent optimistic error state"?
      // I'll just log for now so it doesn't crash.
    }
  };

  return (
    <KeyboardAvoidingView
      style={[styles.container, { backgroundColor: colors.background }]}
      behavior={Platform.OS === 'ios' ? 'padding' : undefined}
    >
      <Stack.Screen
        options={{
          headerShown: true,
          title: 'Spill Some Gossip',
          headerStyle: { backgroundColor: colors.background },
          headerTintColor: colors.text,
          headerLeft: () => (
            <TouchableOpacity
              onPress={() => router.canGoBack() ? router.back() : router.replace('/')}
              style={{ padding: 8, marginLeft: -8 }}
            >
              <X size={24} color={colors.icon} />
            </TouchableOpacity>
          )
        }}
      />

      <View style={[styles.content, { backgroundColor: colors.background }]}>
        <TextInput
          style={[styles.inputTitle, { color: colors.text, backgroundColor: colors.primary, borderColor: colors.border }]}
          placeholder="Title (optional)"
          placeholderTextColor={colors.icon}
          value={title}
          onChangeText={setTitle}
        />

        <View style={styles.divider} />

        <AdaptiveInput
          style={[styles.inputBody, { color: colors.text, backgroundColor: colors.primary, borderColor: colors.border }]}
          placeholder="What's the gossip?"
          placeholderTextColor={colors.icon}
          multiline
          textAlignVertical="top"
          value={content}
          onChangeText={setContent}
          // @ts-ignore - Specific library props
          onPaste={handlePaste}
          onMediaInserted={handleMediaInserted}
          autoFocus={!isProcessing}
        />

        {pendingImages.length > 0 && (
          <ScrollView horizontal showsHorizontalScrollIndicator={false} style={styles.imageSelector}>
            {pendingImages.map((src: string, index: number) => (
              <View key={index} style={styles.imageThumbContainer}>
                <Image source={src} style={styles.imageThumb} />
                <TouchableOpacity
                  style={styles.removeBtn}
                  onPress={() => removeImage(index)}
                >
                  <XCircle size={20} color={'#ef4444'} fill={colors.background} />
                </TouchableOpacity>
              </View>
            ))}
          </ScrollView>
        )}
      </View>

      <View style={[styles.footer, { borderTopColor: colors.border, backgroundColor: colors.background }]}>
        <TouchableOpacity
          style={styles.attachBtn}
          onPress={handlePickImage}
          disabled={isProcessing}
        >
          {isProcessing ? (
            <View style={{ width: 24, height: 24, borderRadius: 12, borderTopColor: colors.tint, borderWidth: 2 }} />
          ) : (
            <ImageIcon size={24} color={colors.icon} />
          )}
        </TouchableOpacity>

        <TouchableOpacity
          style={[
            styles.sendBtn,
            { backgroundColor: colors.tint, shadowColor: colors.tint },
            (!content.trim() && pendingImages.length === 0) && styles.sendBtnDisabled
          ]}
          disabled={!content.trim() && pendingImages.length === 0}
          onPress={handleSend}
        >
          <Send size={20} color="#fff" />
        </TouchableOpacity>
      </View>
    </KeyboardAvoidingView>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: Colors.dark.background,
  },
  content: {
    flex: 1,
    padding: 16,
  },
  inputTitle: {
    fontSize: 20,
    fontFamily: Fonts.bold,
    color: Colors.dark.text,
    backgroundColor: Colors.dark.primary,
    borderWidth: 1,
    borderColor: Colors.dark.border,
    borderRadius: 14,
    paddingHorizontal: 14,
    paddingVertical: 12,
  },
  divider: {
    height: 12,
  },
  inputBody: {
    flex: 1,
    fontSize: 16,
    fontFamily: Fonts.sans,
    color: Colors.dark.text,
    lineHeight: 24,
    backgroundColor: Colors.dark.primary,
    borderWidth: 1,
    borderColor: Colors.dark.border,
    borderRadius: 14,
    paddingHorizontal: 14,
    paddingVertical: 14,
  },
  footer: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'space-between',
    paddingHorizontal: 20,
    paddingVertical: 12,
    paddingBottom: Platform.OS === 'ios' ? 34 : 12,
    borderTopWidth: 1,
    borderTopColor: Colors.dark.border,
    backgroundColor: Colors.dark.background,
  },
  attachBtn: {
    padding: 8,
  },
  sendBtn: {
    backgroundColor: Colors.dark.tint,
    width: 48,
    height: 48,
    borderRadius: 24,
    alignItems: 'center',
    justifyContent: 'center',
    shadowColor: Colors.dark.tint,
    shadowOffset: { width: 0, height: 4 },
    shadowOpacity: 0.3,
    shadowRadius: 6,
    elevation: 4,
  },
  sendBtnDisabled: {
    opacity: 0.4,
    shadowOpacity: 0,
  },
  imageSelector: {
    maxHeight: 120,
    marginTop: 10,
  },
  imageThumbContainer: {
    marginRight: 10,
    position: 'relative',
    paddingVertical: 10,
  },
  imageThumb: {
    width: 100,
    height: 100,
    borderRadius: 12,
    borderWidth: 1,
    borderColor: Colors.dark.border,
  },
  removeBtn: {
    position: 'absolute',
    top: 0,
    right: -5,
    zIndex: 1,
  }
});
