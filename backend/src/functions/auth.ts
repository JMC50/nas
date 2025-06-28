import jwt from "jsonwebtoken";
import { private_key } from "../config/config";
import getDatabase from "../sqlite";

export type intentList = "ADMIN"|"VIEW"|"OPEN"|"DOWNLOAD"|"UPLOAD"|"COPY"|"DELETE"|"RENAME";

interface UserRow {
    id: number;
    userId: string;
    username: string;
    krname: string;
    global_name: string;
}

interface IntentsRow {
    id: number;
    user_id: number,
    intent: string
}

export interface userdata {
    userId: string;
    username: string;
    krname: string;
    global_name: string;
}

export interface user {
    userId: string;
    username: string;
    krname: string;
    global_name: string;
    intents: intentList[];
}

export interface DiscordUserDetail {
    id: string;
    username: string;
    global_name: string;
}

export interface logdata {
    activity: string;
    description: string;
    token: string;
    time: number;
    loc: string;
}

export function jwtVerify(token: string) {
    const check = jwt.verify(token, private_key);
    console.log(check, "ASdasd");

    return;
}

export async function getAllUsers() {
    const db = getDatabase();
    const userRows = db.prepare('SELECT * FROM users').all() as UserRow[];

    const getIntents = db.prepare(`
        SELECT user_id, intent FROM user_intents
    `).all() as IntentsRow[];

    const intentMap = new Map<number, string[]>();
    for (const row of getIntents) {
        if (!intentMap.has(row.user_id)) intentMap.set(row.user_id, []);
        intentMap.get(row.user_id)!.push(row.intent);
    }

    return userRows.map(user => ({
        ...user,
        intents: intentMap.get(user.id) || []
    }));
}

export async function getUser(discordUserId: string) {
    const db = getDatabase();
    const userRow = db
        .prepare('SELECT * FROM users WHERE userId = ?')
        .get(discordUserId) as UserRow;

    if (!userRow) return null;

    const intentsRow = db
        .prepare('SELECT intent FROM user_intents WHERE user_id = ?')
        .all(userRow.id) as IntentsRow[];
    
    const intents = intentsRow.map(row => row.intent);

    return {
        ...userRow,
        intents
    };
}

export async function saveUser(user: DiscordUserDetail, krname: string) {
    const db = getDatabase();
    const insertUser = db.prepare(`
        INSERT INTO users (userId, username, global_name, krname)
        VALUES (?, ?, ?, ?)
    `);

    const getUserRow = db
        .prepare('SELECT id FROM users WHERE userId = ?')
        .get(user.id) as UserRow;

    let userId: number;

    if (!getUserRow) {
        const result = insertUser.run(
            user.id,
            user.username,
            user.global_name,
            krname
        );
        userId = result.lastInsertRowid as number;

        // 기본 intent 부여
        // const insertIntent = db.prepare(`
        //     INSERT INTO user_intents (user_id, intent) VALUES (?, ?)
        // `);
        // insertIntent.run(userId, 'VIEW');
    } else {
        // 이미 존재함, 업데이트만
        db.prepare(`
            UPDATE users SET username = ?, global_name = ?, krname = ?
            WHERE userId = ?
        `).run(user.username, user.global_name, krname, user.id);

        userId = getUserRow.id;
    }

    return userId;
}

export function createToken(userId: string) {
    const expires_in = Date.now() + 7 * 24 * 60 * 60 * 1000
    const token = jwt.sign({
        userId,
        expires_in
    }, private_key);

    return token;
}

export async function getActivityLogs() {
    const db = getDatabase();
    const rows = db.prepare(`
        SELECT log.activity, log.description, log.time, log.loc, users.userId, users.username, users.krname
        FROM log
        LEFT JOIN users ON log.user_id = users.id
        ORDER BY log.time DESC
    `).all();

    return rows;
}

export function convertLoc(loc:string) {
    const splited = loc.split("/");
    while(splited.length > 1 && splited[0] == ""){
        splited.shift();
    }

    return splited.join("/");
}

export async function insertLog(data: logdata) {
    const db = getDatabase();
    const decoded = <jwt.JwtPayload>jwt.decode(data.token);
    const userId = decoded.userId;

    // 유저 id 가져오기
    const userRow = db.prepare('SELECT id FROM users WHERE userId = ?').get(userId) as UserRow;

    if (!userRow) {
        throw new Error("User not found for logging");
    }

    const insert = db.prepare(`
        INSERT INTO log (activity, description, user_id, time, loc)
        VALUES (?, ?, ?, ?, ?)
    `);

    insert.run(
        data.activity,
        data.description,
        userRow.id,
        data.time,
        convertLoc(data.loc)
    );
}

export async function editIntent(discordUserId: string, intent: string) {
    const db = getDatabase();
    const userRow = db
        .prepare('SELECT id FROM users WHERE userId = ?')
        .get(discordUserId) as UserRow;

    if (!userRow) return;

    const hasIntent = db
        .prepare('SELECT * FROM user_intents WHERE user_id = ? AND intent = ?')
        .get(userRow.id, intent);

    if (hasIntent) {
        db.prepare(
            'DELETE FROM user_intents WHERE user_id = ? AND intent = ?'
        ).run(userRow.id, intent);
    } else {
        db.prepare(
            'INSERT INTO user_intents (user_id, intent) VALUES (?, ?)'
        ).run(userRow.id, intent);
    }
}

export async function checkIntent(userId: string, intent: intentList) {
    const db = getDatabase();
    const userRow = db.prepare('SELECT id FROM users WHERE userId = ?').get(userId) as UserRow;
    if (!userRow) return false;

    // 유저가 ADMIN 권한인지 먼저 체크
    const adminIntent = db.prepare(`
        SELECT 1 FROM user_intents WHERE user_id = ? AND intent = 'ADMIN'
    `).get(userRow.id);

    if (adminIntent) return true;

    // 원하는 intent 체크
    const hasIntent = db.prepare(`
        SELECT 1 FROM user_intents WHERE user_id = ? AND intent = ?
    `).get(userRow.id, intent);

    return !!hasIntent;
}