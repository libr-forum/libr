package com.libr.app

import com.facebook.react.bridge.*
import com.facebook.react.bridge.Arguments
import com.facebook.react.bridge.WritableMap
import com.facebook.react.modules.core.DeviceEventManagerModule
import java.lang.reflect.Proxy
import java.lang.reflect.InvocationHandler
import java.lang.reflect.Method
import android.app.NotificationChannel
import android.app.NotificationManager
import android.app.PendingIntent
import android.content.Context
import android.content.Intent
import android.os.Build
import androidx.core.app.NotificationCompat

class LibrCoreModule(reactContext: ReactApplicationContext) : ReactContextBaseJavaModule(reactContext) {

    init {
        android.util.Log.e("LibrApp", "===== LibrCoreModule init started =====")
        android.util.Log.d("LibrApp", "LibrCoreModule init")
        try {
            val bridgeClass = Class.forName("bridge.Bridge")
            val notifHandlerClass = Class.forName("bridge.NotificationHandler")
            
            val notifProxy = Proxy.newProxyInstance(
                notifHandlerClass.classLoader,
                arrayOf(notifHandlerClass),
                object : InvocationHandler {
                    override fun invoke(proxy: Any, method: Method, args: Array<out Any>?): Any? {
                        if (method.name == "onNotification" && args != null && args.size >= 2) {
                            val title = args[0] as? String ?: ""
                            val message = args[1] as? String ?: ""
                            android.util.Log.d("LibrApp", "Direct background notification from Go: $title - $message")
                            postNotificationDirectly(title, message)
                        }
                        return null
                    }
                }
            )
            
            val setNotificationHandler = bridgeClass.getMethod("setNotificationHandler", notifHandlerClass)
            setNotificationHandler.invoke(null, notifProxy)
            android.util.Log.e("LibrApp", "===== LibrCoreModule setNotificationHandler SUCCESS =====")
        } catch (e: Throwable) {
            android.util.Log.e("LibrApp", "===== LibrCoreModule setNotificationHandler failed: ${e.message} =====", e)
        }
    }

    private fun sendEvent(eventName: String, params: WritableMap?) {
        reactApplicationContext
            .getJSModule(DeviceEventManagerModule.RCTDeviceEventEmitter::class.java)
            .emit(eventName, params)
    }

    @ReactMethod
    fun addListener(eventName: String) {
        // Required for NativeEventEmitter
    }

    @ReactMethod
    fun removeListeners(count: Int) {
        // Required for NativeEventEmitter
    }

    override fun getName(): String {
        return "LibrCore"
    }

    @ReactMethod
    fun stopNode(promise: Promise) {
        try {
            Class.forName("bridge.Bridge").getMethod("stopNode").invoke(null)
            promise.resolve(null)
        } catch (e: Exception) {
            promise.reject("ERR_STOP", e.message)
        }
    }

    // Bridge calls by reflection, so the module compiles against any AAR version. Every
    // bridge parameter is a String: getMethod matches boxed types, which never equal
    // Go's primitive bool/int parameters.
    private fun callBridgeString(method: String, vararg args: Any?): Any? {
        return try {
            val argTypes = args.map { it?.javaClass ?: String::class.java }.toTypedArray()
            val m = Class.forName("bridge.Bridge").getMethod(method, *argTypes)
            m.invoke(null, *args)
        } catch (e: Throwable) {
            null
        }
    }

    // Quick calls answer on the JS thread; network calls run on their own thread.
    private fun quick(promise: Promise, code: String, fallback: Any?, call: () -> Any?) {
        try { promise.resolve(call() ?: fallback) } catch (e: Exception) { promise.reject(code, e.message) }
    }

    private fun slow(promise: Promise, code: String, fallback: Any?, call: () -> Any?) {
        Thread { quick(promise, code, fallback, call) }.start()
    }

    @ReactMethod fun initApp(promise: Promise) {
        val keyDir = reactApplicationContext.filesDir.absolutePath + "/keys"
        quick(promise, "ERR_INIT_APP", "") { callBridgeString("initApp", keyDir)?.toString() }
    }
    @ReactMethod fun getPublicKey(promise: Promise) = quick(promise, "ERR_GET_KEY", "") { callBridgeString("getPublicKey")?.toString() }
    @ReactMethod fun regenKeys(promise: Promise) = slow(promise, "ERR_REGEN_KEYS", "") { callBridgeString("regenKeys")?.toString() }
    @ReactMethod fun enableIncognito(promise: Promise) = slow(promise, "ERR_ENABLE_INCOGNITO", "") { callBridgeString("enableIncognito")?.toString() }
    @ReactMethod fun disableIncognito(promise: Promise) = slow(promise, "ERR_DISABLE_INCOGNITO", "") { callBridgeString("disableIncognito")?.toString() }
    @ReactMethod fun isIncognitoEnabled(promise: Promise) = quick(promise, "ERR_IS_INCOGNITO", false) { callBridgeString("isIncognitoEnabled") as? Boolean }

    @ReactMethod fun startNode(configJSON: String, promise: Promise) = slow(promise, "ERR_START", "error:not_built") { callBridgeString("startNode", configJSON)?.toString() }
    @ReactMethod fun getPeerID(promise: Promise) = quick(promise, "ERR_PEER_ID", "") { callBridgeString("getPeerID")?.toString() }
    @ReactMethod fun networkStatus(promise: Promise) = quick(promise, "ERR_STATUS", "{}") { callBridgeString("networkStatus")?.toString() }
    @ReactMethod fun listCommunities(promise: Promise) = quick(promise, "ERR_COMMUNITIES", "[]") { callBridgeString("listCommunities")?.toString() }
    @ReactMethod fun isMod(community: String, promise: Promise) = quick(promise, "ERR_IS_MOD", false) { callBridgeString("isMod", community) as? Boolean }

    @ReactMethod fun post(community: String, content: String, promise: Promise) = slow(promise, "ERR_POST", "error:not_built") { callBridgeString("post", community, content)?.toString() }
    @ReactMethod fun feed(community: String, promise: Promise) = slow(promise, "ERR_FEED", "[]") { callBridgeString("feed", community)?.toString() }
    @ReactMethod fun report(hash: String, reason: String, promise: Promise) = slow(promise, "ERR_REPORT", "error:not_built") { callBridgeString("report", hash, reason)?.toString() }
    @ReactMethod fun deletePost(hash: String, promise: Promise) = slow(promise, "ERR_DELETE", "error:not_built") { callBridgeString("delete", hash)?.toString() }
    @ReactMethod fun tick(promise: Promise) = slow(promise, "ERR_TICK", "error:not_built") { callBridgeString("tick")?.toString() }
    @ReactMethod fun pendingStatus(promise: Promise) = quick(promise, "ERR_PENDING", "{}") { callBridgeString("pendingStatus")?.toString() }
    @ReactMethod fun dismissReport(key: String, promise: Promise) = quick(promise, "ERR_DISMISS", "ok") { callBridgeString("dismissReport", key)?.toString() }

    @ReactMethod fun modQueue(promise: Promise) = quick(promise, "ERR_MOD_QUEUE", "[]") { callBridgeString("modQueue")?.toString() }
    @ReactMethod fun modDecide(key: String, decision: String, promise: Promise) = slow(promise, "ERR_MOD_DECIDE", "error:not_built") { callBridgeString("modDecide", key, decision)?.toString() }

    @ReactMethod fun generateAlias(key: String, promise: Promise) = quick(promise, "ERR_ALIAS", key.take(8)) { callBridgeString("generateAlias", key)?.toString() }
    @ReactMethod fun generateAvatar(key: String, promise: Promise) = quick(promise, "ERR_AVATAR", "") { callBridgeString("generateAvatar", key)?.toString() }

    companion object {
        const val NOTIFICATION_CHANNEL_ID = "libr_moderation_alerts"
        const val NOTIFICATION_CHANNEL_NAME = "Moderation Alerts"
    }

    private fun createNotificationChannel() {
        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.O) {
            val channel = NotificationChannel(
                NOTIFICATION_CHANNEL_ID,
                NOTIFICATION_CHANNEL_NAME,
                NotificationManager.IMPORTANCE_HIGH
            ).apply {
                description = "Notifications for approved or rejected messages and reports"
                enableLights(true)
                enableVibration(true)
            }
            val manager = reactApplicationContext.getSystemService(Context.NOTIFICATION_SERVICE) as? NotificationManager
            manager?.createNotificationChannel(channel)
        }
    }

    fun postNotificationDirectly(title: String, message: String) {
        try {
            createNotificationChannel()

            val intent = Intent(reactApplicationContext, MainActivity::class.java).apply {
                flags = Intent.FLAG_ACTIVITY_SINGLE_TOP or Intent.FLAG_ACTIVITY_CLEAR_TOP
            }
            val pendingIntent = PendingIntent.getActivity(
                reactApplicationContext,
                0,
                intent,
                PendingIntent.FLAG_UPDATE_CURRENT or (if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.M) PendingIntent.FLAG_IMMUTABLE else 0)
            )

            val appIcon = reactApplicationContext.applicationInfo.icon.let {
                if (it != 0) it else android.R.drawable.ic_dialog_info
            }

            val builder = NotificationCompat.Builder(reactApplicationContext, NOTIFICATION_CHANNEL_ID)
                .setSmallIcon(appIcon)
                .setContentTitle(title)
                .setContentText(message)
                .setStyle(NotificationCompat.BigTextStyle().bigText(message))
                .setPriority(NotificationCompat.PRIORITY_HIGH)
                .setAutoCancel(true)
                .setContentIntent(pendingIntent)
                .setDefaults(NotificationCompat.DEFAULT_ALL)

            val manager = reactApplicationContext.getSystemService(Context.NOTIFICATION_SERVICE) as? NotificationManager
            val notificationId = (System.currentTimeMillis() % 100000).toInt()
            manager?.notify(notificationId, builder.build())
        } catch (e: Exception) {
            android.util.Log.e("LibrApp", "Failed to post direct notification: ${e.message}", e)
        }
    }

    @ReactMethod
    fun showSystemNotification(title: String, message: String, promise: Promise) {
        Thread {
            try {
                postNotificationDirectly(title, message)
                promise.resolve("ok")
            } catch (e: Exception) {
                android.util.Log.e("LibrApp", "Failed to show system notification: ${e.message}", e)
                promise.reject("ERR_NOTIFICATION", e.message)
            }
        }.start()
    }
}
