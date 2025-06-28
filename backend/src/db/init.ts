import fs from 'fs';
import path from 'path';
import getDatabase from '../sqlite';
import { EntitySchema } from './metadata';

export function initializeEntities() {
    const db = getDatabase();
    const entityDir = path.join(__dirname, '..', 'entity');
    const files = fs.readdirSync(entityDir).filter(file => file.endsWith('.ts') || file.endsWith('.js'));

    for (const file of files) {
        const entityPath = path.join(entityDir, file);
        const entityModule = require(entityPath); // ts-node 환경일 경우 OK
        const entity: EntitySchema = Object.values(entityModule)[0] as EntitySchema;
        if (!entity?.tableName || !entity?.columns) continue;

        const columnDefs = Object.entries(entity.columns).map(([name, opts]) => {
            let def = `${name} ${opts.type}`;
            if (opts.primary) def += " PRIMARY KEY";
            if (opts.autoincrement) def += " AUTOINCREMENT";
            if (opts.notNull) def += " NOT NULL";
            if (opts.unique) def += " UNIQUE";
            return def;
        });

        const foreignDefs = (entity.foreignKeys || []).map(fk =>
            `FOREIGN KEY (${fk.column}) REFERENCES ${fk.references}` +
            (fk.onDelete ? ` ON DELETE ${fk.onDelete}` : "")
        );

        const sql = `
            CREATE TABLE IF NOT EXISTS ${entity.tableName} (
                ${[...columnDefs, ...foreignDefs].join(',\n')}
            );
        `;

        db.prepare(sql).run();
        console.log(`[DB] Created table: ${entity.tableName}`);
    }
}
