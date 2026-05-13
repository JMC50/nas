import type { Notification, NotificationKind } from "$lib/types";

const DEFAULT_DURATION_MS = 5000;

function randomId() {
  return crypto.randomUUID();
}

class NotificationStore {
  queue = $state<Notification[]>([]);
  private timers = new Map<string, ReturnType<typeof setTimeout>>();

  notify(input: {
    kind: NotificationKind;
    message: string;
    durationMs?: number;
  }) {
    const item: Notification = {
      id: randomId(),
      kind: input.kind,
      message: input.message,
      durationMs: input.durationMs ?? DEFAULT_DURATION_MS,
      createdAt: Date.now(),
    };
    this.queue = [...this.queue, item];
    if (item.durationMs > 0) {
      const handle = setTimeout(() => this.dismiss(item.id), item.durationMs);
      this.timers.set(item.id, handle);
    }
    return item.id;
  }

  info(message: string, durationMs?: number) {
    return this.notify({ kind: "info", message, durationMs });
  }

  success(message: string, durationMs?: number) {
    return this.notify({ kind: "success", message, durationMs });
  }

  warning(message: string, durationMs?: number) {
    return this.notify({ kind: "warning", message, durationMs });
  }

  error(message: string, durationMs?: number) {
    return this.notify({ kind: "error", message, durationMs });
  }

  dismiss(id: string) {
    const handle = this.timers.get(id);
    if (handle !== undefined) {
      clearTimeout(handle);
      this.timers.delete(id);
    }
    this.queue = this.queue.filter((item) => item.id !== id);
  }

  clear() {
    for (const handle of this.timers.values()) clearTimeout(handle);
    this.timers.clear();
    this.queue = [];
  }
}

export const notifications = new NotificationStore();
