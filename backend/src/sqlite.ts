import path from 'path';
import Database, { Database as DatabaseType } from 'better-sqlite3';
import fs from 'fs';

const dbPath = path.join(__dirname, '..', 'db', 'nas.sqlite');
let db: DatabaseType | null = null;

export function initializeDatabase() {
    if(!db){
        const dbDir = path.dirname(dbPath);
        if(!fs.existsSync(dbDir)){
            fs.mkdirSync(dbDir, { recursive: true });
        }
        db = new Database(dbPath);
        console.log("created db folder");
    }
    return db;
}

export function getDatabase() {
    if (!db) {
        throw new Error('Database not initialized. Call initializeDatabase() first.');
    }
    return db;
}

export default getDatabase;
