// Cross-context unique identifier.
//
// crypto.randomUUID is only exposed on **secure contexts** (https / localhost).
// LAN access over plain http (e.g. http://192.168.x.x) leaves
// `crypto.randomUUID` as undefined even though `crypto` itself exists, which
// breaks notifications/tabs/uploads ID generation. These IDs are UI keys, not
// security tokens, so a non-cryptographic Math.random fallback is fine.
export function randomId(): string {
  if (typeof crypto !== "undefined" && typeof crypto.randomUUID === "function") {
    return crypto.randomUUID();
  }
  return fallbackUUIDv4();
}

function fallbackUUIDv4(): string {
  // RFC 4122 v4 shape: version nibble fixed to 4, variant nibble (y) forced
  // to the 10xx range so it renders as 8/9/a/b in hex.
  return "xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx".replace(/[xy]/g, (char) => {
    const nibble = Math.floor(Math.random() * 16);
    const value = char === "x" ? nibble : (nibble & 0x3) | 0x8;
    return value.toString(16);
  });
}
