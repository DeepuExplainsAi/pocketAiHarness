import assert from "node:assert/strict";
import test from "node:test";
import { ProviderRegistry, type AuthStatus } from "../src/accounts.ts";

const adapter = (id: string, status: AuthStatus | Error) => ({
  id,
  status: async () => {
    if (status instanceof Error) throw status;
    return status;
  },
});

test("no provider is enabled by default", async () => {
  const r = new ProviderRegistry([adapter("a", "authenticated")]);
  assert.deepEqual(await r.snapshot(), [{ id: "a", enabled: false, status: "authenticated" }]);
  assert.equal(await r.canSend("a"), false);
});

test("sending requires enabled and authenticated", async () => {
  const r = new ProviderRegistry([adapter("ok", "authenticated"), adapter("old", "expired")]);
  r.enable("ok");
  r.enable("old");
  assert.equal(await r.canSend("ok"), true);
  assert.equal(await r.canSend("old"), false);
});

test("provider error is reported as unavailable, not as success", async () => {
  const r = new ProviderRegistry([adapter("x", new Error("down"))]);
  r.enable("x");
  assert.equal((await r.snapshot())[0]?.status, "unavailable");
  assert.equal(await r.canSend("x"), false);
});

test("disable keeps the provider registered; unknown id throws", async () => {
  const r = new ProviderRegistry([adapter("a", "authenticated")]);
  r.enable("a");
  r.disable("a");
  assert.equal((await r.snapshot()).length, 1);
  assert.throws(() => r.enable("nope"));
});
