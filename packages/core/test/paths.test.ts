import assert from "node:assert/strict";
import { mkdirSync, mkdtempSync, symlinkSync, writeFileSync } from "node:fs";
import { tmpdir } from "node:os";
import { join } from "node:path";
import test from "node:test";
import { PathEscapeError, resolveInside } from "../src/paths.ts";

const make = () => {
  const base = mkdtempSync(join(tmpdir(), "pah-"));
  const root = join(base, "ws");
  mkdirSync(join(root, "src"), { recursive: true });
  mkdirSync(join(base, "outside"));
  writeFileSync(join(base, "outside", "secret.txt"), "x");
  return { base, root };
};

test("accepts paths inside the workspace, including not-yet-existing ones", () => {
  const { root } = make();
  assert.ok(resolveInside(root, "src/new/file.ts").endsWith(join("src", "new", "file.ts")));
});

test("rejects parent traversal", () => {
  const { root } = make();
  assert.throws(() => resolveInside(root, "../outside/secret.txt"), PathEscapeError);
});

test("rejects absolute paths outside the workspace", () => {
  const { root, base } = make();
  assert.throws(() => resolveInside(root, join(base, "outside", "secret.txt")), PathEscapeError);
});

test("rejects symlinks that point outside the workspace", () => {
  const { root, base } = make();
  symlinkSync(join(base, "outside"), join(root, "link"));
  assert.throws(() => resolveInside(root, "link/secret.txt"), PathEscapeError);
  assert.throws(() => resolveInside(root, "link/new.txt"), PathEscapeError);
});

test("rejects sibling directory sharing the root name prefix", () => {
  const { root, base } = make();
  mkdirSync(join(base, "ws-evil"));
  assert.throws(() => resolveInside(root, "../ws-evil/a"), PathEscapeError);
});
