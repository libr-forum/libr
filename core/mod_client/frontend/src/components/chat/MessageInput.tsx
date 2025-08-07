import React, {
  forwardRef,
  useRef,
  useImperativeHandle,
  useState,
  useMemo,
} from 'react';
import { motion } from 'framer-motion';
import { Send, X } from 'lucide-react';
import { cn } from '../../lib/utils';
import { useAppStore } from '../../store/useAppStore';
import { apiService } from '../../services/api';
import { EditorContent, useEditor } from '@tiptap/react';
import StarterKit from '@tiptap/starter-kit';
import Strike from '@tiptap/extension-strike';
import CodeBlock from '@tiptap/extension-code-block';
import Placeholder from '@tiptap/extension-placeholder';

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

    const containerRef = useRef<HTMLDivElement>(null);
    const { currentCommunity, addMessage } = useAppStore();
    

    const CustomStrike = Strike.extend({
      addKeyboardShortcuts() {
        return {
          'Mod-Shift-x': () => this.editor.commands.toggleStrike(),
        };
      },
    });

    const CustomCodeBlock = CodeBlock.extend({
      addKeyboardShortcuts(){
        return{
          'Mod-`':()=>this.editor.commands.toggleCodeBlock(),
        };
      },
    });

    useImperativeHandle(ref, () => containerRef.current!);

    const editor = useEditor({
      extensions: [StarterKit.configure({strike:false,codeBlock:false}),CustomStrike,CustomCodeBlock,Placeholder.configure({placeholder:'Message'}),],
      content: '',
      editorProps: {
        attributes: {
          class:
            'h-full prose-mirror-editor min-h-[27rem] p-3 m-1 w-[99%] bg-muted/30 border border-border/50 rounded-2xl resize-none focus:outline-none focus:ring-2 focus:ring-libr-accent1/50 transition-all duration-200 text-sm',
        },
        handleKeyDown(view, event) {
          // Ctrl+Enter or Cmd+Enter
          if ((event.ctrlKey || event.metaKey) && event.key === 'Enter') {
            event.preventDefault();
            handleSend(); // your message sending function
            return true;
          }
          return false;
        },
      },
      
      onUpdate: ({ editor }) => {
        const text = editor.getText();
        const cleanHTML = text.replace(/(<p>\s*<\/p>)+$/g, '');
        setBodyText(text);
        if (shake) setShake(false);
      },
    });

    const handleSend = async () => {
      const trimmedText = bodyText.trim();
      const trimmedTitle = title.trim();

      if (!trimmedText && trimmedTitle) {
        setShake(true);
        return;
      }

      if (!trimmedText || !currentCommunity || isSending) return;

      setIsSending(true);
      try {
        const formatted = trimmedTitle
          ? `<HEAD>${trimmedTitle}</HEAD><BODY>${editor?.getHTML()}</BODY>`
          : `<BODY>${editor?.getHTML()}</BODY>`;

        const newMsg = await apiService.sendMessage(currentCommunity.id, formatted);
        addMessage(newMsg);

        setTitle('');
        setBodyText('');
        editor?.commands.setContent('');

        // ✅ Only auto-close if message is not approved
        if (newMsg.status !== 'approved') {
          onClose?.();
        }

      } catch (err) {
        console.error('Send failed:', err);
      } finally {
        setIsSending(false);
      }
    };

    const randomTitle = useMemo(() => {
      return titles[Math.floor(Math.random() * titles.length)];
    }, []);

    return (
      <motion.div
        ref={containerRef}
        initial={{ scale: 0.9, opacity: 0 }}
        animate={{ scale: 1, opacity: 1 }}
        exit={{ scale: 0.9, opacity: 0 }}
        transition={{ duration: 0.2 }}
        className="absolute w-[50%] h-[75%] z-50 bg-card border border-border rounded-3xl p-4 shadow-2xl"
      >
        <div className="h-full flex flex-col">
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
            onChange={(e) => setTitle(e.target.value)}
            className="w-full mb-3 p-3 text-sm h-[20%] max-h-20 border border-border/50 rounded-2xl bg-muted/30 text-foreground placeholder-muted-foreground focus:outline-none focus:ring-2 focus:ring-libr-accent1/50 resize-none leading-tight"
          />

          {/* Body Editor */}
          {/* <div className={cn('w-full h-full bg-muted/30 border border-border/50 rounded-2xl resize-none focus:outline-none focus:ring-2 focus:ring-libr-accent1/50 transition-all duration-200', shake && 'border-red-500 ring-red-500 animate-shake')}> */}
          <div className="flex-1 overflow-hidden">
            <div className="h-full overflow-y-auto">
              {editor && <EditorContent editor={editor} />}
            </div>
          </div>

          {/* Footer */}
          <div className="flex justify-end items-center mt-4">
            <motion.button
              whileHover={{ scale: 1.05 }}
              whileTap={{ scale: 0.95 }}
              onClick={handleSend}
              disabled={isSending || bodyText.trim() === ''}
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
