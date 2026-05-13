// One-off: grant ADMIN intent to the user with userId "admin".
// Run with: node scripts/grant-admin.mjs
// Backend must be stopped while this runs (SQLite write lock).

import Database from "better-sqlite3";
import path from "node:path";
import { fileURLToPath } from "node:url";

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);
const DB_PATH = path.resolve(__dirname, "..", "data", "db", "nas.sqlite");

const targetUserId = process.argv[2] ?? "admin";

const db = new Database(DB_PATH);
const user = db.prepare("SELECT id, userId, username FROM users WHERE userId = ?").get(targetUserId);
if (!user) {
  console.error(`User not found: userId=${targetUserId}`);
  process.exit(1);
}
console.log(`Found user: id=${user.id}, userId=${user.userId}, username=${user.username}`);

const existing = db
  .prepare("SELECT 1 FROM user_intents WHERE user_id = ? AND intent = 'ADMIN'")
  .get(user.id);
if (existing) {
  console.log("Already has ADMIN intent. Nothing to do.");
} else {
  db.prepare("INSERT INTO user_intents (user_id, intent) VALUES (?, 'ADMIN')").run(user.id);
  console.log("Granted ADMIN intent.");
}

const intents = db
  .prepare("SELECT intent FROM user_intents WHERE user_id = ?")
  .all(user.id)
  .map((row) => row.intent);
console.log(`Current intents for ${user.userId}: [${intents.join(", ")}]`);

db.close();
