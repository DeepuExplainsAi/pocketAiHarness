package ai.pocketharness.app

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch
import java.util.UUID

class ChatViewModel(
    private val bridge: BridgeClient,
    private val provider: String,
    private val cwd: String,
) : ViewModel() {
    private var chat = UUID.randomUUID().toString()
    private val _state = MutableStateFlow(ChatState())
    val state: StateFlow<ChatState> = _state

    private val _online = MutableStateFlow(false)
    val online: StateFlow<Boolean> = _online

    init {
        viewModelScope.launch {
            bridge.messages.collect { m -> if (m.belongsTo(chat)) _state.update { it.reduce(m) } }
        }
        viewModelScope.launch {
            bridge.connected.collect { up ->
                _online.value = up
                if (up) {
                    // Each connection starts a fresh provider session, so it gets a fresh id and sequence.
                    chat = UUID.randomUUID().toString()
                    _state.update { it.copy(lastSeq = 0, lostMessages = false) }
                    bridge.send(Protocol.start(chat, provider, cwd, mode = if (provider == "claude") "manual" else null))
                } else {
                    _state.update { it.disconnected() }
                }
            }
        }
        bridge.connect()
    }

    fun send(text: String) {
        if (text.isBlank() || !_online.value) return
        if (bridge.send(Protocol.send(chat, text))) _state.update { it.userSent(text) }
    }

    fun stop() { bridge.send(Protocol.interrupt(chat)) }

    fun decide(request: String, allow: Boolean) {
        if (bridge.send(Protocol.decide(request, allow, if (allow) null else "Denied by user"))) {
            _state.update { it.resolved(request) }
        }
    }

    override fun onCleared() {
        bridge.send(Protocol.close(chat))
        bridge.close()
    }

    private fun Inbound.belongsTo(id: String) = when (this) {
        is Inbound.Started -> chat == id
        is Inbound.Event -> chat == id
        is Inbound.Permission -> chat == id
        is Inbound.Closed -> chat == id
        is Inbound.Error -> chat == null || chat == id
    }
}
