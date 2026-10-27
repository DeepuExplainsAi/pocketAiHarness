import assert from "node:assert/strict";
import test from "node:test";
import { PermissionBroker, type Decision, type PermissionRequest } from "../src/permissions.ts";

const req: PermissionRequest = { kind: "process.exec", target: "npm test", workspace: "/ws", effect: "Runs tests" };
const broker = (decision: Decision, now = () => 1_000) => {
  let prompts = 0;
  const b = new PermissionBroker(async () => (prompts++, decision), now);
  return { b, prompts: () => prompts };
};

test("denies by default and on explicit deny", async () => {
  const { b } = broker({ type: "deny" });
  assert.equal(await b.check(req), false);
});

test("allow-once prompts every time", async () => {
  const { b, prompts } = broker({ type: "allow-once" });
  assert.equal(await b.check(req), true);
  assert.equal(await b.check(req), true);
  assert.equal(prompts(), 2);
});

test("scoped grant is reused until it expires", async () => {
  let t = 1_000;
  const { b, prompts } = broker({ type: "allow-scoped", expiresAt: 2_000 }, () => t);
  await b.check(req);
  await b.check(req);
  assert.equal(prompts(), 1);
  t = 2_000;
  await b.check(req);
  assert.equal(prompts(), 2);
});

test("grant does not leak to another target, kind or workspace", async () => {
  const { b, prompts } = broker({ type: "allow-always", confirmedPersistent: true });
  await b.check(req);
  await b.check({ ...req, target: "rm -rf /" });
  await b.check({ ...req, kind: "fs.delete" });
  await b.check({ ...req, workspace: "/other" });
  assert.equal(prompts(), 4);
});

test("already-expired scoped grant is refused", async () => {
  const { b } = broker({ type: "allow-scoped", expiresAt: 500 });
  assert.equal(await b.check(req), false);
});

test("revokeAll clears grants", async () => {
  const { b, prompts } = broker({ type: "allow-always", confirmedPersistent: true });
  await b.check(req);
  b.revokeAll();
  await b.check(req);
  assert.equal(prompts(), 2);
});
