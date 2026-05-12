export interface ColumnOptions {
    type: string;
    primary?: boolean;
    notNull?: boolean;
    unique?: boolean;
    autoincrement?: boolean;
}

export interface ForeignKey {
    column: string;
    references: string;
    onDelete?: "CASCADE" | "SET NULL";
}

export interface EntitySchema {
    tableName: string;
    columns: Record<string, ColumnOptions>;
    foreignKeys?: ForeignKey[];
}
