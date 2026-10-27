export type StreamState = "idle" | "streaming" | "interrupted" | "complete" | "failed";

export interface StreamEvent {
  /** Monotonic sequence assigned by the provider adapter. */
  seq: number;
  text: string;
}

/**
 * Accumulates a single agent reply. Duplicate or out-of-order events (reconnects)
 * are dropped, and a broken stream is never reported as complete.
 */
export class ChatStream {
  state: StreamState = "idle";
  text = "";
  error?: string;
  private lastSeq = -1;

  start(): void {
    this.state = "streaming";
  }

  push(event: StreamEvent): boolean {
    if (this.state !== "streaming" && this.state !== "interrupted") return false;
    if (event.seq <= this.lastSeq) return false;
    this.lastSeq = event.seq;
    this.text += event.text;
    this.state = "streaming";
    return true;
  }

  /** Connection dropped. Output so far is kept but marked incomplete. */
  interrupt(): void {
    if (this.state === "streaming") this.state = "interrupted";
  }

  complete(): void {
    if (this.state === "streaming") this.state = "complete";
  }

  fail(message: string): void {
    this.state = "failed";
    this.error = message;
  }

  get isIncomplete(): boolean {
    return this.state === "interrupted" || this.state === "failed";
  }
}
