import React from 'react';
import { requireNativeComponent, ViewProps, NativeSyntheticEvent, StyleSheet, Platform } from 'react-native';

interface MediaInsertedEvent {
  uri: string;
  fileSize: number;
  mimeType: string;
  fileName: string;
}

interface RichTextInputProps extends ViewProps {
  value?: string;
  placeholder?: string;
  multiline?: boolean;
  textAlignVertical?: string;
  placeholderTextColor?: string;
  onChangeText?: (text: string) => void;
  onMediaInserted?: (event: any) => void;
  autoFocus?: boolean;
}

const NativeRichTextInput = requireNativeComponent<any>('RichTextInput');

const RichTextInput = (props: RichTextInputProps) => {
  const { onChangeText, value, ...rest } = props;

  const handleChange = (event: any) => {
    // Standard ReactTextInputManager emits an object with text field
    if (onChangeText && event.nativeEvent && event.nativeEvent.text !== undefined) {
      onChangeText(event.nativeEvent.text);
    }
  };

  return (
    <NativeRichTextInput 
      {...rest} 
      text={value} // ReactTextInputManager uses the 'text' property natively
      onChange={handleChange}
      style={[styles.default, props.style]} 
    />
  );
};

const styles = StyleSheet.create({
  default: {
    minHeight: 120,
    backgroundColor: 'transparent',
  },
});

export default RichTextInput;
