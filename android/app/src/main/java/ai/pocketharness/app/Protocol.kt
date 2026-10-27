package ai.pocketharness.app

import kotlinx.serialization.json.Json
import kotlinx.serialization.json.JsonElement
import kotlinx.serialization.json.JsonObject
import kotlinx.serialization.json.JsonPrimitive
import kotlinx.serialization.json.buildJsonObject
import kotlinx.serialization.json.contentOrNull
import kotlinx.serialization.json.int
import kotlinx.serialization.json.jsonObject
import kotlinx.serialization.json.jsonPrimitive
import kotlinx.serialization.json.put

/** Messages from the bridge. See docs/protocol.md. */
sealed interface Inbound {
    data class Started(val chat: String) : Inbound
    data class Event(val chat: String, val seq: Int, val raw: JsonObject) : Inbound
    data class Permission(
        val chat: String,
        val request: String,
        val tool: String,
        val description: String,
        val input: JsonElement?,
    ) : Inbound
    data class Error(val chat: String?, val message: String) : Inbound
    data class Closed(val chat: String) : Inbound
}

object Protocol {
    private val json = Json { ignoreUnknownKeys = true }

    fun parse(text: String): Inbound? {
        val o = runCatching { json.parseToJsonElement(text).jsonObject }.getOrNull() ?: return null
        val chat = o.str("chat")
        return when (o.str("type")) {
            "started" -> chat?.let { Inbound.Started(it) }
            "closed" -> chat?.let { Inbound.Closed(it) }
            "error" -> Inbound.Error(chat, o.str("message") ?: "unknown error")
            "event" -> {
                val raw = o["raw"] as? JsonObject
                val seq = o["seq"]?.jsonPrimitive?.int
                if (chat != null && raw != null && seq != null) Inbound.Event(chat, seq, raw) else null
            }
            "permission" -> {
                val request = o.str("request")
                val tool = o.str("tool")
                if (chat != null && request != null && tool != null) {
                    Inbound.Permission(chat, request, tool, o.str("description").orEmpty(), o["input"])
                } else null
            }
            else -> null
        }
    }

    fun start(chat: String, provider: String, cwd: String, model: String? = null, mode: String? = null, resume: String? = null) =
        buildJsonObject {
            put("op", "start"); put("chat", chat); put("provider", provider); put("cwd", cwd)
            model?.let { put("model", it) }; mode?.let { put("mode", it) }; resume?.let { put("resume", it) }
        }.toString()

    fun send(chat: String, text: String) = buildJsonObject { put("op", "send"); put("chat", chat); put("text", text) }.toString()
    fun interrupt(chat: String) = buildJsonObject { put("op", "interrupt"); put("chat", chat) }.toString()
    fun mode(chat: String, mode: String) = buildJsonObject { put("op", "mode"); put("chat", chat); put("mode", mode) }.toString()
    fun model(chat: String, model: String) = buildJsonObject { put("op", "model"); put("chat", chat); put("model", model) }.toString()
    fun close(chat: String) = buildJsonObject { put("op", "close"); put("chat", chat) }.toString()

    fun decide(request: String, allow: Boolean, reason: String? = null) = buildJsonObject {
        put("op", "decide"); put("request", request); put("allow", allow); reason?.let { put("reason", it) }
    }.toString()

    private fun JsonObject.str(key: String): String? = (this[key] as? JsonPrimitive)?.contentOrNull
}
