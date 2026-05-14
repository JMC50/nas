export interface ActivityEntry {
  id?: number;
  activity: string;
  description: string;
  userId?: string;
  krname?: string;
  username?: string;
  time: number; // ms epoch
  loc?: string;
}

export interface DayBucket {
  day: string; // local YYYY-MM-DD
  count: number;
}

export interface TypeShare {
  type: string;
  count: number;
  ratio: number;
}

function localDayKey(time: number): string {
  const d = new Date(time);
  const y = d.getFullYear();
  const m = String(d.getMonth() + 1).padStart(2, "0");
  const day = String(d.getDate()).padStart(2, "0");
  return `${y}-${m}-${day}`;
}

export function dailyCounts(entries: ActivityEntry[], days = 30): DayBucket[] {
  const buckets: DayBucket[] = [];
  const index = new Map<string, number>();
  const now = new Date();
  now.setHours(0, 0, 0, 0);
  for (let i = days - 1; i >= 0; i--) {
    const d = new Date(now);
    d.setDate(now.getDate() - i);
    const key = localDayKey(d.getTime());
    index.set(key, buckets.length);
    buckets.push({ day: key, count: 0 });
  }
  for (const entry of entries) {
    const key = localDayKey(entry.time);
    const slot = index.get(key);
    if (slot !== undefined) buckets[slot].count++;
  }
  return buckets;
}

export function typeDistribution(entries: ActivityEntry[]): TypeShare[] {
  // OPEN → VIEW fold: reviewer-approved consolidation. The list-row icon map
  // (ActivityLog.svelte ACTIVITY_DOT) gives OPEN and VIEW the same color
  // (bg-fg-muted) since they are semantically the same "user read a file"
  // action. Folding them in the distribution avoids two indistinguishable
  // adjacent segments in the stacked bar / legend.
  const counts = new Map<string, number>();
  let total = 0;
  for (const entry of entries) {
    let type = entry.activity.toUpperCase();
    if (type === "OPEN") type = "VIEW";
    counts.set(type, (counts.get(type) ?? 0) + 1);
    total++;
  }
  const shares: TypeShare[] = [];
  for (const [type, count] of counts) {
    shares.push({ type, count, ratio: total === 0 ? 0 : count / total });
  }
  shares.sort((a, b) => b.count - a.count);
  return shares;
}
