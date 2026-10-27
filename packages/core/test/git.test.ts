import assert from "node:assert/strict";
import { mkdirSync, mkdtempSync } from "node:fs";
import { tmpdir } from "node:os";
import { join } from "node:path";
import test from "node:test";
import { buildGitArgs, GitOperationError } from "../src/git.ts";
import { PathEscapeError } from "../src/paths.ts";

const repo = mkdtempSync(join(tmpdir(), "pah-git-"));
mkdirSync(join(repo, "src"));

test("stage builds an argv with absolute in-repo paths after --", () => {
  const args = buildGitArgs(repo, { type: "stage", paths: ["src/a.ts"] });
  assert.equal(args[0], "add");
  assert.equal(args[1], "--");
  assert.ok(args[2]?.endsWith(join("src", "a.ts")));
});

test("paths outside the repository are refused", () => {
  assert.throws(() => buildGitArgs(repo, { type: "stage", paths: ["../x"] }), PathEscapeError);
});

test("shell metacharacters stay inside a single argument", () => {
  const args = buildGitArgs(repo, { type: "commit", message: "fix; rm -rf / $(id)" });
  assert.deepEqual(args, ["commit", "-m", "fix; rm -rf / $(id)"]);
});

test("destructive and remote operations need confirmation", () => {
  assert.throws(() => buildGitArgs(repo, { type: "discard", paths: ["a"], confirmed: false }), GitOperationError);
  assert.throws(() => buildGitArgs(repo, { type: "push", confirmed: false }), GitOperationError);
  assert.throws(() => buildGitArgs(repo, { type: "pull", confirmed: false }), GitOperationError);
  assert.deepEqual(buildGitArgs(repo, { type: "push", confirmed: true }), ["push"]);
});

test("empty selection and empty commit message are refused", () => {
  assert.throws(() => buildGitArgs(repo, { type: "stage", paths: [] }), GitOperationError);
  assert.throws(() => buildGitArgs(repo, { type: "commit", message: "  " }), GitOperationError);
});

test("branch names that could act as options or revisions are refused", () => {
  for (const branch of ["--force", "a..b", "a b", "x~1", "a.lock", "/a", "a/"]) {
    assert.throws(() => buildGitArgs(repo, { type: "checkout", branch, create: false }), GitOperationError, branch);
  }
  assert.deepEqual(buildGitArgs(repo, { type: "checkout", branch: "feature/x-1", create: true }), ["checkout", "-b", "feature/x-1"]);
});
