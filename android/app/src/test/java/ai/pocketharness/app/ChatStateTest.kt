package ai.pocketharness.app

import kotlinx.serialization.json.Json
import kotlinx.serialization.json.jsonObject
import org.junit.Assert.assertEquals
import org.junit.Assert.assertNotNull
import org.junit.Assert.assertNull
import org.junit.Assert.assertTrue
import org.junit.Test

class ChatStateTest {
    private fun delta(chat: String, seq: Int, text: String) = Inbound.Event(
        chat, seq,
        Json.parseToJsonElement(
            """{"type":"stream_event","event":{"type":"content_block_delta","delta":{"type":"text_delta","text":"$text"}}}""",
        ).jsonObject,
    )

    private fun result(seq: Int, error: Boolean = false) = Inbound.Event(
        "c", seq, Json.parseToJsonElement("""{"type":"result","is_error":$error}""").jsonObject,
    )

    private fun agentText(s: ChatState) = s.items.filterIsInstance<Item.Agent>().last()

    @Test fun deltasAccumulateAndResultCompletes() {
        val s = ChatState().userSent("hi")
            .reduce(delta("c", 1, "Hel")).reduce(delta("c", 2, "lo")).reduce(result(3))
        assertEquals("Hello", agentText(s).text)
        assertEquals(StreamStatus.Complete, agentText(s).status)
        assertEquals(StreamStatus.Complete, s.status)
    }

    @Test fun duplicateAndOldSeqAreDropped() {
        val s = ChatState().userSent("hi")
            .reduce(delta("c", 1, "a")).reduce(delta("c", 2, "b"))
            .reduce(delta("c", 2, "b")).reduce(delta("c", 1, "a")).reduce(delta("c", 3, "c"))
        assertEquals("abc", agentText(s).text)
    }

    @Test fun closeMidReplyIsNotComplete() {
        val s = ChatState().userSent("hi").reduce(delta("c", 1, "part")).reduce(Inbound.Closed("c"))
        assertEquals(StreamStatus.Interrupted, s.status)
        assertEquals(StreamStatus.Interrupted, agentText(s).status)
    }

    @Test fun disconnectMarksInterruptedAndClearsPermissionCards() {
        val s = ChatState().userSent("hi").reduce(delta("c", 1, "part"))
            .reduce(Inbound.Permission("c", "r1", "Bash", "ls", null)).disconnected()
        assertTrue(s.permissions.isEmpty())
        assertEquals(StreamStatus.Interrupted, agentText(s).status)
    }

    @Test fun errorWhileStreamingFailsTheReply() {
        val s = ChatState().userSent("hi").reduce(delta("c", 1, "x")).reduce(Inbound.Error("c", "boom"))
        assertEquals(StreamStatus.Failed, agentText(s).status)
        assertTrue(s.items.any { it is Item.Notice && it.text == "boom" })
    }

    @Test fun errorResultFailsTheReply() {
        val s = ChatState().userSent("hi").reduce(delta("c", 1, "x")).reduce(result(2, error = true))
        assertEquals(StreamStatus.Failed, s.status)
    }

    @Test fun permissionCardLifecycle() {
        val s = ChatState().reduce(Inbound.Permission("c", "r1", "Bash", "list files", null))
        assertEquals(1, s.permissions.size)
        assertEquals("Bash", s.permissions[0].tool)
        assertTrue(s.resolved("r1").permissions.isEmpty())
        assertEquals(1, s.resolved("other").permissions.size)
    }

    @Test fun initCapturesSessionId() {
        val init = Inbound.Event("c", 1, Json.parseToJsonElement("""{"type":"system","subtype":"init","session_id":"abc"}""").jsonObject)
        assertEquals("abc", ChatState().reduce(init).sessionId)
    }

    @Test fun protocolParsesBridgeMessages() {
        val p = Protocol.parse("""{"type":"permission","chat":"c","request":"r","tool":"Write","description":"d","input":{"file_path":"a"}}""")
        assertTrue(p is Inbound.Permission)
        assertEquals("r", (p as Inbound.Permission).request)
        assertNotNull(Protocol.parse("""{"type":"event","chat":"c","seq":4,"raw":{"type":"result"}}"""))
        assertNull(Protocol.parse("not json"))
        assertNull(Protocol.parse("""{"type":"event","chat":"c"}"""))
        assertNull(Protocol.parse("""{"type":"unknown"}"""))
    }

    @Test fun outgoingMessagesMatchTheDocumentedProtocol() {
        val d = Json.parseToJsonElement(Protocol.decide("r1", true)).jsonObject
        assertEquals("decide", d["op"]!!.toString().trim('"'))
        assertEquals("r1", d["request"]!!.toString().trim('"'))
        assertEquals("true", d["allow"]!!.toString())
        val s = Json.parseToJsonElement(Protocol.start("c1", "claude", "/w", mode = "manual")).jsonObject
        assertEquals("manual", s["mode"]!!.toString().trim('"'))
        assertNull(s["model"])
    }

    @Test fun gapInSequenceNeverShowsAsComplete() {
        val s = ChatState().userSent("hi").reduce(delta("c", 1, "a")).reduce(result(3))
        assertEquals(StreamStatus.Interrupted, s.status)
        assertTrue(s.lostMessages)
    }

    @Test fun closeClearsPermissionCards() {
        val s = ChatState().reduce(Inbound.Permission("c", "r", "Bash", "", null)).reduce(Inbound.Closed("c"))
        assertTrue(s.permissions.isEmpty())
    }

    @Test fun malformedEnvelopesDoNotThrow() {
        assertNull(Protocol.parse("""{"type":"event","chat":"c","seq":"x","raw":{}}"""))
        assertNull(Protocol.parse("""{"type":"event","chat":"c","seq":{},"raw":{}}"""))
        assertNull(Protocol.parse("""{"type":"permission","chat":{},"request":"r","tool":"t"}"""))
        val odd = Inbound.Event("c", 1, Json.parseToJsonElement("""{"type":{"a":1},"event":{"type":["x"]}}""").jsonObject)
        ChatState().reduce(odd)
    }

    @Test fun bridgeErrorFailsTheReply() {
        val e = Inbound.Event("c", 2, Json.parseToJsonElement("""{"type":"bridge_error","message":"too long"}""").jsonObject)
        val s = ChatState().userSent("hi").reduce(delta("c", 1, "x")).reduce(e)
        assertEquals(StreamStatus.Failed, s.status)
    }
}
