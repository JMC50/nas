import getDatabase from "../sqlite";

/**
 * Migration to add local authentication support
 * Adds password and auth_type columns to users table
 */
export function migrateLocalAuth() {
    const db = getDatabase();
    
    try {
        // Check if password column exists
        const tableInfo = db.prepare("PRAGMA table_info(users)").all() as any[];
        const hasPasswordColumn = tableInfo.some(col => col.name === 'password');
        const hasAuthTypeColumn = tableInfo.some(col => col.name === 'auth_type');
        
        if (!hasPasswordColumn) {
            console.log("[Migration] Adding password column to users table...");
            db.prepare(`
                ALTER TABLE users 
                ADD COLUMN password TEXT
            `).run();
            console.log("[Migration] Password column added successfully");
        }
        
        if (!hasAuthTypeColumn) {
            console.log("[Migration] Adding auth_type column to users table...");
            db.prepare(`
                ALTER TABLE users 
                ADD COLUMN auth_type TEXT DEFAULT 'oauth'
            `).run();
            
            // Update existing users to have oauth auth_type
            db.prepare(`
                UPDATE users 
                SET auth_type = 'oauth' 
                WHERE auth_type IS NULL
            `).run();
            
            console.log("[Migration] Auth_type column added successfully");
        }
        
        console.log("[Migration] Local authentication migration completed");
        return true;
    } catch (error) {
        console.error("[Migration] Failed to migrate local auth:", error);
        return false;
    }
}

// Run migration if this file is executed directly
if (require.main === module) {
    const { initializeDatabase } = require("../sqlite");
    initializeDatabase();
    migrateLocalAuth();
}