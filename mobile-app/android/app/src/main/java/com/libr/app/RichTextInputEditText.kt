package com.libr.app

import android.annotation.SuppressLint
import android.content.Context
import android.net.Uri
import android.os.Build
import android.view.inputmethod.EditorInfo
import android.view.inputmethod.InputConnection
import android.webkit.MimeTypeMap
import androidx.core.view.inputmethod.EditorInfoCompat
import androidx.core.view.inputmethod.InputConnectionCompat
import androidx.core.view.inputmethod.InputContentInfoCompat
import com.facebook.react.bridge.Arguments
import com.facebook.react.bridge.ReactContext
import com.facebook.react.bridge.WritableMap
import com.facebook.react.uimanager.ThemedReactContext
import com.facebook.react.uimanager.events.RCTEventEmitter
import com.facebook.react.views.textinput.ReactEditText
import java.io.File
import java.io.FileOutputStream

@SuppressLint("ViewConstructor")
class RichTextInputEditText(context: ThemedReactContext) : ReactEditText(context) {

    override fun onCreateInputConnection(outAttrs: EditorInfo): InputConnection? {
        val ic = super.onCreateInputConnection(outAttrs) ?: return null

        // Samsung and some other keyboards are picky. 
        // Explicitly listing these instead of */* or image/*.
        val mimeTypes = arrayOf(
            "image/gif", 
            "image/png", 
            "image/jpg", 
            "image/jpeg", 
            "image/webp",
            "image/vnd.microsoft.icon", // Some stickers use this
            "image/jxr",
            "image/bmp"
        )
        EditorInfoCompat.setContentMimeTypes(outAttrs, mimeTypes)

        val callback = InputConnectionCompat.OnCommitContentListener { inputContentInfo, flags, _ ->
            try {
                handleCommitContent(inputContentInfo, flags)
            } catch (e: Exception) {
                android.util.Log.e("LibrRichInput", "CommitContent failed: ${e.message}")
                false
            }
        }

        return InputConnectionCompat.createWrapper(ic, outAttrs, callback)
    }

    private fun handleCommitContent(inputContentInfo: InputContentInfoCompat, flags: Int): Boolean {
        android.util.Log.d("LibrRichInput", "Content received: ${inputContentInfo.contentUri}")

        // 1. Permissions handling
        var hasPermission = false
        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.N_MR1) {
            val lacksPermission = (flags and InputConnectionCompat.INPUT_CONTENT_GRANT_READ_URI_PERMISSION) != 0
            if (lacksPermission) {
                try {
                    inputContentInfo.requestPermission()
                    hasPermission = true
                } catch (e: Exception) {
                    android.util.Log.e("LibrRichInput", "Failed to request permission: ${e.message}")
                    return false
                }
            }
        }

        // 2. Metadata
        val uri = inputContentInfo.contentUri
        val mimeType = inputContentInfo.description.getMimeType(0) ?: "image/jpeg"
        
        // 3. Robust Copy to Cache
        // Samsung URIs can be transient, so we MUST copy them immediately.
        val cachedFile = copyUriToCache(context, uri, mimeType)
        
        if (hasPermission) {
            try {
                inputContentInfo.releasePermission()
            } catch (e: Exception) {
                // Ignore release errors
            }
        }

        if (cachedFile == null) {
            android.util.Log.e("LibrRichInput", "Failed to copy URI to cache")
            return false
        }
        
        // 4. Dispatch Event
        val eventData = Arguments.createMap()
        eventData.putString("uri", "file://" + cachedFile.absolutePath)
        eventData.putDouble("fileSize", cachedFile.length().toDouble())
        eventData.putString("mimeType", mimeType)
        eventData.putString("fileName", cachedFile.name)
        
        dispatchEvent("onMediaInserted", eventData)

        return true
    }

    private fun copyUriToCache(context: Context, uri: Uri, mimeType: String): File? {
        return try {
            val cacheDir = File(context.cacheDir, "rich_media_cache")
            if (!cacheDir.exists()) cacheDir.mkdirs()

            val extension = MimeTypeMap.getSingleton().getExtensionFromMimeType(mimeType) ?: "img"
            val fileName = "kbd_${System.currentTimeMillis()}.$extension"
            val destFile = File(cacheDir, fileName)

            context.contentResolver.openInputStream(uri)?.use { input ->
                FileOutputStream(destFile).use { output ->
                    input.copyTo(output)
                }
            }
            destFile
        } catch (e: Exception) {
            android.util.Log.e("LibrRichInput", "copyUriToCache Exception: ${e.message}")
            null
        }
    }

    private fun dispatchEvent(name: String, data: WritableMap) {
        val reactContext = context as ReactContext
        reactContext.getJSModule(RCTEventEmitter::class.java).receiveEvent(
            id,
            name,
            data
        )
    }
}
