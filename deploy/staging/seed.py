"""시험용 NAS에 시험 계정, 시험 파일, 시험 활동 기록을 넣는다.

사용: python3 -I seed.py
up.sh 와 같은 폴더에서 돌린다. 비밀번호는 run/test-accounts.env 에만 남고 출력하지 않는다.
다시 돌려도 이미 있는 계정, 권한, 파일, 기록은 그대로 둔다.
"""

import json
import math
import os
import random
import secrets
import shutil
import struct
import sys
import time
import urllib.error
import urllib.parse
import urllib.request
import wave
import zipfile
import zlib
from pathlib import Path

HERE = Path(__file__).resolve().parent
RUN = HERE / "run"
DATA = HERE / "data" / "nas"
PORT = os.environ.get("STAGING_PORT", "7778")
BASE = f"http://127.0.0.1:{PORT}"
TESTDATA = Path(os.environ.get("NAS_TESTDATA", HERE.parent.parent / "backend" / "tests" / "testdata"))

ALL_INTENTS = ["VIEW", "OPEN", "DOWNLOAD", "UPLOAD", "COPY", "DELETE", "RENAME"]
USERS = [
    # (userId, username, krname, intents, admin)
    ("owner", "owner", "김관리", ALL_INTENTS, True),
    ("member", "member", "이팀원", ["VIEW", "OPEN", "DOWNLOAD", "UPLOAD", "RENAME"], False),
    ("viewer", "viewer", "박열람", [], False),
    ("longname", "longname-display-name", "아주긴한글이름을가진시험용팀원계정", ["VIEW"], False),
]


def read_env(path):
    values = {}
    for line in path.read_text().splitlines():
        if "=" in line and not line.startswith("#"):
            key, value = line.split("=", 1)
            values[key] = value
    return values


def ensure_passwords():
    path = RUN / "test-accounts.env"
    if not path.exists():
        lines = [f"TEST_{user[0].upper()}_PASSWORD={secrets.token_hex(12)}" for user in USERS]
        old = os.umask(0o077)
        try:
            path.write_text("\n".join(lines) + "\n")
        finally:
            os.umask(old)
    return read_env(path)


def call(method, path, body=None, token=None):
    data = json.dumps(body).encode() if body is not None else None
    request = urllib.request.Request(BASE + path, data=data, method=method)
    if data is not None:
        request.add_header("Content-Type", "application/json")
    if token:
        request.add_header("Authorization", f"Bearer {token}")
    try:
        with urllib.request.urlopen(request) as response:
            raw = response.read()
            return response.status, raw
    except urllib.error.HTTPError as error:
        return error.code, error.read()


def sign_in(user_id, username, krname, password):
    status, raw = call("POST", "/auth/login", {"userId": user_id, "password": password})
    if status == 200:
        return json.loads(raw)["token"]
    status, raw = call(
        "POST",
        "/auth/register",
        {"userId": user_id, "username": username, "password": password, "krname": krname},
    )
    if status != 200:
        sys.exit(f"register {user_id} failed: {status}")
    return json.loads(raw)["token"]


def current_intents(owner_token):
    status, raw = call("GET", "/getAllUsers", token=owner_token)
    if status != 200:
        sys.exit(f"user list failed: {status}")
    return {user["userId"]: set(user["intents"]) for user in json.loads(raw)["users"]}


def seed_accounts(admin_password, passwords):
    tokens = {}
    for user_id, username, krname, _, _ in USERS:
        tokens[user_id] = sign_in(user_id, username, krname, passwords[f"TEST_{user_id.upper()}_PASSWORD"])

    owner = tokens["owner"]
    status, raw = call("GET", "/checkAdmin", token=owner)
    if status != 200 or not json.loads(raw)["isAdmin"]:
        status, _ = call("POST", "/requestAdminIntent", {"pwd": admin_password}, token=owner)
        if status != 200:
            sys.exit(f"admin grant failed: {status}")

    granted = current_intents(owner)
    for user_id, _, _, intents, _ in USERS:
        for intent in intents:
            if intent not in granted.get(user_id, set()):
                query = urllib.parse.urlencode({"userId": user_id, "intent": intent})
                status, _ = call("GET", f"/authorize?{query}", token=owner)
                if status != 200:
                    sys.exit(f"grant {intent} to {user_id} failed: {status}")
    return tokens


def png(path, width, height, top, bottom):
    rows = bytearray()
    for y in range(height):
        mix = y / max(1, height - 1)
        color = bytes(int(top[i] + (bottom[i] - top[i]) * mix) for i in range(3))
        rows.append(0)
        rows.extend(color * width)

    def chunk(kind, payload):
        body = kind + payload
        return struct.pack(">I", len(payload)) + body + struct.pack(">I", zlib.crc32(body) & 0xFFFFFFFF)

    header = struct.pack(">IIBBBBB", width, height, 8, 2, 0, 0, 0)
    path.write_bytes(
        b"\x89PNG\r\n\x1a\n" + chunk(b"IHDR", header) + chunk(b"IDAT", zlib.compress(bytes(rows), 9)) + chunk(b"IEND", b"")
    )


def wav(path, seconds, frequency):
    rate = 22050
    with wave.open(str(path), "wb") as out:
        out.setnchannels(1)
        out.setsampwidth(2)
        out.setframerate(rate)
        frames = bytearray()
        for n in range(int(rate * seconds)):
            fade = min(1.0, n / 2000, (rate * seconds - n) / 2000)
            frames += struct.pack("<h", int(9000 * fade * math.sin(2 * math.pi * frequency * n / rate)))
        out.writeframes(bytes(frames))


def pdf(path, pages):
    objects = ["<< /Type /Catalog /Pages 2 0 R >>", None, "<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica >>"]
    kids = []
    for number, lines in enumerate(pages):
        text = "BT /F1 20 Tf 72 720 Td " + " ".join(f"({line}) Tj 0 -28 Td" for line in lines) + " ET"
        objects.append(f"<< /Length {len(text)} >>\nstream\n{text}\nendstream")
        content = len(objects)
        objects.append(f"<< /Type /Page /Parent 2 0 R /MediaBox [0 0 612 792] /Contents {content} 0 R /Resources << /Font << /F1 3 0 R >> >> >>")
        kids.append(f"{len(objects)} 0 R")
    objects[1] = f"<< /Type /Pages /Kids [{' '.join(kids)}] /Count {len(kids)} >>"
    out = bytearray(b"%PDF-1.4\n")
    offsets = []
    for index, body in enumerate(objects, start=1):
        offsets.append(len(out))
        out += f"{index} 0 obj\n{body}\nendobj\n".encode()
    xref = len(out)
    out += f"xref\n0 {len(objects) + 1}\n0000000000 65535 f \n".encode()
    for offset in offsets:
        out += f"{offset:010d} 00000 n \n".encode()
    out += f"trailer\n<< /Size {len(objects) + 1} /Root 1 0 R >>\nstartxref\n{xref}\n%%EOF\n".encode()
    path.write_bytes(bytes(out))


def write(path, text):
    path.write_text(text, encoding="utf-8")


def seed_files():
    marker = RUN / ".files-seeded"
    if marker.exists():
        return
    docs = DATA / "문서"
    photos = DATA / "사진"
    music = DATA / "음악"
    videos = DATA / "동영상"
    code = DATA / "코드"
    logs = DATA / "로그 보관"
    for folder in (docs, photos, music, videos, code, logs, DATA / "빈 폴더", DATA / "올리기 시험"):
        folder.mkdir(parents=True, exist_ok=True)

    long_doc = "2026 프로젝트 회의록 최종 수정본 (팀 공유용) 버전 3 - 검토 의견 반영 완료.docx"
    shutil.copy(TESTDATA / "sample.docx", docs / long_doc)
    shutil.copy(TESTDATA / "sample.pptx", docs / "발표 자료.pptx")
    shutil.copy(TESTDATA / "sample.xlsx", docs / "예산표 2026.xlsx")
    pdf(docs / "설명서.pdf", [["Staging test document", "Page 1 of 2"], ["Staging test document", "Page 2 of 2"]])
    write(docs / "메모.txt", "시험용 메모입니다.\n긴 줄과 짧은 줄을 섞었습니다. 이 파일은 시험 데이터입니다.\n")

    png(photos / "노을 풍경 2026-10-03.png", 1200, 800, (40, 44, 80), (238, 150, 90))
    png(photos / "바다.png", 1200, 800, (12, 60, 110), (120, 200, 220))
    png(photos / "숲길 산책로에서 찍은 아주 긴 이름의 사진 파일 예시입니다.png", 900, 1200, (20, 60, 30), (170, 210, 120))
    png(photos / "grey.png", 400, 400, (90, 90, 90), (200, 200, 200))

    wav(music / "시험 음원 440Hz.wav", 6, 440)
    wav(music / "긴 제목의 시험 음원 - 2026 가을 합주 녹음 세 번째 테이크.wav", 8, 330)
    wav(music / "beep.wav", 3, 660)

    write(code / "README.md", "# 시험 저장소\n\n시험 데이터입니다.\n\n```bash\ngo test ./...\n```\n")
    write(code / "main.go", 'package main\n\nimport "fmt"\n\nfunc main() {\n\tfmt.Println("staging")\n}\n')
    write(code / "config.json", json.dumps({"name": "staging", "port": 7778, "debug": False}, indent=2) + "\n")
    write(code / "app.ts", "export function sum(values: number[]): number {\n  return values.reduce((a, b) => a + b, 0);\n}\n")

    for day in range(1, 41):
        write(logs / f"server-2026-10-{day:02d}.log" if day <= 31 else logs / f"server-2026-11-{day - 31:02d}.log",
              f"2026-10-{min(day, 31):02d} 시험 로그 {day}\n" * (day * 7))

    with zipfile.ZipFile(DATA / "보관 묶음.zip", "w") as archive:
        archive.writestr("a.txt", "시험 a\n")
        archive.writestr("b/c.txt", "시험 c\n")
    write(DATA / "2025-roadmap.txt", "시험용 로드맵 파일입니다.\n")
    marker.write_text("done\n")


def seed_activity(tokens):
    marker = RUN / ".activity-seeded"
    if marker.exists():
        return
    rng = random.Random(20261110)
    kinds = ["UPLOAD", "DOWNLOAD", "VIEW", "OPEN", "RENAME", "DELETE", "COPY"]
    weights = [6, 5, 8, 6, 2, 1, 2]
    names = ["발표 자료.pptx", "설명서.pdf", "바다.png", "main.go", "시험 음원 440Hz.wav", "예산표 2026.xlsx", "메모.txt"]
    now = int(time.time() * 1000)
    for _ in range(140):
        user = rng.choice(["owner", "owner", "member", "longname"])
        kind = rng.choices(kinds, weights)[0]
        when = now - rng.randint(0, 29) * 86_400_000 - rng.randint(0, 86_000_000)
        name = rng.choice(names)
        status, _ = call(
            "POST",
            "/log",
            {"activity": kind, "description": name, "token": tokens[user], "time": when, "loc": "/문서"},
        )
        if status != 200:
            sys.exit(f"activity seed failed: {status}")
    marker.write_text("done\n")


def main():
    staging = read_env(RUN / "staging.env")
    passwords = ensure_passwords()
    tokens = seed_accounts(staging["ADMIN_PASSWORD"], passwords)
    seed_files()
    seed_activity(tokens)
    print("seeded users:", ", ".join(user[0] for user in USERS))


if __name__ == "__main__":
    main()
