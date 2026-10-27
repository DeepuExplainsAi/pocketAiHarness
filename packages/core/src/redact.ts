const PATTERNS: readonly RegExp[] = [
  /\b(gh[pousr]_[A-Za-z0-9]{20,})\b/g,
  /\b(sk-[A-Za-z0-9_-]{20,})\b/g,
  /\bAKIA[0-9A-Z]{16}\b/g,
  /\bBearer\s+[A-Za-z0-9._~+/=-]{16,}/gi,
  /\beyJ[A-Za-z0-9_-]{8,}\.[A-Za-z0-9_-]{8,}\.[A-Za-z0-9_-]{8,}\b/g,
  /-----BEGIN [A-Z ]*PRIVATE KEY-----[\s\S]*?-----END [A-Z ]*PRIVATE KEY-----/g,
];

const KEY_VALUE = /\b([A-Za-z0-9_]*(?:token|secret|password|passwd|api[_-]?key)[A-Za-z0-9_]*)\s*([=:])\s*("[^"]*"|'[^']*'|\S+)/gi;

export const REDACTED = "[REDACTED]";

/** Applied to every string before it reaches logs, crash reports, exports or prompts. */
export function redact(text: string): string {
  let out = text.replace(KEY_VALUE, (_m, key: string, sep: string) => `${key}${sep}${REDACTED}`);
  for (const pattern of PATTERNS) out = out.replace(pattern, REDACTED);
  return out;
}
