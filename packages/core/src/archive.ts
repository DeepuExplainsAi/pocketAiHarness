export interface ArchiveEntry {
  name: string;
  size: number;
  kind: "file" | "directory" | "symlink" | "other";
}

export interface ArchiveLimits {
  maxEntries: number;
  maxTotalBytes: number;
  freeBytes: number;
}

export type ArchiveIssue =
  | { code: "unsafe-path"; entry: string }
  | { code: "symlink"; entry: string }
  | { code: "unsupported-type"; entry: string }
  | { code: "duplicate"; entry: string }
  | { code: "too-many-entries"; count: number }
  | { code: "too-large"; bytes: number }
  | { code: "insufficient-disk"; needed: number; free: number };

const WINDOWS_DRIVE = /^[a-zA-Z]:/;

export function isUnsafeEntryName(name: string): boolean {
  if (name === "" || name.includes("\0") || name.includes("\\")) return true;
  if (name.startsWith("/") || WINDOWS_DRIVE.test(name)) return true;
  return name.split("/").some((part) => part === "..");
}

/** Pure pre-extraction check. Extraction must not start while the result is non-empty. */
export function validateArchive(entries: readonly ArchiveEntry[], limits: ArchiveLimits): ArchiveIssue[] {
  const issues: ArchiveIssue[] = [];
  const seen = new Set<string>();
  let total = 0;

  if (entries.length > limits.maxEntries) issues.push({ code: "too-many-entries", count: entries.length });

  for (const entry of entries) {
    if (isUnsafeEntryName(entry.name)) issues.push({ code: "unsafe-path", entry: entry.name });
    if (entry.kind === "symlink") issues.push({ code: "symlink", entry: entry.name });
    if (entry.kind === "other") issues.push({ code: "unsupported-type", entry: entry.name });

    const key = entry.name.normalize("NFC").toLowerCase();
    if (seen.has(key)) issues.push({ code: "duplicate", entry: entry.name });
    seen.add(key);

    total += entry.size;
  }

  if (total > limits.maxTotalBytes) issues.push({ code: "too-large", bytes: total });
  if (total > limits.freeBytes) issues.push({ code: "insufficient-disk", needed: total, free: limits.freeBytes });
  return issues;
}
