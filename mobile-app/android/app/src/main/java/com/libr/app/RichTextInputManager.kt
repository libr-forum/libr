package com.libr.app

import com.facebook.react.module.annotations.ReactModule
import com.facebook.react.uimanager.ThemedReactContext
import com.facebook.react.views.textinput.ReactTextInputManager
import com.facebook.react.common.MapBuilder
import com.facebook.react.uimanager.ViewManager

@ReactModule(name = "RichTextInput")
class RichTextInputManager : ReactTextInputManager() {

    override fun getName(): String {
        return "RichTextInput"
    }

    override fun createViewInstance(reactContext: ThemedReactContext): RichTextInputEditText {
        return RichTextInputEditText(reactContext)
    }

    override fun getExportedCustomDirectEventTypeConstants(): Map<String, Any> {
        val base = super.getExportedCustomDirectEventTypeConstants() ?: mutableMapOf<String, Any>()
        val map = HashMap<String, Any>(base)
        map["onMediaInserted"] = MapBuilder.of("registrationName", "onMediaInserted")
        return map
    }
}
