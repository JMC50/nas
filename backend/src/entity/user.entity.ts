import { EntitySchema } from "../db/metadata";

export const UserEntity: EntitySchema = {
    tableName: "users",
    columns: {
        id: { type: "INTEGER", primary: true, autoincrement: true },
        userId: { type: "TEXT", unique: true, notNull: true },
        username: { type: "TEXT", notNull: true },
        global_name: { type: "TEXT" },
        krname: { type: "TEXT" }
    }
};
