import assert from "node:assert/strict";
import test from "node:test";
import { ChatStream } from "../src/chat-stream.ts";

test("accumulates in order and completes", () => {
  const s = new ChatStream();
  s.start();
  s.push({ seq: 0, text: "Hel" });
  s.push({ seq: 1, text: "lo" });
  s.complete();
  assert.equal(s.text, "Hello");
  assert.equal(s.state, "complete");
  assert.equal(s.isIncomplete, false);
});

test("drops duplicate and out-of-order events after a reconnect", () => {
  const s = new ChatStream();
  s.start();
  s.push({ seq: 0, text: "a" });
  s.push({ seq: 1, text: "b" });
  s.interrupt();
  assert.equal(s.push({ seq: 1, text: "b" }), false);
  assert.equal(s.push({ seq: 0, text: "a" }), false);
  assert.equal(s.push({ seq: 2, text: "c" }), true);
  assert.equal(s.text, "abc");
  assert.equal(s.state, "streaming");
});

test("interrupted stream is never reported as complete", () => {
  const s = new ChatStream();
  s.start();
  s.push({ seq: 0, text: "partial" });
  s.interrupt();
  s.complete();
  assert.equal(s.state, "interrupted");
  assert.equal(s.isIncomplete, true);
});

test("failure keeps partial text and the error", () => {
  const s = new ChatStream();
  s.start();
  s.push({ seq: 0, text: "x" });
  s.fail("provider unavailable");
  assert.equal(s.text, "x");
  assert.equal(s.error, "provider unavailable");
  assert.equal(s.isIncomplete, true);
});

test("ignores events when not streaming", () => {
  const s = new ChatStream();
  assert.equal(s.push({ seq: 0, text: "x" }), false);
});
