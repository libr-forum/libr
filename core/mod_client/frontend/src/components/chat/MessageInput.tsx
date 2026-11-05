import React, {
  forwardRef,
  useRef,
  useImperativeHandle,
  useState,
  useMemo,
  useEffect,
  useCallback
} from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import { Send, X, ImageIcon, XCircle } from 'lucide-react';
import { cn } from '../../lib/utils';
import { useAppStore } from '../../store/useAppStore';
import { apiService } from '../../services/api';
import { EditorContent, useEditor } from '@tiptap/react';
import StarterKit from '@tiptap/starter-kit';
import Strike from '@tiptap/extension-strike';
import CodeBlock from '@tiptap/extension-code-block';
import Placeholder from '@tiptap/extension-placeholder';
import BulletList from '@tiptap/extension-bullet-list';
import ListItem from '@tiptap/extension-list-item';
import { toast } from 'sonner';

import { logger } from '../../logger/logger';

interface MessageInputProps {
  onClose?: () => void;
}

const titles = [
  'Create Post',
  'New Post',
  'Share Some Tea',
  'Spill Some Gossip',
];

export const MessageInput = forwardRef<HTMLDivElement, MessageInputProps>(
  ({ onClose }, ref) => {
    const [title, setTitle] = useState('');
    const [bodyText, setBodyText] = useState('');
    const [isSending, setIsSending] = useState(false);
    const [shake, setShake] = useState(false);
    const [isDragOver, setIsDragOver] = useState(false);
    // Track pending image dataURLs so user can preview & remove before sending
    const [pendingImages, setPendingImages] = useState<string[]>([]);

    const containerRef = useRef<HTMLDivElement>(null);
    const fileInputRef = useRef<HTMLInputElement>(null);
    const { currentCommunity, addMessage } = useAppStore();

    useEffect(() => {
      logger.info('[MessageInput] Mounted');
      return () => {
        logger.info('[MessageInput] Unmounted');
      };
    }, []);

    const CustomStrike = Strike.extend({
      addKeyboardShortcuts() {
        return {
          'Mod-Shift-x': () => this.editor.commands.toggleStrike(),
        };
      },
    });

    const CustomCodeBlock = CodeBlock.extend({
      addKeyboardShortcuts() {
        return {
          'Mod-`': () => this.editor.commands.toggleCodeBlock(),
        };
      },
    });

    useImperativeHandle(ref, () => {
      logger.debug('[MessageInput] Imperative handle set');
      return containerRef.current!;
    });

    const processImageFile = useCallback((file: File) => {
      if (!file.type.startsWith('image/')) return;
      // 1MB guard
      if (file.size > 1024 * 1024) {
        toast.error('Image must be smaller than 1 MB.');
        return;
      }

      const isGif = file.type === 'image/gif';

      const reader = new FileReader();
      reader.onload = (e) => {
        const dataUrlFromReader = e.target?.result as string;

        if (isGif) {
          // Bypass canvas for GIFs to preserve animation
          setPendingImages(prev => [...prev, dataUrlFromReader]);
          return;
        }

        const img = new window.Image();
        img.onload = () => {
          const canvas = document.createElement('canvas');
          let width = img.width;
          let height = img.height;
          const maxDimension = 800;
          if (width > maxDimension || height > maxDimension) {
            if (width > height) {
              height = Math.round((height * maxDimension) / width);
              width = maxDimension;
            } else {
              width = Math.round((width * maxDimension) / height);
              height = maxDimension;
            }
          }
          canvas.width = width;
          canvas.height = height;
          const ctx = canvas.getContext('2d');
          if (ctx) {
            ctx.drawImage(img, 0, 0, width, height);
            const dataUrl = canvas.toDataURL('image/jpeg', 0.8);
            // Show preview, but do NO LONGER insert into editor
            setPendingImages(prev => [...prev, dataUrl]);
          }
        };
        img.src = dataUrlFromReader;
      };
      reader.readAsDataURL(file);
    }, []);

    const editor = useEditor({
      extensions: [
        StarterKit.configure({ strike: false, codeBlock: false, bulletList: false, listItem: false }),
        CustomStrike,
        CustomCodeBlock,
        BulletList,
        ListItem,
        Placeholder.configure({ placeholder: 'Message' }),
      ],
      content: '',
      editorProps: {
        attributes: {
          class:
            'h-full prose-mirror-editor min-h-[27rem] p-3 m-1 w-[99%] bg-muted/30 border border-border/50 rounded-2xl resize-none focus:outline-none focus:ring-2 focus:ring-libr-accent1/50 transition-all duration-200 text-sm',
        },
        handleKeyDown(view, event) {
          if ((event.ctrlKey || event.metaKey) && event.key === 'Enter') {
            logger.info('[MessageInput] Ctrl/Cmd+Enter detected — sending message');
            event.preventDefault();
            handleSend();
            return true;
          }
          return false;
        },
        handleDrop(view, event) {
          const file = event.dataTransfer?.files?.[0];
          if (file && file.type.startsWith('image/')) {
            processImageFile(file);
            return true;
          }
          return false;
        },
        handlePaste(view, event) {
          const items = event.clipboardData?.items;
          if (items) {
            for (const item of items) {
              if (item.type.startsWith('image/')) {
                const file = item.getAsFile();
                if (file) {
                  processImageFile(file);
                  return true;
                }
              }
            }
          }
          return false;
        },
      },
      onUpdate: ({ editor }) => {
        const text = editor.getText();
        setBodyText(text);
        logger.debug('[MessageInput] Editor updated', { textLength: text.length });
        if (shake) setShake(false);
      },
    });

    const handleImageUpload = (event: React.ChangeEvent<HTMLInputElement>) => {
      const file = event.target.files?.[0];
      if (!file) return;
      processImageFile(file);
      if (fileInputRef.current) fileInputRef.current.value = '';
    };

    const handleDrop = useCallback((e: React.DragEvent<HTMLDivElement>) => {
      e.preventDefault();
      setIsDragOver(false);
      const file = e.dataTransfer.files?.[0];
      if (file) processImageFile(file);
    }, [processImageFile]);

    const removePendingImage = (dataUrl: string) => {
      // Remove from preview list
      setPendingImages(prev => prev.filter(u => u !== dataUrl));
    };

    const handleSend = async () => {
      const trimmedText = bodyText.trim();
      const trimmedTitle = title.trim();

      logger.info('[MessageInput] Send attempt', {
        hasText: !!trimmedText,
        hasTitle: !!trimmedTitle,
        communityId: currentCommunity?.id,
        isSending
      });

      if (!trimmedText && trimmedTitle) {
        logger.warn('[MessageInput] Title provided but body empty — shake triggered');
        setShake(true);
        return;
      }

      if ((!trimmedText && pendingImages.length === 0) || !currentCommunity || isSending) {
        logger.warn('[MessageInput] Send aborted — missing content or already sending');
        return;
      }

      setIsSending(true);
      try {
        const contentHtml = editor?.getHTML() ?? '';
        const imagesHtml = pendingImages.map(src => `<img src="${src}" />`).join('');
        const bodyWithImages = `<BODY>${contentHtml}${imagesHtml}</BODY>`;

        const formatted = trimmedTitle
          ? `<HEAD>${trimmedTitle}</HEAD>${bodyWithImages}`
          : bodyWithImages;

        logger.debug('[MessageInput] Sending formatted message', { formatted });

        const newMsg = await apiService.sendMessage(currentCommunity.id, formatted);
        addMessage(newMsg);

        logger.info('[MessageInput] Message sent successfully', { messageContent: newMsg?.content });

        if (/<img\s+[^>]*src="data:image/i.test(formatted) && newMsg.status === 'pending') {
          toast.message('Sent for manual approval', {
            description: 'Your image post is now waiting for moderator review.',
          });
        }

        setTitle('');
        setBodyText('');
        setPendingImages([]);
        editor?.commands.setContent('');

        onClose?.();
      } catch (err) {
        logger.error('[MessageInput] Send failed', err);
      } finally {
        setIsSending(false);
      }
    };

    const randomTitle = useMemo(() => {
      const chosen = titles[Math.floor(Math.random() * titles.length)];
      logger.debug('[MessageInput] Random title chosen', chosen);
      return chosen;
    }, []);

    return (
      <motion.div
        ref={containerRef}
        initial={{ scale: 0.9, opacity: 0 }}
        animate={{ scale: 1, opacity: 1 }}
        exit={{ scale: 0.9, opacity: 0 }}
        transition={{ duration: 0.2 }}
        className="absolute w-[50%] h-[75%] z-50 bg-card border border-border rounded-3xl p-4 shadow-2xl"
        onDragOver={(e) => { e.preventDefault(); setIsDragOver(true); }}
        onDragLeave={() => setIsDragOver(false)}
        onDrop={handleDrop}
      >
        {/* Drag overlay */}
        <AnimatePresence>
          {isDragOver && (
            <motion.div
              initial={{ opacity: 0 }}
              animate={{ opacity: 1 }}
              exit={{ opacity: 0 }}
              className="absolute inset-0 z-50 flex items-center justify-center bg-libr-accent1/20 border-2 border-dashed border-libr-accent1 rounded-3xl pointer-events-none"
            >
              <div className="text-center">
                <ImageIcon className="w-12 h-12 text-libr-accent1 mx-auto mb-2" />
                <p className="text-libr-accent1 font-semibold">Drop image here</p>
              </div>
            </motion.div>
          )}
        </AnimatePresence>
        <style>
          {`
            .prose-mirror-editor ul {
              list-style-type: disc;
              margin-left: 1.5em;
              padding-left: 1.5em;
            }
            .prose-mirror-editor ol {
              list-style-type: decimal;
              margin-left: 1.5em;
              padding-left: 1.5em;
            }
            .prose-mirror-editor li {
              margin-bottom: 0.25em;
            }

          `}
        </style>
        <div className="h-full flex flex-col">
          {/* Image Previews Strip */}
          <AnimatePresence>
            {pendingImages.length > 0 && (
              <motion.div
                initial={{ opacity: 0, height: 0 }}
                animate={{ opacity: 1, height: 'auto' }}
                exit={{ opacity: 0, height: 0 }}
                className="flex flex-wrap gap-2 mb-2 mt-1"
              >
                {pendingImages.map((src, idx) => (
                  <div key={idx} className="relative group">
                    <img
                      src={src}
                      alt={`attachment-${idx}`}
                      className="h-20 w-auto rounded-xl object-cover border border-border/50 shadow"
                    />
                    <button
                      onClick={() => removePendingImage(src)}
                      className="absolute -top-2 -right-2 bg-destructive text-white rounded-full p-0.5 opacity-0 group-hover:opacity-100 transition-opacity shadow"
                      title="Remove image"
                    >
                      <XCircle className="w-4 h-4" />
                    </button>
                    <span className="absolute bottom-1 left-1 text-[10px] bg-black/50 text-white px-1 rounded">
                      {idx + 1}
                    </span>
                  </div>
                ))}
              </motion.div>
            )}
          </AnimatePresence>

          {/* Header */}
          <div className="flex flex-row items-center justify-between mb-4">
            <h2 className="text-lg font-semibold">{randomTitle}</h2>
            <button
              onClick={onClose}
              className="text-muted-foreground hover:text-foreground"
            >
              <X className="p-1 h-[10%] aspect-square bg-muted rounded-full" />
            </button>
          </div>

          {/* Title Input */}
          <textarea
            placeholder="Title (optional)"
            value={title}
            onChange={(e) => {
              logger.debug('[MessageInput] Title changed', e.target.value);
              setTitle(e.target.value);
            }}
            className="w-full mb-3 p-3 text-sm h-[20%] max-h-20 border border-border/50 rounded-2xl bg-muted/30 text-foreground placeholder-muted-foreground focus:outline-none focus:ring-2 focus:ring-libr-accent1/50 resize-none leading-tight"
          />

          {/* Body Editor */}
          <div className="flex-1 overflow-hidden">
            <div className="h-full overflow-y-auto">
              {editor && <EditorContent editor={editor} />}
            </div>
          </div>

          {/* Footer */}
          <div className="flex justify-between items-center mt-4">
            <div className="flex items-center">
              <input
                type="file"
                accept="image/*"
                ref={fileInputRef}
                className="hidden"
                onChange={handleImageUpload}
              />
              <button
                onClick={() => fileInputRef.current?.click()}
                className="p-3 text-muted-foreground hover:bg-muted rounded-2xl transition"
                title="Attach Image"
                disabled={isSending}
              >
                <ImageIcon className="w-5 h-5 pointer-events-none" />
              </button>
            </div>

            <motion.button
              whileHover={{ scale: 1.05 }}
              whileTap={{ scale: 0.95 }}
              onClick={handleSend}
              disabled={isSending || (bodyText.trim() === '' && pendingImages.length === 0)}
              className="p-4 bg-libr-accent1 hover:bg-libr-accent1/80 disabled:bg-muted disabled:cursor-not-allowed rounded-2xl text-white text-sm"
            >
              {isSending ? (
                <div className="w-4 h-4 border-2 border-white border-t-transparent rounded-full animate-spin" />
              ) : (
                <Send className="w-4 h-4" />
              )}
            </motion.button>
          </div>
        </div>
      </motion.div>
    );
  }
);

MessageInput.displayName = 'MessageInput';
