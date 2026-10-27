import { createHash } from "node:crypto";

export const hashContent = (content: string): string => createHash("sha256").update(content).digest("hex");

export interface Disk {
  read(path: string): Promise<string | null>;
  write(path: string, content: string): Promise<void>;
}

export type SaveResult = { status: "saved" } | { status: "conflict"; diskContent: string | null };

/** Tracks in-memory content separately from the last version known to be on disk. */
export class EditorBuffer {
  readonly path: string;
  content: string;
  private knownDiskHash: string | null;

  constructor(path: string, content: string, diskContent: string | null) {
    this.path = path;
    this.content = content;
    this.knownDiskHash = diskContent === null ? null : hashContent(diskContent);
  }

  get dirty(): boolean {
    return this.knownDiskHash !== hashContent(this.content);
  }

  async save(disk: Disk): Promise<SaveResult> {
    const current = await disk.read(this.path);
    const currentHash = current === null ? null : hashContent(current);
    if (currentHash !== this.knownDiskHash) return { status: "conflict", diskContent: current };
    await this.commit(disk);
    return { status: "saved" };
  }

  /** User explicitly chose to replace the newer disk version. */
  async overwrite(disk: Disk): Promise<void> {
    await this.commit(disk);
  }

  /** User explicitly chose to discard local edits in favour of disk. */
  reload(diskContent: string | null): void {
    this.content = diskContent ?? "";
    this.knownDiskHash = diskContent === null ? null : hashContent(diskContent);
  }

  private async commit(disk: Disk): Promise<void> {
    await disk.write(this.path, this.content);
    this.knownDiskHash = hashContent(this.content);
  }
}
