import { EntitySchema } from "../db/metadata";

export const LogEntity: EntitySchema = {
  tableName: "log",
  columns: {
    id: { type: "INTEGER", primary: true, autoincrement: true },
    activity: { type: "TEXT", notNull: true },
    description: { type: "TEXT" },
    user_id: { type: "INTEGER", notNull: true },
    time: { type: "INTEGER", notNull: true },
    loc: { type: "TEXT" },
  },
  foreignKeys: [
    {
      column: "user_id",
      references: "users(id)",
      onDelete: "SET NULL",
    },
  ],
};
