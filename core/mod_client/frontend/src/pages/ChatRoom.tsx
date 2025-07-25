
import React, { useEffect, useRef } from 'react';
import { motion } from 'framer-motion';
import { useAppStore } from '../store/useAppStore';
import { MessageBubble } from '../components/chat/MessageBubble';
import { MessageInput } from '../components/chat/MessageInput';
import { apiService } from '../services/api';
import { TopBar } from '../components/layout/TopBar';
import { ArrowDown, Clock, Calendar, RotateCcw } from 'lucide-react';

export const ChatRoom: React.FC = () => {
  const { 
    currentCommunity, 
    messages, 
    setMessages, 
    setLoading, 
    isLoading, 
  } = useAppStore();
  
  const messagesEndRef = useRef<HTMLDivElement>(null);
  const [sortByNewest, setSortByNewest] = React.useState(true);
  const [showScrollButton, setShowScrollButton] = React.useState(false);
  const [selectedDate, setSelectedDate] = React.useState<Date | null>(null);
  const [showDatePicker, setShowDatePicker] = React.useState(false);
  const messageInputRef = useRef<HTMLDivElement>(null);
  const [inputHeight, setInputHeight] = React.useState(0);
  const scrollButtonRef = useRef<HTMLButtonElement>(null);


  // Measure input height
  useEffect(() => {
    if (!messageInputRef.current) return;

    const resizeObserver = new ResizeObserver(entries => {
      for (const entry of entries) {
        const height = entry.contentRect.height;
        setInputHeight(height);
        
        // Force reflow for Wails/DOM issues
        window.dispatchEvent(new Event('resize'));
      }
    });

    resizeObserver.observe(messageInputRef.current);

    return () => resizeObserver.disconnect();
  }, []);

  useEffect(() => {
    const container = messagesContainerRef.current;
    if (!container) return;

    const handleScroll = () => {
      const { scrollTop, scrollHeight, clientHeight } = container;
      const atBottom = scrollTop + clientHeight >= scrollHeight - 50; // buffer
      setShowScrollButton(!atBottom);
    };

    container.addEventListener("scroll", handleScroll);
    return () => container.removeEventListener("scroll", handleScroll);
  }, []);
  

  // useEffect(() => {
  //   if (currentCommunity) {
  //     loadMessages();
  //   }
  // }, [currentCommunity]);

  const loadMessages = async () => {
    if (!currentCommunity) return;
    
    setLoading(true);
    try {
      const fetchedMessages = await apiService.getMessages(currentCommunity.id);
      setMessages(fetchedMessages);
    } catch (error) {
      console.error('Failed to load messages:', error);
    } finally {
      setLoading(false);
    }
  };

  const scrollToBottom = () => {
    const container = messagesContainerRef.current;
    if (container) {
      container.scrollTo({ top: container.scrollHeight, behavior: "smooth" });
    }
  };


  const sortedMessages = React.useMemo(() => {
    let filtered = [...messages];
    
    // Filter by selected date if any
    if (selectedDate) {
      const selectedDateString = selectedDate.toDateString();
      filtered = filtered.filter(message => 
        new Date(message.timestamp).toDateString() === selectedDateString
      );
    }
    
    return filtered.sort((a, b) => {
      const dateA = new Date(a.timestamp).getTime();
      const dateB = new Date(b.timestamp).getTime();
      return sortByNewest ? dateB - dateA : dateA - dateB;
    });
  }, [messages, sortByNewest, selectedDate]);

  const messagesContainerRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (!messagesContainerRef.current) return;

    const el = messagesContainerRef.current;
    if (sortByNewest) {
      el.scrollTo({ top: 0, behavior: 'smooth' });
    } else {
      el.scrollTo({ top: el.scrollHeight, behavior: 'smooth' });
    }
  }, [sortedMessages.length, sortByNewest]);

  if (!currentCommunity) {
    return (
      <div className="flex-1 flex items-center justify-center bg-libr-primary">
        <motion.div
          initial={{ opacity: 0, scale: 0.9 }}
          animate={{ opacity: 1, scale: 1 }}
          className="text-center"
        >
          <div className="w-20 h-20 bg-libr-accent1/20 rounded-2xl flex items-center justify-center mx-auto mb-4">
            <motion.div
              animate={{ rotate: 360 }}
              transition={{ duration: 2, repeat: Infinity, ease: "linear" }}
              className="w-10 h-10 border-3 border-libr-accent1 border-t-transparent rounded-full"
            />
          </div>
          <h2 className="text-xl font-semibold text-foreground mb-2">
            Welcome to libr
          </h2>
          <p className="text-muted-foreground">
            Select a community from the sidebar to start chatting
          </p>
        </motion.div>
      </div>
    );
  }

  return (
    <div className="flex flex-col h-screen">
      <TopBar />
      
      {/* Messages Area */}
      <div className="flex-1 min-h-0 flex flex-col">
        {/* Toolbar */}
        <div className="bg-card/50 border-b border-border/30 h-16 p-3 pl-5 flex items-center justify-between">
          <div className="flex items-center space-x-3">
            {/* <button
              onClick={() => setSortByNewest(!sortByNewest)}
              className="libr-button bg-muted/50 hover:bg-muted flex items-center space-x-2 text-sm"
            >
              <Clock className="w-4 h-4" />
              <span className='mt-0.5'>{sortByNewest ? 'Newest First' : 'Oldest First'}</span>
            </button> */}
            <button
              onClick={loadMessages}
              className="libr-button bg-muted/50 hover:bg-muted flex items-center space-x-2 text-sm"
              title="Reload Messages"
            >
              <RotateCcw className="w-4 h-4" />
              <span className='mt-0.5'>Reload</span>
            </button>
            <span className="text-sm text-muted-foreground">
              {selectedDate 
                ? `${sortedMessages.length} messages on ${selectedDate.toLocaleDateString()}`
                : `${messages.length} messages`
              }
            </span>
          </div>
          
          {/* <div className="flex items-center space-x-2 relative">
            <div
              // onClick={() => setShowDatePicker(!showDatePicker)}
              className="libr-button bg-muted/50 hover:bg-muted flex items-center space-x-2 text-sm"
            >
              <Calendar className="w-4 h-4" />
              <span className='mt-0.5'>{selectedDate ? selectedDate.toLocaleDateString() : 'Today'}</span>
            </div>
            
            {selectedDate && (
              <button
                onClick={() => setSelectedDate(null)}
                className="libr-button bg-red-500/10 hover:bg-red-500/20 text-red-500 text-sm"
              >
                Clear
              </button>
            )}
            
            {showDatePicker && (
              <motion.div
                initial={{ opacity: 0, scale: 0.95 }}
                animate={{ opacity: 1, scale: 1 }}
                className="absolute right-0 top-full mt-2 bg-card border border-border rounded-lg shadow-lg z-20 p-4"
              >
                <input
                  type="date"
                  onChange={(e) => {
                    if (e.target.value) {
                      setSelectedDate(new Date(e.target.value));
                      setShowDatePicker(false);
                    }
                  }}
                  className="bg-muted border border-border rounded px-3 py-2 text-sm"
                />
              </motion.div>
            )}
          </div> */}
        </div>

        {/* Messages */}
        <div className="flex-1 overflow-y-auto p-4 space-y-2" ref={messagesContainerRef} style={{ paddingBottom: Math.max(inputHeight, 100) + 36 }}>
          {isLoading ? (
            <div className="flex items-center justify-center py-12">
              <motion.div
                animate={{ rotate: 360 }}
                transition={{ duration: 1, repeat: Infinity, ease: "linear" }}
                className="w-8 h-8 border-2 border-libr-accent1 border-t-transparent rounded-full"
              />
            </div>
          ) : sortedMessages.length === 0 ? (
            <motion.div
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              className="flex items-center justify-center py-12"
            >
              <div className="text-center">
                <div className="w-16 h-16 bg-libr-accent1/20 rounded-2xl flex items-center justify-center mx-auto mb-4">
                  <motion.div
                    animate={{ scale: [1, 1.1, 1] }}
                    transition={{ duration: 2, repeat: Infinity }}
                    className="text-2xl"
                  >
                    💬
                  </motion.div>
                </div>
                <h3 className="text-lg font-medium text-foreground mb-2">
                  Start the conversation
                </h3>
                <p className="text-muted-foreground">
                  Be the first to send a message in #{currentCommunity.name}
                </p>
              </div>
            </motion.div>
          ) : (
            sortedMessages.map((message) => (
              <MessageBubble
                key={message.id}
                message={message}
              />
            ))
          )}
          <div ref={messagesEndRef} />
        </div>

        {/* Scroll to bottom button */}
        {showScrollButton && (
          <motion.button
            ref={scrollButtonRef}
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            exit={{ opacity: 0 }}
            transition={{ duration: 0.5 }}
            onClick={scrollToBottom}
            className="fixed right-6 w-12 h-12 bg-libr-accent1 hover:bg-libr-accent1/80 rounded-full flex items-center justify-center shadow-lg transition-all"
            style={{
              bottom: Math.max(inputHeight, 80) + 45,
            }}
          >
            <ArrowDown className="w-5 h-5 text-white" />
          </motion.button>
        )}
      </div>

      {/* Fixed Message Input */}
      <MessageInput ref={messageInputRef} />
    </div>
  );
};
