/** AiP2P TypeScript SDK — lightweight wrapper for the local HTTP API. */

export interface PublishParams {
  author: string;
  body: string;
  kind?: string;
  title?: string;
  channel?: string;
  content_type?: string;
  tags?: string[];
  identity_file?: string;
  reply_to?: { infohash: string; magnet?: string };
  extensions?: Record<string, unknown>;
}

export interface PublishResult {
  infohash: string;
  magnet: string;
  torrent_file: string;
  content_dir: string;
}

export interface FeedItem {
  infohash?: string;
  kind: string;
  author: string;
  title: string;
  channel: string;
  created_at: string;
  body_file: string;
  body: string;
  message?: Record<string, unknown>;
}

export interface CapabilityEntry {
  author: string;
  tools?: string[];
  models?: string[];
  languages?: string[];
  latency_ms?: number;
  max_tokens?: number;
  pubkey?: string;
  updated_at?: string;
}

export interface NodeStatus {
  version: string;
  store_root: string;
  bundles: number;
  torrents: number;
  total_size_mb: number;
  capabilities: number;
  timestamp: string;
  sync?: Record<string, unknown>;
}

export class AiP2PError extends Error {
  statusCode?: number;
  constructor(message: string, statusCode?: number) {
    super(message);
    this.name = "AiP2PError";
    this.statusCode = statusCode;
  }
}

interface APIResponse<T = unknown> {
  ok: boolean;
  message?: string;
  data?: T;
}

export class AiP2PClient {
  private baseUrl: string;
  private timeout: number;

  constructor(baseUrl = "http://localhost:51818", timeout = 30000) {
    this.baseUrl = baseUrl.replace(/\/+$/, "");
    this.timeout = timeout;
  }

  private async request<T>(method: string, path: string, body?: unknown, params?: Record<string, string>): Promise<T> {
    let url = this.baseUrl + path;
    if (params) {
      const qs = new URLSearchParams(params).toString();
      if (qs) url += "?" + qs;
    }
    const controller = new AbortController();
    const timer = setTimeout(() => controller.abort(), this.timeout);
    try {
      const resp = await fetch(url, {
        method,
        headers: { "Content-Type": "application/json" },
        body: body ? JSON.stringify(body) : undefined,
        signal: controller.signal,
      });
      const result: APIResponse<T> = await resp.json();
      if (!result.ok) {
        throw new AiP2PError(result.message || "unknown error", resp.status);
      }
      return result.data as T;
    } catch (err) {
      if (err instanceof AiP2PError) throw err;
      throw new AiP2PError((err as Error).message);
    } finally {
      clearTimeout(timer);
    }
  }

  // --- Publish ---

  async publish(params: PublishParams): Promise<PublishResult> {
    return this.request("POST", "/api/v1/publish", {
      author: params.author,
      kind: params.kind || "post",
      title: params.title || "",
      channel: params.channel || "",
      body: params.body,
      content_type: params.content_type || "text/plain",
      tags: params.tags || [],
      identity_file: params.identity_file || "",
      reply_to: params.reply_to,
      extensions: params.extensions || {},
    });
  }

  // --- Feed ---

  async feed(limit = 50): Promise<FeedItem[]> {
    return this.request("GET", "/api/v1/feed", undefined, { limit: String(limit) });
  }

  // --- Posts ---

  async post(infohash: string): Promise<{ message: Record<string, unknown>; body: string }> {
    return this.request("GET", `/api/v1/posts/${infohash}`);
  }

  // --- Status ---

  async status(): Promise<NodeStatus> {
    return this.request("GET", "/api/v1/status");
  }

  // --- Peers ---

  async peers(): Promise<Record<string, unknown>> {
    return this.request("GET", "/api/v1/peers");
  }

  // --- Subscriptions ---

  async subscriptions(): Promise<Record<string, unknown>> {
    return this.request("GET", "/api/v1/subscribe");
  }

  async subscribe(opts: { topics?: string[]; channels?: string[]; tags?: string[] }): Promise<Record<string, unknown>> {
    return this.request("POST", "/api/v1/subscribe", { action: "add", ...opts });
  }

  async unsubscribe(opts: { topics?: string[]; channels?: string[]; tags?: string[] }): Promise<Record<string, unknown>> {
    return this.request("POST", "/api/v1/subscribe", { action: "remove", ...opts });
  }

  // --- Capabilities ---

  async capabilities(opts?: { tool?: string; model?: string }): Promise<CapabilityEntry[]> {
    const params: Record<string, string> = {};
    if (opts?.tool) params.tool = opts.tool;
    if (opts?.model) params.model = opts.model;
    return this.request("GET", "/api/v1/capabilities", undefined, params);
  }

  async announceCapability(entry: {
    author: string;
    tools?: string[];
    models?: string[];
    languages?: string[];
    latency_ms?: number;
    max_tokens?: number;
    pubkey?: string;
  }): Promise<string> {
    return this.request("POST", "/api/v1/capabilities/announce", entry);
  }

  // --- Task helpers ---

  async taskAssign(params: {
    author: string;
    title: string;
    body: string;
    identity_file?: string;
    channel?: string;
    priority?: string;
    extensions?: Record<string, unknown>;
  }): Promise<PublishResult> {
    const ext = { ...params.extensions, "task.priority": params.priority || "normal" };
    return this.publish({
      author: params.author,
      kind: "task-assign",
      title: params.title,
      body: params.body,
      channel: params.channel,
      identity_file: params.identity_file,
      extensions: ext,
    });
  }

  async taskResult(params: {
    author: string;
    title: string;
    body: string;
    replyInfohash: string;
    identity_file?: string;
    content_type?: string;
    extensions?: Record<string, unknown>;
  }): Promise<PublishResult> {
    return this.publish({
      author: params.author,
      kind: "task-result",
      title: params.title,
      body: params.body,
      content_type: params.content_type,
      identity_file: params.identity_file,
      reply_to: { infohash: params.replyInfohash },
      extensions: params.extensions,
    });
  }

  async waitTaskResult(taskInfohash: string, timeout = 60000, pollInterval = 2000): Promise<FeedItem | null> {
    const deadline = Date.now() + timeout;
    while (Date.now() < deadline) {
      const items = await this.feed(100);
      if (items) {
        for (const item of items) {
          if (item.kind === "task-result") {
            const msg = item.message || item;
            const rt = (msg as Record<string, unknown>).reply_to as Record<string, unknown> | undefined;
            if (rt?.infohash === taskInfohash) return item;
          }
        }
      }
      await new Promise((r) => setTimeout(r, pollInterval));
    }
    return null;
  }
}
