package ai.pocketharness.app

import kotlinx.serialization.json.JsonObject
import kotlinx.serialization.json.contentOrNull
import kotlinx.serialization.json.JsonElement
import kotlinx.serialization.json.JsonPrimitive

enum class StreamStatus { Idle, Streaming, Complete, Interrupted, Failed }

sealed interface Item {
    data class User(val text: String) : Item
    data class Agent(val text: String, val status: StreamStatus) : Item
    data class Notice(val text: String) : Item
}

data class PendingPermission(val request: String, val tool: String, val description: String, val input: String)

data class ChatState(
    val items: List<Item> = emptyList(),
    val status: StreamStatus = StreamStatus.Idle,
    val permissions: List<PendingPermission> = emptyList(),
    val sessionId: String? = null,
    internal val lastSeq: Int = 0,
    val lostMessages: Boolean = false,
) {
    /** Applies one bridge message. A message with a seq already seen is dropped. */
    fun reduce(m: Inbound): ChatState = when (m) {
        is Inbound.Event -> when {
            m.seq <= lastSeq -> this
            // A gap means events were lost. Never present what follows as a complete answer.
            m.seq > lastSeq + 1 -> applyEvent(m.raw).copy(lastSeq = m.seq, lostMessages = true).let {
                if (it.status == StreamStatus.Complete) it.copy(status = StreamStatus.Interrupted).markAgent(StreamStatus.Interrupted) else it
            }
            else -> applyEvent(m.raw).copy(lastSeq = m.seq)
        }
        is Inbound.Permission -> copy(
            permissions = permissions + PendingPermission(m.request, m.tool, m.description, m.input?.toString().orEmpty()),
        )
        is Inbound.Error -> {
            val failing = status == StreamStatus.Streaming
            val marked = if (failing) markAgent(StreamStatus.Failed) else this
            marked.copy(items = marked.items + Item.Notice(m.message), status = if (failing) StreamStatus.Failed else status)
        }
        // Closing mid-reply must not look like a finished answer.
        is Inbound.Closed -> copy(permissions = emptyList()).let {
            if (it.status == StreamStatus.Streaming) it.copy(status = StreamStatus.Interrupted).markAgent(StreamStatus.Interrupted) else it
        }
        is Inbound.Started -> this
    }

    fun userSent(text: String) = copy(items = items + Item.User(text), status = StreamStatus.Streaming)

    fun resolved(request: String) = copy(permissions = permissions.filterNot { it.request == request })

    /** The bridge denies pending requests when the app disconnects, so the cards must go too. */
    fun disconnected() = copy(permissions = emptyList())
        .let { if (it.status == StreamStatus.Streaming) it.copy(status = StreamStatus.Interrupted).markAgent(StreamStatus.Interrupted) else it }

    private fun applyEvent(raw: JsonObject): ChatState {
        val type = raw.str("type")
        return when (type) {
            "system" -> if (raw.str("subtype") == "init") copy(sessionId = raw.str("session_id") ?: sessionId) else this
            "stream_event" -> textDelta(raw)?.let { appendAgent(it) } ?: this
            "bridge_error" -> copy(items = items + Item.Notice(raw.str("message") ?: "Provider error"), status = StreamStatus.Failed).markAgent(StreamStatus.Failed)
            "result" -> {
                val failed = raw.str("is_error") == "true"
                val s = if (failed) StreamStatus.Failed else StreamStatus.Complete
                copy(status = s).markAgent(s)
            }
            else -> this
        }
    }

    private fun textDelta(raw: JsonObject): String? {
        val ev = raw["event"] as? JsonObject ?: return null
        if (ev.str("type") != "content_block_delta") return null
        val delta = ev["delta"] as? JsonObject ?: return null
        return if (delta.str("type") == "text_delta") delta.str("text") else null
    }

    private fun appendAgent(text: String): ChatState {
        val last = items.lastOrNull()
        val next = if (last is Item.Agent && last.status == StreamStatus.Streaming) {
            items.dropLast(1) + last.copy(text = last.text + text)
        } else {
            items + Item.Agent(text, StreamStatus.Streaming)
        }
        return copy(items = next, status = StreamStatus.Streaming)
    }

    private fun markAgent(s: StreamStatus?): ChatState {
        if (s == null) return this
        val last = items.lastOrNull() as? Item.Agent ?: return this
        return copy(items = items.dropLast(1) + last.copy(status = s))
    }
}

/** Reads a string field; any other shape counts as absent instead of throwing. */
private fun JsonObject.str(key: String): String? = ((this[key] as? JsonElement) as? JsonPrimitive)?.contentOrNull
