export type AuthStatus = "disconnected" | "pending" | "authenticated" | "expired" | "unavailable";

export interface ProviderAdapter {
  readonly id: string;
  /** Status reported by the provider's official client, never inferred by the UI. */
  status(): Promise<AuthStatus>;
}

export interface ProviderSnapshot {
  id: string;
  enabled: boolean;
  status: AuthStatus;
}

/** Providers are opt-in. Enabling or disabling never signs in, signs out or deletes data. */
export class ProviderRegistry {
  private enabled = new Set<string>();
  private readonly adapters: readonly ProviderAdapter[];

  constructor(adapters: readonly ProviderAdapter[]) {
    this.adapters = adapters;
  }

  enable(id: string): void {
    this.require(id);
    this.enabled.add(id);
  }

  disable(id: string): void {
    this.enabled.delete(id);
  }

  async snapshot(): Promise<ProviderSnapshot[]> {
    return Promise.all(
      this.adapters.map(async (a) => ({
        id: a.id,
        enabled: this.enabled.has(a.id),
        status: await a.status().catch((): AuthStatus => "unavailable"),
      })),
    );
  }

  /** A message may only be sent when the provider is enabled and authenticated. */
  async canSend(id: string): Promise<boolean> {
    const adapter = this.require(id);
    return this.enabled.has(id) && (await adapter.status().catch(() => "unavailable")) === "authenticated";
  }

  private require(id: string): ProviderAdapter {
    const adapter = this.adapters.find((a) => a.id === id);
    if (!adapter) throw new Error(`Unknown provider: ${id}`);
    return adapter;
  }
}
