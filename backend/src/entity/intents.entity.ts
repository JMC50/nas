import { EntitySchema } from "../db/metadata";

export const UserIntentEntity: EntitySchema = {
  tableName: "user_intents",
  columns: {
    id: { type: "INTEGER", primary: true, autoincrement: true },
    user_id: { type: "INTEGER", notNull: true },
    intent: { type: "TEXT", notNull: true },
  },
  foreignKeys: [
    {
      column: "user_id",
      references: "users(id)",
      onDelete: "CASCADE",
    },
  ],
};
