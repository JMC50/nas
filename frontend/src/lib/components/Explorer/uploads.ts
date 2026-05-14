import { uploads } from "$lib/store/uploads.svelte";
import { notifications } from "$lib/store/notifications.svelte";
import { pumpQueue } from "$lib/upload-worker";
import { locPath } from "./actions";

export function enqueueFiles(loc: string[], input: HTMLInputElement) {
  if (!input.files || input.files.length === 0) return;
  const here = locPath(loc);
  for (const file of input.files) {
    uploads.enqueue({ file, loc: here, filename: file.name });
  }
  pumpQueue();
  input.value = "";
}

export function enqueueFolder(loc: string[], input: HTMLInputElement) {
  if (!input.files || input.files.length === 0) return;
  const inputCount = input.files.length;
  const hereStr = locPath(loc);
  let enqueuedCount = 0;
  for (const file of input.files) {
    const rel = (file as File & { webkitRelativePath?: string }).webkitRelativePath ?? file.name;
    // Skip files whose relative path contains a dot-prefixed segment
    // (e.g. .git/, .DS_Store, __MACOSX is non-dot and untouched).
    if (rel.split("/").some((seg) => seg.startsWith("."))) continue;
    const parts = rel.split("/").slice(0, -1);
    const fileLoc =
      parts.length > 0
        ? hereStr === "/"
          ? "/" + parts.join("/")
          : hereStr + "/" + parts.join("/")
        : hereStr;
    uploads.enqueue({ file, loc: fileLoc, filename: file.name });
    enqueuedCount++;
  }
  pumpQueue();
  input.value = "";

  if (enqueuedCount === 0 && inputCount > 0) {
    notifications.warning("No uploadable files in selected folder");
  } else if (enqueuedCount > 0) {
    notifications.info(`Queued ${enqueuedCount} file(s) for upload`);
  }
}
