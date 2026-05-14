declare global {
  namespace App {
    // interface Error {}
    // interface Locals {}
    // interface PageData {}
    // interface Platform {}
  }

  // Injected by Vite `define` from the repo-root VERSION file. Single source
  // of truth for the app version — keep `scripts/release.sh` in sync with
  // package.json so the UI, git tag, and manifest never drift.
  const __APP_VERSION__: string;
}

export {};
