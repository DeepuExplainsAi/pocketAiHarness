export type PermissionKind =
  | "fs.read"
  | "fs.write"
  | "fs.delete"
  | "process.exec"
  | "network"
  | "browser.action"
  | "git.write"
  | "mcp.tool"
  | "attachment.send";

export interface PermissionRequest {
  kind: PermissionKind;
  /** Concrete target: path, command line, URL, tool name. */
  target: string;
  /** Workspace the request is confined to. */
  workspace: string;
  /** Human-readable effect shown to the user before deciding. */
  effect: string;
}

export type Decision =
  | { type: "deny"; reason?: string }
  | { type: "allow-once" }
  | { type: "allow-scoped"; expiresAt: number }
  | { type: "allow-always"; confirmedPersistent: true };

export type Prompter = (request: PermissionRequest) => Promise<Decision>;

interface Grant {
  kind: PermissionKind;
  target: string;
  workspace: string;
  expiresAt: number | null;
}

/**
 * Default-deny broker. Agent output is never consent: only a Prompter
 * (the user interface) can produce an allow decision.
 */
export class PermissionBroker {
  private grants: Grant[] = [];
  private readonly prompter: Prompter;
  private readonly now: () => number;

  constructor(prompter: Prompter, now: () => number = Date.now) {
    this.prompter = prompter;
    this.now = now;
  }

  async check(request: PermissionRequest): Promise<boolean> {
    if (this.hasGrant(request)) return true;

    const decision = await this.prompter(request);
    switch (decision.type) {
      case "allow-once":
        return true;
      case "allow-scoped":
        if (decision.expiresAt <= this.now()) return false;
        this.addGrant(request, decision.expiresAt);
        return true;
      case "allow-always":
        if (decision.confirmedPersistent !== true) return false;
        this.addGrant(request, null);
        return true;
      default:
        return false;
    }
  }

  revokeAll(workspace?: string): void {
    this.grants = workspace ? this.grants.filter((g) => g.workspace !== workspace) : [];
  }

  list(): readonly Readonly<Grant>[] {
    this.prune();
    return this.grants;
  }

  private hasGrant(r: PermissionRequest): boolean {
    this.prune();
    return this.grants.some((g) => g.kind === r.kind && g.target === r.target && g.workspace === r.workspace);
  }

  private addGrant(r: PermissionRequest, expiresAt: number | null): void {
    this.grants.push({ kind: r.kind, target: r.target, workspace: r.workspace, expiresAt });
  }

  private prune(): void {
    const t = this.now();
    this.grants = this.grants.filter((g) => g.expiresAt === null || g.expiresAt > t);
  }
}
