import assert from "node:assert/strict";
import test from "node:test";
import { validateArchive, type ArchiveEntry } from "../src/archive.ts";

const limits = { maxEntries: 100, maxTotalBytes: 1_000, freeBytes: 10_000 };
const file = (name: string, size = 1): ArchiveEntry => ({ name, size, kind: "file" });
const codes = (entries: ArchiveEntry[], l = limits) => validateArchive(entries, l).map((i) => i.code);

test("clean archive passes", () => {
  assert.deepEqual(codes([file("a.txt"), file("dir/b.txt")]), []);
});

test("blocks path traversal, absolute and drive paths, backslashes", () => {
  for (const name of ["../x", "a/../../x", "/etc/passwd", "C:/x", "a\\b", "a\0b"]) {
    assert.deepEqual(codes([file(name)]), ["unsafe-path"], name);
  }
});

test("blocks symlinks and special entries", () => {
  assert.deepEqual(codes([{ name: "l", size: 0, kind: "symlink" }]), ["symlink"]);
  assert.deepEqual(codes([{ name: "d", size: 0, kind: "other" }]), ["unsupported-type"]);
});

test("blocks duplicates including case and unicode-normalisation collisions", () => {
  assert.deepEqual(codes([file("a.txt"), file("A.TXT")]), ["duplicate"]);
  assert.deepEqual(codes([file("\u00e9.txt"), file("e\u0301.txt")]), ["duplicate"]);
});

test("enforces entry count, total size and free disk", () => {
  assert.deepEqual(codes([file("a"), file("b")], { ...limits, maxEntries: 1 }), ["too-many-entries"]);
  assert.deepEqual(codes([file("a", 2_000)], { ...limits, freeBytes: 99_999 }), ["too-large"]);
  assert.deepEqual(codes([file("a", 500)], { ...limits, freeBytes: 100 }), ["insufficient-disk"]);
});
