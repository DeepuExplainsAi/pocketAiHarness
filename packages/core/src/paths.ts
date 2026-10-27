import { realpathSync } from "node:fs";
import { dirname, isAbsolute, relative, resolve, sep } from "node:path";

export class PathEscapeError extends Error {
  readonly root: string;
  readonly candidate: string;

  constructor(root: string, candidate: string) {
    super(`Path "${candidate}" is outside the workspace "${root}"`);
    this.name = "PathEscapeError";
    this.root = root;
    this.candidate = candidate;
  }
}

function realOfNearestAncestor(path: string): string {
  let current = path;
  for (;;) {
    try {
      const real = realpathSync(current);
      return resolve(real, relative(current, path));
    } catch {
      const parent = dirname(current);
      if (parent === current) return path;
      current = parent;
    }
  }
}

/**
 * Resolves `candidate` against `root` and guarantees the result stays inside it,
 * including after symlinks on any existing ancestor are followed.
 */
export function resolveInside(root: string, candidate: string): string {
  const realRoot = realpathSync(root);
  const target = realOfNearestAncestor(resolve(realRoot, candidate));
  const rel = relative(realRoot, target);
  if (rel === ".." || rel.startsWith(`..${sep}`) || isAbsolute(rel)) {
    throw new PathEscapeError(root, candidate);
  }
  return target;
}
