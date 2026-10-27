import assert from "node:assert/strict";
import test from "node:test";
import { EditorBuffer, type Disk } from "../src/editor-buffer.ts";

const memoryDisk = (initial: string | null) => {
  const files = new Map<string, string>();
  if (initial !== null) files.set("a.ts", initial);
  const disk: Disk = {
    read: async (p) => files.get(p) ?? null,
    write: async (p, c) => void files.set(p, c),
  };
  return { disk, files };
};

test("dirty tracks difference from disk and clears after save", async () => {
  const { disk, files } = memoryDisk("one");
  const buf = new EditorBuffer("a.ts", "one", "one");
  assert.equal(buf.dirty, false);
  buf.content = "two";
  assert.equal(buf.dirty, true);
  assert.deepEqual(await buf.save(disk), { status: "saved" });
  assert.equal(files.get("a.ts"), "two");
  assert.equal(buf.dirty, false);
});

test("external change produces a conflict and writes nothing", async () => {
  const { disk, files } = memoryDisk("one");
  const buf = new EditorBuffer("a.ts", "one", "one");
  buf.content = "mine";
  files.set("a.ts", "theirs");
  const result = await buf.save(disk);
  assert.deepEqual(result, { status: "conflict", diskContent: "theirs" });
  assert.equal(files.get("a.ts"), "theirs");
  assert.equal(buf.content, "mine");
});

test("overwrite keeps local edits; reload keeps disk version", async () => {
  const a = memoryDisk("one");
  const bufA = new EditorBuffer("a.ts", "one", "one");
  bufA.content = "mine";
  a.files.set("a.ts", "theirs");
  await bufA.overwrite(a.disk);
  assert.equal(a.files.get("a.ts"), "mine");

  const bufB = new EditorBuffer("a.ts", "mine", "one");
  bufB.reload("theirs");
  assert.equal(bufB.content, "theirs");
  assert.equal(bufB.dirty, false);
});

test("file deleted on disk is a conflict; new file saves cleanly", async () => {
  const gone = memoryDisk(null);
  const buf = new EditorBuffer("a.ts", "x", "x");
  assert.equal((await buf.save(gone.disk)).status, "conflict");

  const fresh = memoryDisk(null);
  const created = new EditorBuffer("a.ts", "x", null);
  assert.equal((await created.save(fresh.disk)).status, "saved");
});
