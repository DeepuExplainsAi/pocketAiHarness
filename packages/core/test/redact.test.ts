import assert from "node:assert/strict";
import test from "node:test";
import { REDACTED, redact } from "../src/redact.ts";

test("redacts common token formats", () => {
  const github = "ghp_" + "a".repeat(36);
  const bearer = "Authorization: Bearer abcdefghijklmnop1234";
  assert.equal(redact(`token ${github} end`), `token ${REDACTED} end`);
  assert.ok(!redact(bearer).includes("abcdefghijklmnop1234"));
  assert.ok(!redact("AKIAABCDEFGHIJKLMNOP").includes("AKIA"));
});

test("redacts key=value secrets and keeps the key name", () => {
  assert.equal(redact("API_KEY=hunter2"), `API_KEY=${REDACTED}`);
  assert.equal(redact('password: "my pass"'), `password:${REDACTED}`);
});

test("redacts private key blocks", () => {
  const pem = "-----BEGIN PRIVATE KEY-----\nMIIabc\n-----END PRIVATE KEY-----";
  assert.equal(redact(`k=${pem}`), `k=${REDACTED}`);
});

test("leaves ordinary text unchanged", () => {
  const text = "Opened src/index.ts in 12 ms";
  assert.equal(redact(text), text);
});
