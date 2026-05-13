# NAS

브라우저로 접속하는 셀프호스트 파일 서버. Go 백엔드와 SvelteKit 프론트엔드, Docker 이미지 한 개로 배포합니다.

> English version: [README_EN.md](README_EN.md)

![File explorer](Docs/screenshots/02-explorer.png)

## 무엇을 할 수 있나

- **파일 관리** — 폴더 탐색, 업/다운로드, 복사·이동·이름변경·삭제, ZIP 압축/해제
- **재개 가능한 업로드** — [tus 프로토콜](https://tus.io) (`tusd/v2`) 기반, 단일 파일 상한은 `MAX_FILE_SIZE`(기본 50GB)
- **에디터 내장** — Monaco 에디터(Gruvbox 테마)로 텍스트/코드 파일 인라인 편집
- **인증** — 로컬 계정(bcrypt) + Discord OAuth + Google OAuth, 백엔드가 단일 출처
- **권한 모델** — VIEW · OPEN · DOWNLOAD · UPLOAD · COPY · DELETE · RENAME · ADMIN 8종을 사용자별로 토글
- **운영 도구** — 시스템 지표 대시보드(CPU/메모리/디스크/업타임), 활동 로그, 관리자용 OAuth 자격증명 UI

## 기술 스택

| 영역 | 사용 기술 |
|------|----------|
| 백엔드 | Go 1.25 · chi v5 · modernc/sqlite(순수 Go) · tusd v2 · gopsutil v3 · golang-jwt v5 |
| 프론트엔드 | SvelteKit (adapter-static) · Svelte 5 runes · Tailwind 4 · Monaco editor · Vite 6 |
| 디자인 | Gruvbox dark/light · `mode-watcher` · `lucide-svelte` 아이콘 |
| 배포 | 멀티스테이지 Docker(Alpine), Watchtower 옵셔널 자동 업데이트 |
| 저장소 | SQLite 단일 파일, 시작 시 스키마 자동 생성·검증 |

단일 바이너리 + 빌드된 프론트엔드를 Alpine 이미지 하나에 담아 배포합니다. CGO 없이 빌드되므로 크로스 컴파일이 단순합니다.

## 빠른 시작

### Docker로 실행 (권장)

```bash
git clone https://github.com/<owner>/nas.git
cd nas
cp .env.example .env
# .env 에서 PRIVATE_KEY, ADMIN_PASSWORD 만 수정해도 동작
docker compose up -d
```

`http://localhost:7777` 에서 회원가입 후 사용. 헤더 우측의 계정 아이콘 → **Request admin** 으로 관리자 권한을 신청할 수 있습니다(`ADMIN_PASSWORD` 와 일치하는 비밀번호 입력 시 즉시 부여). 관리자 화면에 직접 접근하는 경우에도 동일한 신청 패널이 403 화면에서 노출됩니다.

### 로컬 개발

```bash
# 백엔드
cd backend
go run ./cmd/server

# 프론트엔드 (다른 터미널)
cd frontend
npm install
npm run dev    # Vite 가 백엔드 7777 로 /server/* 프록시
```

## 스크린샷

### 파일 탐색기

루트 그리드 뷰와 하위 폴더 진입 모습. VSCode 패턴의 좌측 네비게이션, 상단 탭, 하단 상태바로 구성됩니다.

| 루트 | 하위 폴더 |
|------|----------|
| ![Root](Docs/screenshots/02-explorer.png) | ![Folder](Docs/screenshots/03-explorer-folder.png) |

### 시스템 대시보드

`gopsutil` 기반 실시간 지표. 5초마다 폴링하여 60샘플(=5분) 슬라이딩 윈도우를 스파크라인으로 보여줍니다.

![System metrics](Docs/screenshots/04-system.png)

### 사용자/권한 관리

관리자가 사용자별로 8개 인텐트를 토글합니다. ADMIN 인텐트는 그 자체로 관리자 화면 접근 권한입니다.

![User permissions](Docs/screenshots/05-users.png)

### OAuth 설정 (관리자)

Discord/Google 자격증명을 런타임에 DB에 저장합니다. 프론트엔드는 빌드 시점 환경 변수에 의존하지 않습니다.

![Settings — Server](Docs/screenshots/06-settings-server.png)

### Quick Open (`Ctrl+P`)

열린 탭과 등록된 화면을 한 번에 검색합니다.

![Quick Open](Docs/screenshots/07-quickopen.png)

### 로그인

OAuth 제공자가 설정된 경우 로컬 로그인 위에 함께 노출됩니다.

![Sign-in](Docs/screenshots/01-login.png)

## 환경 변수

| 변수 | 기본값 | 설명 |
|------|--------|------|
| `PORT` | `7777` | HTTP 리스닝 포트 |
| `DATA_PATH` | `./data` | 호스트 측 데이터 루트 (Docker 마운트 기준) |
| `PRIVATE_KEY` / `JWT_SECRET` | (필수) | JWT 서명 키. 32바이트 이상 권장 |
| `ADMIN_PASSWORD` | (필수) | `Request admin` 호출 시 검증되는 비밀번호 |
| `AUTH_TYPE` | `both` | `local`, `oauth`, `both` 중 선택 |
| `CORS_ORIGIN` | `*` | CORS 허용 오리진 |
| `MAX_FILE_SIZE` | `50gb` | 단일 업로드 상한 |
| `DISCORD_CLIENT_ID/SECRET/REDIRECT_URI` | — | OAuth 부트스트랩용. 이후 관리자 UI 로 덮어쓰기 가능 |
| `GOOGLE_CLIENT_ID/SECRET/REDIRECT_URI` | — | 동일 |
| `TZ` | `UTC` | 컨테이너 타임존 |

전체 목록은 [`.env.example`](.env.example) 을 참고하세요.

## 아키텍처

```
┌──────────────────────────────────────────────────────┐
│                     Browser (SPA)                     │
│  SvelteKit static · Tailwind · Monaco · Svelte 5      │
└──────────────────────────────────────────────────────┘
                          │  HTTP/JSON, tus
                          ▼
┌──────────────────────────────────────────────────────┐
│                Go server (single binary)              │
│  chi router · JWT middleware · intent middleware      │
│  ├─ /auth/*        local & OAuth                      │
│  ├─ /files, /readFolder, /saveTextFile …  파일 조작    │
│  ├─ /files/*       tus 재개 업로드                     │
│  ├─ /getVideoData, /getAudioData, /download  스트리밍 │
│  ├─ /admin/oauth-config, /authorize  관리자           │
│  ├─ /getSystemInfo  gopsutil 메트릭                   │
│  └─ /  SPA(adapter-static) 정적 서빙                  │
└──────────────────────────────────────────────────────┘
        │                                       │
        ▼                                       ▼
┌────────────────────┐                ┌────────────────────┐
│  SQLite (modernc)  │                │  filesystem        │
│  users, intents,   │                │  /data/nas         │
│  activity_log,     │                │  /data/nas-admin   │
│  oauth_config      │                │  /tmp/nas (tus 스테이징) │
└────────────────────┘                └────────────────────┘
```

라우터 전체 정의는 [`backend/internal/server/router.go`](backend/internal/server/router.go) 한 파일에 모여 있습니다.

## 개발

```bash
# 백엔드 테스트 (integration 포함)
cd backend
go test ./...

# 프론트엔드 타입 체크
cd frontend
npm run check

# 전체 프로덕션 빌드
cd backend && go build -o bin/server ./cmd/server
cd frontend && npm run build
```

자세한 문서: [`Docs/`](Docs/README.md).

## 자동 업데이트 (옵션)

Watchtower 가 5분마다 GHCR 이미지 변경을 감지해 무중단 롤링 재시작을 수행합니다. 기본은 비활성화이며 명시적으로 켭니다.

```bash
docker compose --profile autoupdate up -d
```

GitHub Actions(`.github/workflows/build-and-deploy.yml`)가 main 브랜치 push 마다 이미지를 빌드해 `ghcr.io/<owner>/nas:latest` 로 푸시합니다. 본인 계정에서 사용하려면 저장소를 fork 한 뒤 `GITHUB_REPOSITORY` 환경 변수와 GHCR 패키지 가시성을 본인 것으로 맞춰 주세요.

## 라이선스

MIT.
