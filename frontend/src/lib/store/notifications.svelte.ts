import type { Notification, NotificationKind } from "$lib/types";

const DEFAULT_DURATION_MS = 5000;

function randomId() {
  return crypto.randomUUID();
}

class NotificationStore {
  queue = $state<Notification[]>([]);

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
      setTimeout(() => this.dismiss(item.id), item.durationMs);
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
    this.queue = this.queue.filter((item) => item.id !== id);
  }

  clear() {
    this.queue = [];
  }
}

export const notifications = new NotificationStore();
