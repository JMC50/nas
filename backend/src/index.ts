import express from "express";
import path from "path";
import fs from 'fs/promises';
import fsNP from "fs";
import yauzl from "yauzl";
import jwt from "jsonwebtoken"
import { formatDate, formatSize, getCpuUsage, getDiskUsage, getMemoryUsage, getUptime, searchFilesInDir } from "./functions/general";
import { checkIntent, convertLoc, createToken, DiscordUserDetail, editIntent, getActivityLogs, getAllUsers, getUser, insertLog, intentList, jwtVerify, logdata, saveUser, userdata } from "./functions/auth";
import { admin_passowrd, KAKAO_CLIENT_SECRET, KAKAO_REDIRECT_URL, KAKAO_REST_API_KEY, PORT, private_key } from "./config/config";
import { initializeEntities } from "./db/init";
import archiver from "archiver";
import * as mkdirp from "mkdirp";
import { pipeline } from "stream";
import { promisify } from "util";
import { v4 as uuidv4 } from "uuid";

const app = express();

app.use(express.json({ limit: "50gb" }));
app.use(express.raw({ type: "application/octet-stream", limit: "50gb" }));
app.use('/monaco-editor', express.static(path.join(__dirname, "..", "..", 'node_modules', 'monaco-editor', 'min', 'vs')));
app.use((req, res, next) => {
    res.setHeader("Connection", "keep-alive");
    next();
});

process.on("SIGINT", async () => {
    process.exit(0);
});

const check3 = fsNP.existsSync(`./db`);
if(!check3){
    fsNP.mkdirSync(`./db`);
    console.log("created db folder");
}

(async() => {
    const check = fsNP.existsSync(`../../nas-data`);
    if(!check){
        await fs.mkdir(`../../nas-data`);
        console.log("created data folder");
    }

    const check2 = fsNP.existsSync(`../../nas-data-admin`);
    if(!check2){
        await fs.mkdir(`../../nas-data-admin`);
        console.log("created admin data folder");
    }


    initializeEntities();
})()

app.get('/', async (req, res) => {
    res.end("server is running :D");
});

process.on('uncaughtException', (err) => {
    console.error('Uncaught Exception:', err);
});

process.on('unhandledRejection', (reason, promise) => {
    console.error('Unhandled Rejection at:', promise, 'reason:', reason);
});

app.get("/getSystemInfo", async (req, res) => {
    const disk = await getDiskUsage();
    res.json({
        cpu: getCpuUsage(),
        memory: getMemoryUsage(),
        uptime: getUptime(),
        disk
    })
})

app.get("/stat", async (req, res) => {
    const loc = decodeURIComponent(req.query.loc as string);
    const name = decodeURIComponent(req.query.name as string);
    const targetPath =  path.resolve(`../../nas-data/${loc}/${name}`);
    const stat = await fs.stat(targetPath);

    const fileInfo = {
        name,
        size: stat.isFile() ? formatSize(stat.size) : undefined,
        type: stat.isDirectory() ? "folder" : "file",
        createdAt: formatDate(stat.birthtime),
        modifiedAt: formatDate(stat.mtime)
    };

    res.json(fileInfo);
});

app.get("/download", async (req, res) => {
    const loc = decodeURIComponent(req.query.loc as string);
    const name = decodeURIComponent(req.query.name as string);

    const token = req.query.token as string;
    if(!token){
        res.status(500).end("token required");
        return;
    }

    let decoded: jwt.JwtPayload;
    try{
        decoded = jwt.verify(token, private_key) as jwt.JwtPayload;
    }catch(err){
        res.status(500).end("wtf is this token");
        return;
    }

    const checkI = await checkIntent(decoded.userId, "DOWNLOAD");
    if(!checkI){
        res.status(500).end("no intents");
        return;
    }

    const filepath = path.resolve(`../../nas-data/${loc}/${name}`);
    res.download(filepath, `${name}`);
})

app.get("/getTextFile", async (req, res) => {
    const loc = decodeURIComponent(req.query.loc as string);
    const name = decodeURIComponent(req.query.name as string);

    const token = req.query.token as string;
    if(!token){
        res.status(500).end("token required");
        return;
    }

    let decoded: jwt.JwtPayload;
    try{
        decoded = jwt.verify(token, private_key) as jwt.JwtPayload;
    }catch(err){
        res.status(500).end("wtf is this token");
        return;
    }

    const checkI = await checkIntent(decoded.userId, "OPEN");
    if(!checkI){
        res.status(500).end("no intents");
        return;
    }

    const targetPath = path.resolve(`../../nas-data/${loc}/${name}`);
    const content = await fs.readFile(targetPath, "utf-8");

    res.json({ name, content });
});

app.post("/saveTextFile", async (req, res) => {
    const loc = decodeURIComponent(req.query.loc as string);
    const name = decodeURIComponent(req.query.name as string);
    const data = req.body;
    const saveText = <string>data.text;

    const token = req.query.token as string;
    if(!token){
        res.status(500).end("token required");
        return;
    }

    let decoded: jwt.JwtPayload;
    try{
        decoded = jwt.verify(token, private_key) as jwt.JwtPayload;
    }catch(err){
        res.status(500).end("wtf is this token");
        return;
    }

    const checkI = await checkIntent(decoded.userId, "UPLOAD");
    if(!checkI){
        res.status(500).end("no intents");
        return;
    }
    
    const targetPath = path.resolve(`../../nas-data/${loc}/${name}`);
    await fs.writeFile(targetPath, saveText)

    res.end("complete");
});

app.get("/getVideoData", async (req, res) => {
    const loc = decodeURIComponent(req.query.loc as string);
    const name = decodeURIComponent(req.query.name as string);

    const token = req.query.token as string;
    if(!token){
        res.status(500).end("token required");
        return;
    }

    let decoded: jwt.JwtPayload;
    try{
        decoded = jwt.verify(token, private_key) as jwt.JwtPayload;
    }catch(err){
        res.status(500).end("wtf is this token");
        return;
    }

    const checkI = await checkIntent(decoded.userId, "OPEN");
    if(!checkI){
        res.status(500).end("no intents");
        return;
    }

    const videoPath = path.resolve(`../../nas-data/${loc}/${name}`);
    const stat = fsNP.statSync(videoPath);
    const fileSize = stat.size;
    const range = req.headers.range;

    if(!range){
        res.writeHead(200, {
            "Content-Type": "video/mp4",
            "Content-Length": fileSize,
        });
        fsNP.createReadStream(videoPath).pipe(res);
    }else{
        const CHUNK_SIZE = 10 ** 6;
        const start = Number(range.replace(/\D/g, ""));
        const end = Math.min(start + CHUNK_SIZE, fileSize - 1);

        const headers = {
            "Content-Range": `bytes ${start}-${end}/${fileSize}`,
            "Accept-Ranges": "bytes",
            "Content-Length": end - start + 1,
            "Content-Type": "video/mp4",
        };

        res.writeHead(206, headers);
        const videoStream = fsNP.createReadStream(videoPath, { start, end });
        videoStream.pipe(res);
    }
});

app.get("/getAudioData", async (req, res) => {
    const loc = decodeURIComponent(req.query.loc as string);
    const name = decodeURIComponent(req.query.name as string);

    const token = req.query.token as string;
    if(!token){
        res.status(500).end("token required");
        return;
    }

    let decoded: jwt.JwtPayload;
    try{
        decoded = jwt.verify(token, private_key) as jwt.JwtPayload;
    }catch(err){
        res.status(500).end("wtf is this token");
        return;
    }

    const checkI = await checkIntent(decoded.userId, "OPEN");
    if(!checkI){
        res.status(500).end("no intents");
        return;
    }

    const audioPath = path.resolve(`../../nas-data/${loc}/${name}`);
    const stat = fsNP.statSync(audioPath);
    const fileSize = stat.size;
    const range = req.headers.range;

    if(!range){
        res.writeHead(200, {
            "Content-Type": "audio/mpeg",
            "Content-Length": fileSize,
        });
        fsNP.createReadStream(audioPath).pipe(res);
    }else{
        const CHUNK_SIZE = 1024 * 1024;
        const start = Number(range.replace(/\D/g, ""));
        const end = Math.min(start + CHUNK_SIZE, fileSize - 1);

        const headers = {
            "Content-Range": `bytes ${start}-${end}/${fileSize}`,
            "Accept-Ranges": "bytes",
            "Content-Length": end - start + 1,
            "Content-Type": "audio/mpeg",
        };

        res.writeHead(206, headers);
        const audioStream = fsNP.createReadStream(audioPath, { start, end });
        audioStream.pipe(res);
    }
});

app.get("/getImageData", async (req, res) => {
    const loc = decodeURIComponent(<string>req.query.loc);
    const name = decodeURIComponent(<string>req.query.name);

    const token = req.query.token as string;
    if(!token){
        res.status(500).end("token required");
        return;
    }

    let decoded: jwt.JwtPayload;
    try{
        decoded = jwt.verify(token, private_key) as jwt.JwtPayload;
    }catch(err){
        res.status(500).end("wtf is this token");
        return;
    }

    const checkI = await checkIntent(decoded.userId, "OPEN");
    if(!checkI){
        res.status(500).end("no intents");
        return;
    }

    let contentType = "application/octet-stream";
    const ext = path.extname(name).toLowerCase();
    if (ext === ".jpg" || ext === ".jpeg") {
        contentType = "image/jpeg";
    } else if (ext === ".png") {
        contentType = "image/png";
    } else if (ext === ".svg") {
        contentType = "image/svg+xml";
    }

    res.setHeader("Content-Type", contentType);
    const file = await fs.readFile(path.resolve(`../../nas-data/${loc}/${name}`));
    res.end(file);
})

app.get("/forceDelete", async (req, res) => {
    const loc = decodeURIComponent(req.query.loc as string);
    const name = decodeURIComponent(req.query.name as string);

    const token = req.query.token as string;
    if(!token){
        res.status(500).end("token required");
        return;
    }

    let decoded: jwt.JwtPayload;
    try{
        decoded = jwt.verify(token, private_key) as jwt.JwtPayload;
    }catch(err){
        res.status(500).end("wtf is this token");
        return;
    }

    const checkI = await checkIntent(decoded.userId, "DELETE");
    if(!checkI){
        res.status(500).end("no intents");
        return;
    }
    await insertLog({
        token,
        activity: "DELETE",
        description: `DELETE [FILE] AT /${convertLoc(`${loc}/${name}`)}`,
        loc: `/${loc}`,
        time: Date.now()
    })

    const targetPath = path.resolve(`../../nas-data/${loc}/${name}`);

    await fs.rm(targetPath, { recursive: true, force: true });
    res.end("complete");
});

app.get("/copy", async (req, res) => {
    const originLoc = decodeURIComponent(req.query.originLoc as string);
    const fileName = decodeURIComponent(req.query.fileName as string);
    const targetLoc = decodeURIComponent(req.query.targetLoc as string);

    const token = req.query.token as string;
    if(!token){
        res.status(500).end("token required");
        return;
    }

    let decoded: jwt.JwtPayload;
    try{
        decoded = jwt.verify(token, private_key) as jwt.JwtPayload;
    }catch(err){
        res.status(500).end("wtf is this token");
        return;
    }

    const checkI = await checkIntent(decoded.userId, "COPY");
    if(!checkI){
        res.status(500).end("no intents");
        return;
    }
    await insertLog({
        token,
        activity: "COPY",
        description: `COPY [FILE] FROM /${convertLoc(`${originLoc}/${fileName}`)} TO /${convertLoc(`${targetLoc}/${fileName}`)}`,
        loc: `/${targetLoc}`,
        time: Date.now()
    })

    const sourcePath = path.resolve(`../../nas-data/${originLoc}/${fileName}`);
    const targetPath = path.resolve(`../../nas-data/${targetLoc}/${fileName}`);

    const stat = await fs.stat(sourcePath);

    if(stat.isDirectory()){
        await fs.cp(sourcePath, targetPath, { recursive: true });
    }else{
        await fs.copyFile(sourcePath, targetPath);
    }
    res.end("complete");
});

app.get("/move", async (req, res) => {
    const originLoc = decodeURIComponent(req.query.originLoc as string);
    const fileName = decodeURIComponent(req.query.fileName as string);
    const targetLoc = decodeURIComponent(req.query.targetLoc as string);

    const token = req.query.token as string;
    if (!token) {
        res.status(500).end("token required");
        return;
    }
    
    let decoded: jwt.JwtPayload;
    try {
        decoded = jwt.verify(token, private_key) as jwt.JwtPayload;
    } catch (err) {
        res.status(500).end("wtf is this token");
        return;
    }

    const checkI = await checkIntent(decoded.userId, "COPY");
    if (!checkI) {
        res.status(500).end("no intents");
        return;
    }

    await insertLog({
        token,
        activity: "MOVE",
        description: `MOVE [FILE] FROM /${convertLoc(`${originLoc}/${fileName}`)} TO /${convertLoc(`${targetLoc}/${fileName}`)}`,
        loc: `/${targetLoc}`,
        time: Date.now()
    });
    
    const sourcePath = path.resolve(`../../nas-data/${originLoc}/${fileName}`);
    const targetPath = path.resolve(`../../nas-data/${targetLoc}/${fileName}`);

    try {
        await fs.rename(sourcePath, targetPath);
        res.end("complete");
    } catch (err) {
        res.status(500).end("move failed");
    }
});

app.get("/rename", async (req, res) => {
    const loc = decodeURIComponent(req.query.loc as string);
    const name = decodeURIComponent(req.query.name as string);
    const change = decodeURIComponent(req.query.change as string);

    const token = req.query.token as string;
    if(!token){
        res.status(500).end("token required");
        return;
    }

    let decoded: jwt.JwtPayload;
    try{
        decoded = jwt.verify(token, private_key) as jwt.JwtPayload;
    }catch(err){
        res.status(500).end("wtf is this token");
        return;
    }

    const checkI = await checkIntent(decoded.userId, "RENAME");
    if(!checkI){
        res.status(500).end("no intents");
        return;
    }
    await insertLog({
        token,
        activity: "RENAME",
        description: `RENAME [FILE] AT /${convertLoc(`${loc}/${name}`)} TO /${convertLoc(`${loc}/${change}`)}`,
        loc: `/${loc}`,
        time: Date.now()
    })

    await fs.rename(path.resolve(`../../nas-data/${loc}/${name}`), path.resolve(`../../nas-data/${loc}/${change}`));
    res.end("complete");
});

app.post("/zipFiles", async (req, res) => {
    const token = req.query.token as string;
    if(!token){
        res.status(500).end("token required");
        return;
    }

    let decoded: jwt.JwtPayload;
    try{
        decoded = jwt.verify(token, private_key) as jwt.JwtPayload;
    }catch(err){
        res.status(500).end("wtf is this token");
        return;
    }

    const checkI = await checkIntent(decoded.userId, "UPLOAD");
    if(!checkI){
        res.status(500).end("no intents");
        return;
    }

    const files = req.body;
    if (!Array.isArray(files) || files.length === 0) {
        res.status(400).send("no files to zip");
        return;
    }

    const progressId = uuidv4();
    const progressPath = `/tmp/nas-progress-${progressId}.json`;
    fsNP.writeFileSync(progressPath, JSON.stringify({ percent: 0, status: "zipping" }));

    const basePath = "../../nas-data";
    const zipLoc = path.join(basePath, files[0].loc, `archive_${Date.now()}.zip`);
    const output = fsNP.createWriteStream(zipLoc);
    const archive = archiver('zip', { zlib: { level: 9 } });

    let totalFiles = files.length;
    let processedFiles = 0;

    output.on("close", () => {
        fsNP.writeFileSync(progressPath, JSON.stringify({ percent: 100, status: "done" }));
        res.status(200).json({ zipPath: zipLoc, progressId });
    });

    archive.on("error", (err) => {
        fsNP.writeFileSync(progressPath, JSON.stringify({ percent: 100, status: "error" }));
        res.status(500).send("failed to zip files");
    });

    archive.on("entry", () => {
        processedFiles++;
        const percent = Math.floor((processedFiles / totalFiles) * 100);
        fsNP.writeFileSync(progressPath, JSON.stringify({ percent, status: "zipping" }));
    });

    archive.pipe(output);

    for (const file of files) {
        const absPath = path.join(basePath, file.loc, file.name);
        if (file.isFolder) {
            archive.directory(absPath, file.name);
        } else {
            archive.file(absPath, { name: file.name });
        }
    }

    archive.finalize();
});

const pump = promisify(pipeline);

app.post("/unzipFile", async (req, res) => {
    const token = req.query.token as string;
    if (!token) {
        res.status(500).end("token required");
        return;
    }

    let decoded: jwt.JwtPayload;
    try {
        decoded = jwt.verify(token, private_key) as jwt.JwtPayload;
    } catch (err) {
        res.status(500).end("wtf is this token");
        return;
    }

    const checkI = await checkIntent(decoded.userId, "UPLOAD");
    if (!checkI) {
        res.status(500).end("no intents");
        return;
    }

    const file = req.body;
    if (file.extensions !== "zip") {
        res.status(400).send("invalid file");
        return;
    }

    const progressId = uuidv4();
    const progressPath = `/tmp/nas-progress-${progressId}.json`;
    fsNP.writeFileSync(progressPath, JSON.stringify({ percent: 0, status: "unzipping" }));

    const basePath = "../../nas-data";
    const zipFilePath = path.join(basePath, file.loc, file.name);
    const extractTo = path.join(basePath, file.loc, path.parse(file.name).name + "_unzipped");

    try {
        yauzl.open(zipFilePath, { lazyEntries: true }, (err, zipfile) => {
            if (err || !zipfile) {
                fsNP.writeFileSync(progressPath, JSON.stringify({ percent: 100, status: "error" }));
                res.status(500).send("failed to open zip");
                return;
            }

            let totalEntries = 0;
            let processedEntries = 0;

            zipfile.on("entry", () => { totalEntries++; });
            zipfile.readEntry();

            zipfile.on("entry", async (entry) => {
                const entryPath = path.join(extractTo, entry.fileName);
                
                if (/\/$/.test(entry.fileName)) {
                    mkdirp.sync(entryPath);
                    processedEntries++;
                    fsNP.writeFileSync(progressPath, JSON.stringify({ percent: Math.floor((processedEntries / totalEntries) * 100), status: "unzipping" }));
                    zipfile.readEntry();
                } else {
                    mkdirp.sync(path.dirname(entryPath));
                    zipfile.openReadStream(entry, async (err, readStream) => {
                        if (err || !readStream) {
                            fsNP.writeFileSync(progressPath, JSON.stringify({ percent: 100, status: "error" }));
                            zipfile.close();
                            res.status(500).send("failed to extract file");
                            return;
                        }
                        const writeStream = fsNP.createWriteStream(entryPath);
                        writeStream.on("close", () => {
                            processedEntries++;
                            fsNP.writeFileSync(progressPath, JSON.stringify({ percent: Math.floor((processedEntries / totalEntries) * 100), status: "unzipping" }));
                            zipfile.readEntry();
                        });
                        readStream.pipe(writeStream);
                    });
                }
            });

            zipfile.on("end", () => {
                fsNP.writeFileSync(progressPath, JSON.stringify({ percent: 100, status: "done" }));
                res.status(200).json({ extractedPath: extractTo, progressId });
            });

            zipfile.on("error", (err) => {
                fsNP.writeFileSync(progressPath, JSON.stringify({ percent: 100, status: "error" }));
                res.status(500).send("zip processing error");
            });
        });
    } catch (err) {
        fsNP.writeFileSync(progressPath, JSON.stringify({ percent: 100, status: "error" }));
        res.status(500).send("unexpected error");
    }
});

app.get("/progress", (req, res) => {
    const progressId = req.query.progressId as string;
    if (!progressId) {
        res.status(400).send("progressId required");
        return;
    }
    const progressPath = `/tmp/nas-progress-${progressId}.json`;
    if (!fsNP.existsSync(progressPath)) {
        res.status(404).send("not found");
        return;
    }
    const data = fsNP.readFileSync(progressPath, "utf-8");
    res.json(JSON.parse(data));
});

app.get("/makedir", async (req, res) => {
    const loc = decodeURIComponent(req.query.loc as string);
    const name = decodeURIComponent(req.query.name as string);

    const token = req.query.token as string;
    if(!token){
        res.status(500).end("token required");
        return;
    }

    let decoded: jwt.JwtPayload;
    try{
        decoded = jwt.verify(token, private_key) as jwt.JwtPayload;
    }catch(err){
        res.status(500).end("wtf is this token");
        return;
    }

    const checkI = await checkIntent(decoded.userId, "UPLOAD");
    if(!checkI){
        res.status(500).end("no intents");
        return;
    }
    await insertLog({
        token,
        activity: "UPLOAD",
        description: `CREATE [FOLDER] AT /${convertLoc(`${loc}/${name}`)}`,
        loc: `/${loc}`,
        time: Date.now()
    })


    const check = fsNP.existsSync(path.resolve(`../../nas-data/${loc}/${name}`));
    if(check){
        res.end("failed");
    }else{
        await fs.mkdir(path.resolve(`../../nas-data/${loc}/${name}`), { recursive: true });
        res.end("complete");
    }
});

app.get("/readFolder", async (req, res) => {
    const loc = decodeURIComponent(<string>req.query.loc);

    const token = req.query.token as string;
    if(!token){
        res.status(500).end("token required");
        return;
    }

    let decoded: jwt.JwtPayload;
    try{
        decoded = jwt.verify(token, private_key) as jwt.JwtPayload;
    }catch(err){
        res.status(500).end("wtf is this token");
        return;
    }

    const checkI = await checkIntent(decoded.userId, "VIEW");
    if(!checkI){
        res.status(500).end("no intents");
        return;
    }

    const dir = await fs.opendir(path.resolve(`../../nas-data/${loc}`));
    const arr = [];
    for await(let i of dir){
        const match = i.name.match(/\.([^.]+)$/);
        if(match){
            arr.push({
                name: i.name,
                isFolder: i.isDirectory(),
                extensions: match[1]
            });
        }else{
            arr.push({
                name: i.name,
                isFolder: i.isDirectory(),
                extensions: "file"
            });
        }
    }
    res.json(arr);
});

interface FileInfo{
    name: string;
    isFolder: boolean;
    extensions: string;
    loc: string;
}

app.get("/searchInAllFiles", async (req, res) => {
    const query = decodeURIComponent((req.query.query as string)).toLowerCase();

    const token = req.query.token as string;
    if(!token){
        res.status(500).end("token required");
        return;
    }

    let decoded: jwt.JwtPayload;
    try{
        decoded = jwt.verify(token, private_key) as jwt.JwtPayload;
    }catch(err){
        res.status(500).end("wtf is this token");
        return;
    }

    const checkI = await checkIntent(decoded.userId, "VIEW");
    if(!checkI){
        res.status(500).end("no intents");
        return;
    }

    const baseDir = path.resolve("../../nas-data/");
    const result: FileInfo[] = [];

    await searchFilesInDir(baseDir, query, result);
    res.json(result);
});

app.get("/img", async (req, res) => {
    const type = req.query.type;
    const filePath = path.resolve(__dirname, "..", "img", `${type}.png`);
    const check = fsNP.existsSync(filePath);
    if(check){
        const img = await fs.readFile(filePath);
        res.end(img);
    }else{
        const img = await fs.readFile(path.resolve(__dirname, "..", "img", `file.png`));
        res.end(img);
    }
});

app.post("/input", async (req, res) => {
    const name = decodeURIComponent(req.query.name as string);
    const loc = decodeURIComponent(req.query.loc as string);

    const token = req.query.token as string;
    if(!token){
        res.status(500).end("token required");
        return;
    }

    let decoded: jwt.JwtPayload;
    try{
        decoded = jwt.verify(token, private_key) as jwt.JwtPayload;
    }catch(err){
        res.status(500).end("wtf is this token");
        return;
    }

    const checkI = await checkIntent(decoded.userId, "UPLOAD");
    if(!checkI){
        res.status(500).end("no intents");
        return;
    }
    await insertLog({
        token,
        activity: "UPLOAD",
        description: `UPLOAD [FILE] AT /${convertLoc(`${loc}/${name}`)}`,
        loc: `/${loc}`,
        time: Date.now()
    })

    const targetPath = path.resolve(`../../nas-data/${loc}/${name}`);

    try{
        const writeStream = fsNP.createWriteStream(targetPath);
        req.pipe(writeStream);

        writeStream.on("finish", () => {
            res.end("complete");
        });

        writeStream.on("error", (err) => {
            console.error("파일 저장 중 오류 발생:", err);
            res.status(500).end("error");
        });
    }catch(error){
        console.error("오류 발생:", error);
        res.status(500).end("error");
    }
});

app.post("/inputZip", async (req, res) => {
    const name = decodeURIComponent(req.query.name as string);
    const loc = decodeURIComponent(req.query.loc as string);

    const token = req.query.token as string;
    if(!token){
        res.status(500).end("token required");
        return;
    }

    let decoded: jwt.JwtPayload;
    try{
        decoded = jwt.verify(token, private_key) as jwt.JwtPayload;
    }catch(err){
        res.status(500).end("wtf is this token");
        return;
    }

    const checkI = await checkIntent(decoded.userId, "UPLOAD");
    if(!checkI){
        res.status(500).end("no intents");
        return;
    }
    await insertLog({
        token,
        activity: "UPLOAD",
        description: `UPLOAD [FOLDER] AT /${convertLoc(`${loc}/${name}`)}`,
        loc: `/${loc}`,
        time: Date.now()
    })

    const targetZipPath = path.resolve(`../../nas-data/${loc}/${name}`);
    const extractPath = path.resolve(`../../nas-data/${loc}`);
  
    try{
        if(!fsNP.existsSync(extractPath)){
            await fs.mkdir(extractPath, { recursive: true });
        }
  
        const writeStream = fsNP.createWriteStream(targetZipPath);
        req.pipe(writeStream);
    
        writeStream.on("finish", () => {
            yauzl.open(targetZipPath, { lazyEntries: true }, (err, zipfile) => {
                if(err){
                    console.error("압축 해제 중 오류 발생:", err);
                    res.status(500).end("error");
                    return;
                }
                zipfile.readEntry();
        
                zipfile.on("entry", (entry) => {
                    if(/\/$/.test(entry.fileName)){
                        const dirPath = path.join(extractPath, entry.fileName);
                        fsNP.mkdirSync(dirPath, { recursive: true });
                        zipfile.readEntry();
                    }else{
                        zipfile.openReadStream(entry, (err, readStream) => {
                            if(err){
                                console.error("압축 해제 중 오류 발생:", err);
                                res.status(500).end("error");
                                return;
                            }
                            const filePath = path.join(extractPath, entry.fileName);
                            fsNP.mkdirSync(path.dirname(filePath), { recursive: true });
                            const fileWriteStream = fsNP.createWriteStream(filePath);
                            readStream.pipe(fileWriteStream);
                            readStream.on("end", () => {
                                zipfile.readEntry();
                            });
                            readStream.on("error", (err) => {
                                console.error("읽기 스트림 오류:", err);
                                res.status(500).end("error");
                            });
                        });
                    }
                });
        
                zipfile.on("end", () => {
                    fsNP.unlink(targetZipPath, (err) => {
                        if(err) console.error("임시 zip 파일 삭제 실패:", err);
                    });
                    res.end("complete");
                });
        
                zipfile.on("error", (err) => {
                    console.error("압축 해제 중 오류 발생:", err);
                    res.status(500).end("error");
                });
            });
        });
    
        writeStream.on("error", (err) => {
            console.error("파일 저장 중 오류 발생:", err);
            res.status(500).end("error");
        });
    }catch(error){
        console.error("오류 발생:", error);
        res.status(500).end("error");
    }
});

app.get("/login", async (req, res) => {
    const access_token = String(req.query.access_token);

    const getuser = await fetch('https://discord.com/api/users/@me', {
        headers: {
            authorization: `Bearer ${access_token}`,
        },
    })

    const userdata = await getuser.json() as DiscordUserDetail;

    const check = await getUser(userdata.id);
    if(check){
        const token = createToken(check.userId);
        res.json({ ...check, token });
    }else{
        res.json({
            status: "new",
            userId: userdata.id,
            username: userdata.username,
            global_name: userdata.global_name
        })
    }
})

app.get("/kakaoLogin", async (req, res) => {
    const code = decodeURIComponent(String(req.query.code));

    const tokenRes = await fetch('https://kauth.kakao.com/oauth/token', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/x-www-form-urlencoded',
        },
        body: new URLSearchParams({
            grant_type: 'authorization_code',
            client_id: KAKAO_REST_API_KEY,
            redirect_uri: KAKAO_REDIRECT_URL,
            client_secret: KAKAO_CLIENT_SECRET,
            code,
        })
    })

    const tokenData = await tokenRes.json();

    const userRes = await fetch('https://kapi.kakao.com/v2/user/me', {
        headers: {
            Authorization: `Bearer ${tokenData.access_token}`,
            'Content-Type': 'application/x-www-form-urlencoded;charset=utf-8',
        },
    });

    const userData = await userRes.json();

    const check = await getUser(userData.id);
    if(check){
        const token = createToken(check.userId);
        res.json({
            userId: userData.id,
            nickname: userData.properties.nickname,
            token
        })
    }else{
        res.json({
            status: "new",
            userId: userData.id,
            nickname: userData.properties.nickname
        })
    }
})

app.post("/register", async (req, res) => {
    const data = req.body;
    const access_token = data.access_token;
    const krname = data.krname;

    const getuser = await fetch('https://discord.com/api/users/@me', {
        headers: {
            authorization: `Bearer ${access_token}`,
        },
    })

    const userdata = await getuser.json() as DiscordUserDetail;
    await saveUser(userdata, krname);
    const token = createToken(userdata.id);
    
    res.json({
        status: "complete",
        userId: userdata.id,
        username: userdata.username,
        global_name: userdata.global_name,
        token
    })
})

app.post("/registerKakao", async (req, res) => {
    const data = req.body;

    const krname = data.krname;

    const userData = {
        id: String(Math.floor(Number(data.userId))),
        username: data.nickname,
        global_name: data.nickname,
    }

    await saveUser(userData, krname);
    const token = createToken(userData.id);
    
    res.json({
        status: "complete",
        userId: userData.id,
        username: userData.username,
        global_name: userData.global_name,
        token
    })
})

app.get("/getIntents", async (req, res) => {
    const userId = String(req.query.userId);

    const user = await getUser(userId);
    if(user){
        res.json({
            intents: user.intents
        })
    }else{
        res.json({
            intents: []
        })
    }
})

app.get("/checkAdmin", async (req, res) => {
    const token = req.query.token as string;

    if(!token){
        res.status(500).end("token required");
        return;
    }

    let decoded: jwt.JwtPayload;
    try{
        decoded = jwt.verify(token, private_key) as jwt.JwtPayload;
    }catch(err){
        res.status(500).end("wtf is this token");
        return;
    }
    const user = await getUser(decoded.userId);

    if(!user){
        res.status(404);
        res.end("not found");
        return;
    }

    if(user.intents.includes("ADMIN")){
        res.json({
            isAdmin: true
        });
    }else{
        res.json({
            isAdmin: false
        });
    }

})

app.get("/getAllUsers", async (req, res) => {
    const users = await getAllUsers();
    res.json({
        users
    })
})

app.get("/getActivityLog", async (req, res) => {
    const logs = await getActivityLogs();
    res.json({
        data: logs
    })
})

app.get("/checkIntent", async (req, res) => {
    const intent = req.query.intent as intentList;
    const token = req.query.token as string;

    if(!token){
        res.status(500).end("token required");
        return;
    }

    let decoded: jwt.JwtPayload;
    try{
        decoded = jwt.verify(token, private_key) as jwt.JwtPayload;
    }catch(err){
        res.status(500).end("wtf is this token");
        return;
    }
    const user = await getUser(decoded.userId);

    if(!user){
        res.json({
            status: false
        });
        return;
    }

    if(user.intents.includes(intent) || user.intents.includes("ADMIN")){
        res.json({
            status: true
        });
    }else{
        res.json({
            status: false
        });
    }
})

app.get("/authorize", async (req, res) => {
    const target = req.query.userId as string;
    const intent = req.query.intent as intentList;
    const token = req.query.token as string;

    if(!token){
        res.status(500).end("token required");
        return;
    }

    let decoded: jwt.JwtPayload;
    try{
        decoded = jwt.verify(token, private_key) as jwt.JwtPayload;
    }catch(err){
        res.status(500).end("wtf is this token");
        return;
    }

    const user = await getUser(decoded.userId);

    if(!user){
        res.status(404);
        res.end("not found");
        return;
    }

    if(user.intents.includes("ADMIN")){
        await editIntent(target, intent);
        res.end("complete");
    }else{
        res.status(404);
        res.end("not found");
        return;
    }
})

app.get("/unauthorize", async (req, res) => {
    const target = req.query.userId as string;
    const intent = req.query.intent as intentList;
    const token = req.query.token as string;

    if(!token){
        res.status(500).end("token required");
        return;
    }

    let decoded: jwt.JwtPayload;
    try{
        decoded = jwt.verify(token, private_key) as jwt.JwtPayload;
    }catch(err){
        res.status(500).end("wtf is this token");
        return;
    }

    const user = await getUser(decoded.userId);

    if(!user){
        res.status(404);
        res.end("not found");
        return;
    }

    if(user.intents.includes("ADMIN")){
        await editIntent(target, intent);
        res.end("complete");
    }else{
        res.status(404);
        res.end("not found");
        return;
    }
})

app.post("/requestAdminIntent", async (req, res) => {
    const token = req.query.token as string;

    const data = req.body;
    const password = data.pwd as string;

    if(!token){
        res.status(500).end("token required");
        return;
    }

    let decoded: jwt.JwtPayload;
    try{
        decoded = jwt.verify(token, private_key) as jwt.JwtPayload;
    }catch(err){
        res.status(500).end("wtf is this token");
        return;
    }

    if(password == admin_passowrd){
        await editIntent(decoded.userId, "ADMIN");
        res.status(200).end("complete");
    }else{
        res.status(500).end("error");
    }
})

app.post("/log", async (req, res) => {
    const data = <logdata>req.body;

    await insertLog(data);

    res.end("complete");
})

app.get("/downloadZip", async (req, res) => {
    const token = req.query.token as string;
    const zipPath = decodeURIComponent(req.query.zipPath as string);
    if (!token) {
        res.status(500).end("token required");
        return;
    }
    let decoded: jwt.JwtPayload;
    try {
        decoded = jwt.verify(token, private_key) as jwt.JwtPayload;
    } catch (err) {
        res.status(500).end("wtf is this token");
        return;
    }
    if (!zipPath || !zipPath.startsWith("../../nas-data/")) {
        res.status(400).end("invalid path");
        return;
    }
    res.download(zipPath, path.basename(zipPath));
});

app.get("/deleteTempZip", async (req, res) => {
    const token = req.query.token as string;
    const pathToDelete = decodeURIComponent(req.query.path as string);
    if (!token) {
        res.status(500).end("token required");
        return;
    }
    let decoded: jwt.JwtPayload;
    try {
        decoded = jwt.verify(token, private_key) as jwt.JwtPayload;
    } catch (err) {
        res.status(500).end("wtf is this token");
        return;
    }
    try {
        if (pathToDelete && pathToDelete.startsWith("../../nas-data/")) {
            await fs.unlink(pathToDelete);
            res.end("complete");
        } else {
            res.status(400).end("invalid path");
        }
    } catch (err) {
        res.status(500).end("delete failed");
    }
});

const server = app.listen(PORT, () => {
    console.log(`server is running at port : ${PORT}`);
})