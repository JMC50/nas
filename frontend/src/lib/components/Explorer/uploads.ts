import { uploads } from "$lib/store/uploads.svelte";
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
  const hereStr = locPath(loc);
  for (const file of input.files) {
    const rel = (file as File & { webkitRelativePath?: string }).webkitRelativePath ?? file.name;
    const parts = rel.split("/").slice(0, -1);
    const fileLoc =
      parts.length > 0
        ? hereStr === "/"
          ? "/" + parts.join("/")
          : hereStr + "/" + parts.join("/")
        : hereStr;
    uploads.enqueue({ file, loc: fileLoc, filename: file.name });
  }
  pumpQueue();
  input.value = "";
}
