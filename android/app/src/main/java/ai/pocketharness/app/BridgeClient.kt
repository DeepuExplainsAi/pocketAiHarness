package ai.pocketharness.app

import kotlinx.coroutines.channels.BufferOverflow
import kotlinx.coroutines.flow.MutableSharedFlow
import kotlinx.coroutines.flow.SharedFlow
import okhttp3.OkHttpClient
import okhttp3.Request
import okhttp3.Response
import okhttp3.WebSocket
import okhttp3.WebSocketListener

/** Connection to the local pah-bridge. Every request carries the per-run bearer token. */
class BridgeClient(private val listen: String, private val token: String) {
    private val http = OkHttpClient()
    private var socket: WebSocket? = null

    private val _messages = MutableSharedFlow<Inbound>(extraBufferCapacity = 256, onBufferOverflow = BufferOverflow.SUSPEND)
    val messages: SharedFlow<Inbound> = _messages

    private val _connected = MutableSharedFlow<Boolean>(replay = 1, extraBufferCapacity = 4)
    val connected: SharedFlow<Boolean> = _connected

    fun connect() {
        val request = Request.Builder().url("ws://$listen/ws").header("Authorization", "Bearer $token").build()
        socket = http.newWebSocket(request, object : WebSocketListener() {
            override fun onOpen(webSocket: WebSocket, response: Response) { _connected.tryEmit(true) }
            override fun onMessage(webSocket: WebSocket, text: String) {
                val m = Protocol.parse(text) ?: return
                // A full buffer means the UI cannot keep up. Dropping a message would hide a
                // permission card or a truncated reply, so end the connection instead.
                if (!_messages.tryEmit(m)) webSocket.close(1008, "client too slow")
            }
            override fun onClosed(webSocket: WebSocket, code: Int, reason: String) { _connected.tryEmit(false) }
            override fun onFailure(webSocket: WebSocket, t: Throwable, response: Response?) { _connected.tryEmit(false) }
        })
    }

    fun send(message: String): Boolean = socket?.send(message) ?: false

    fun close() {
        socket?.close(1000, null)
        socket = null
    }
}
