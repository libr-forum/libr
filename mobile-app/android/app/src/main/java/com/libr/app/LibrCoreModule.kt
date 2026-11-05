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
        try {
            val bridgeClass = Class.forName("bridge.Bridge")
            val handlerClass = Class.forName("bridge.MessageHandler")
            
            val handlerProxy = Proxy.newProxyInstance(
                handlerClass.classLoader,
                arrayOf(handlerClass),
                object : InvocationHandler {
                    override fun invoke(proxy: Any, method: Method, args: Array<out Any>?): Any? {
                        if (method.name == "onMessage" && args != null && args.size == 2) {
                            val peerID = args[0] as? String
                            val msg = args[1] as? String
                            val params = Arguments.createMap()
                            params.putString("peerID", peerID)
                            params.putString("message", msg)
                            sendEvent("onLibrMessage", params)
                        }
                        return null
                    }
                }
            )
            
            val setMessageHandler = bridgeClass.getMethod("setMessageHandler", handlerClass)
            setMessageHandler.invoke(null, handlerProxy)
            android.util.Log.e("LibrApp", "===== LibrCoreModule init SUCCESS =====")
        } catch (e: Throwable) {
            android.util.Log.e("LibrApp", "===== LibrCoreModule setMessageHandler failed: ${e.message} =====", e)
        }

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
    fun initNode(relayAddrsJson: String, promise: Promise) {
        try {
            promise.resolve(callBridgeString("initNode", relayAddrsJson)?.toString() ?: "")
        } catch (e: Exception) {
            promise.reject("ERR_INIT", e.message)
        }
    }

    @ReactMethod
    fun sendMessage(targetPeerID: String, message: String, promise: Promise) {
        try {
            promise.resolve(callBridgeString("sendMessage", targetPeerID, message)?.toString() ?: "")
        } catch (e: Exception) {
            promise.reject("ERR_SEND", e.message)
        }
    }

    @ReactMethod
    fun getPeerID(promise: Promise) {
        try {
            promise.resolve(callBridgeString("getPeerID")?.toString() ?: "")
        } catch (e: Exception) {
            promise.reject("ERR_PEER_ID", e.message)
        }
    }

    @ReactMethod
    fun stopNode(promise: Promise) {
        try {
            val m = Class.forName("bridge.Bridge").getMethod("stopNode")
            m.invoke(null)
            promise.resolve(null)
        } catch (e: Exception) {
            promise.reject("ERR_STOP", e.message)
        }
    }

    // ── New methods (available after rebuilding the Go AAR with build-android.ps1) ──
    // These use reflection so the module compiles against any AAR version.
    // On the old AAR they resolve to null and the JS mock / guard handles it.

    private fun callBridgeString(method: String, vararg args: Any?): Any? {
        return try {
            val argTypes = args.map { it?.javaClass ?: String::class.java }.toTypedArray()
            val m = Class.forName("bridge.Bridge").getMethod(method, *argTypes)
            m.invoke(null, *args)
        } catch (e: Throwable) {
            null
        }
    }

    @ReactMethod fun initApp(serverURL: String, promise: Promise) {
        val keyDir = reactApplicationContext.filesDir.absolutePath + "/keys"
        try { promise.resolve(callBridgeString("initApp", keyDir, serverURL)?.toString() ?: "") }
        catch (e: Exception) { promise.reject("ERR_INIT_APP", e.message) }
    }

    @ReactMethod fun getPublicKey(promise: Promise) {
        try { promise.resolve(callBridgeString("getPublicKey")?.toString() ?: "") }
        catch (e: Exception) { promise.reject("ERR_GET_KEY", e.message) }
    }

    @ReactMethod fun regenKeys(promise: Promise) {
        try { promise.resolve(callBridgeString("regenKeys")?.toString() ?: "") }
        catch (e: Exception) { promise.reject("ERR_REGEN_KEYS", e.message) }
    }

    @ReactMethod fun enableIncognito(promise: Promise) {
        try { promise.resolve(callBridgeString("enableIncognito")?.toString() ?: "") }
        catch (e: Exception) { promise.reject("ERR_ENABLE_INCOGNITO", e.message) }
    }

    @ReactMethod fun disableIncognito(promise: Promise) {
        try { promise.resolve(callBridgeString("disableIncognito")?.toString() ?: "") }
        catch (e: Exception) { promise.reject("ERR_DISABLE_INCOGNITO", e.message) }
    }

    @ReactMethod fun isIncognitoEnabled(promise: Promise) {
        try { promise.resolve((callBridgeString("isIncognitoEnabled") as? Boolean) ?: false) }
        catch (e: Exception) { promise.reject("ERR_IS_INCOGNITO", e.message) }
    }

    @ReactMethod fun getRelayAddresses(promise: Promise) {
        try { promise.resolve(callBridgeString("getRelayAddresses")?.toString() ?: "[]") }
        catch (e: Exception) { promise.reject("ERR_RELAY_ADDRS", e.message) }
    }

    @ReactMethod fun getOnlineMods(promise: Promise) {
        try { promise.resolve(callBridgeString("getOnlineMods")?.toString() ?: "[]") }
        catch (e: Exception) { promise.reject("ERR_ONLINE_MODS", e.message) }
    }

    @ReactMethod fun amIMod(promise: Promise) {
        try { promise.resolve((callBridgeString("amIMod") as? Boolean) ?: false) }
        catch (e: Exception) { promise.reject("ERR_AM_I_MOD", e.message) }
    }

    @ReactMethod fun sendTextMessage(content: String, promise: Promise) {
        Thread {
            try { promise.resolve(callBridgeString("sendTextMessage", content)?.toString() ?: "{\"status\":\"error:not_built\"}") }
            catch (e: Exception) { promise.reject("ERR_SEND_TEXT", e.message) }
        }.start()
    }

    @ReactMethod fun fetchMessages(promise: Promise) {
        Thread {
            try { promise.resolve(callBridgeString("fetchMessages")?.toString() ?: "[]") }
            catch (e: Exception) { promise.reject("ERR_FETCH", e.message) }
        }.start()
    }

    @ReactMethod fun deleteMessage(msgCertJSON: String, promise: Promise) {
        Thread {
            try { promise.resolve(callBridgeString("deleteMessage", msgCertJSON)?.toString() ?: "error:not_built") }
            catch (e: Exception) { promise.reject("ERR_DELETE", e.message) }
        }.start()
    }

    @ReactMethod fun reportMessage(msgCertJSON: String, reason: String, promise: Promise) {
        Thread {
            try { promise.resolve(callBridgeString("reportMessage", msgCertJSON, reason)?.toString() ?: "error:not_built") }
            catch (e: Exception) { promise.reject("ERR_REPORT", e.message) }
        }.start()
    }

    @ReactMethod fun startCron(promise: Promise) {
        Thread {
            try { promise.resolve(callBridgeString("startCron")?.toString() ?: "error:not_built") }
            catch (e: Exception) { promise.reject("ERR_START_CRON", e.message) }
        }.start()
    }

    @ReactMethod fun stopCron(promise: Promise) {
        Thread {
            try { promise.resolve(callBridgeString("stopCron")?.toString() ?: "error:not_built") }
            catch (e: Exception) { promise.reject("ERR_STOP_CRON", e.message) }
        }.start()
    }

    @ReactMethod fun tickCron(promise: Promise) {
        Thread {
            try { promise.resolve(callBridgeString("tickCron")?.toString() ?: "error:not_built") }
            catch (e: Exception) { promise.reject("ERR_TICK_CRON", e.message) }
        }.start()
    }

    @ReactMethod fun fetchReports(promise: Promise) {
        Thread {
            try { promise.resolve(callBridgeString("fetchReports")?.toString() ?: "[]") }
            catch (e: Exception) { promise.reject("ERR_FETCH_REPORTS", e.message) }
        }.start()
    }

    @ReactMethod fun getPendingReports(promise: Promise) {
        Thread {
            try { promise.resolve(callBridgeString("getPendingReports")?.toString() ?: "{}") }
            catch (e: Exception) { promise.reject("ERR_GET_PENDING_REPORTS", e.message) }
        }.start()
    }

    @ReactMethod fun dismissReport(msgSign: String, promise: Promise) {
        Thread {
            try { promise.resolve(callBridgeString("dismissReport", msgSign)?.toString() ?: "ok") }
            catch (e: Exception) { promise.reject("ERR_DISMISS_REPORT", e.message) }
        }.start()
    }

        @ReactMethod fun resendToDb(msgCertJSON: String, promise: Promise) {
        Thread {
            try { promise.resolve(callBridgeString("resendToDb", msgCertJSON)?.toString() ?: "error:not_built") }
            catch (e: Exception) { promise.reject("ERR_RESEND", e.message) }
        }.start()
    }

    @ReactMethod fun moderateMessage(msgCertJSON: String, action: String, promise: Promise) {
        Thread {
            try { promise.resolve(callBridgeString("moderateMessage", msgCertJSON, action)?.toString() ?: "error:not_built") }
            catch (e: Exception) { promise.reject("ERR_MODERATE", e.message) }
        }.start()
    }

    @ReactMethod fun generateAlias(key: String, promise: Promise) {
        try { promise.resolve(callBridgeString("generateAlias", key)?.toString() ?: key.take(8)) }
        catch (e: Exception) { promise.reject("ERR_ALIAS", e.message) }
    }

    @ReactMethod fun generateAvatar(key: String, promise: Promise) {
        try { promise.resolve(callBridgeString("generateAvatar", key)?.toString() ?: "") }
        catch (e: Exception) { promise.reject("ERR_AVATAR", e.message) }
    }

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
