import { resolveInside } from "./paths.ts";

export type GitOperation =
  | { type: "status" }
  | { type: "diff"; staged: boolean; path?: string }
  | { type: "stage"; paths: string[] }
  | { type: "unstage"; paths: string[] }
  | { type: "commit"; message: string }
  | { type: "discard"; paths: string[]; confirmed: boolean }
  | { type: "checkout"; branch: string; create: boolean }
  | { type: "push" | "pull"; confirmed: boolean };

export class GitOperationError extends Error {}

const BRANCH = /^(?![-/])(?!.*\.\.)(?!.*\/\/)(?!.*[\s~^:?*[\\])[A-Za-z0-9._/-]+(?<!\.lock)(?<![/.])$/;

/**
 * Builds an argv array for the git executable. No shell string is ever composed,
 * all paths are confined to the repository root, and destructive operations
 * require an explicit confirmation flag from the caller.
 */
export function buildGitArgs(repoRoot: string, op: GitOperation): string[] {
  const paths = (list: string[]): string[] => {
    if (list.length === 0) throw new GitOperationError("Select at least one path");
    return list.map((p) => resolveInside(repoRoot, p));
  };

  switch (op.type) {
    case "status":
      return ["status", "--porcelain=v2", "--branch", "-z"];
    case "diff":
      return ["diff", ...(op.staged ? ["--cached"] : []), ...(op.path ? ["--", resolveInside(repoRoot, op.path)] : [])];
    case "stage":
      return ["add", "--", ...paths(op.paths)];
    case "unstage":
      return ["restore", "--staged", "--", ...paths(op.paths)];
    case "commit":
      if (!op.message.trim()) throw new GitOperationError("Commit message is empty");
      return ["commit", "-m", op.message];
    case "discard":
      if (!op.confirmed) throw new GitOperationError("Discarding changes requires confirmation");
      return ["restore", "--", ...paths(op.paths)];
    case "checkout":
      if (!BRANCH.test(op.branch)) throw new GitOperationError(`Invalid branch name: ${op.branch}`);
      return op.create ? ["checkout", "-b", op.branch] : ["checkout", op.branch, "--"];
    case "push":
    case "pull":
      if (!op.confirmed) throw new GitOperationError(`${op.type} requires confirmation`);
      return [op.type];
  }
}
