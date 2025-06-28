import path from 'path';
import Database from 'better-sqlite3';

const dbPath = path.join(__dirname, '..', 'db', 'nas.sqlite');
const db = new Database(dbPath);

export default db;
