# NAS (Network Attached Storage) System

웹 기반의 네트워크 저장소 시스템으로, 파일 관리, 사용자 인증, 권한 관리 기능을 제공합니다.

## ❗ 주의 : 이 시스템은 리눅스 시스템을 기반으로 개발되었습니다.

리눅스 이외의 시스템에선 예상치 못한 오류가 발생할 수 있습니다.

## 📋 목차

- [시스템 개요](#-시스템-개요)
- [주요 기능](#-주요-기능)
- [아키텍처](#️-아키텍처)
- [기술 스택](#️-기술-스택)
- [설치 및 실행](#-설치-및-실행)
- [API 문서](#-api-문서)
- [데이터베이스 스키마](#️-데이터베이스-스키마)
- [권한 시스템](#-권한-시스템)
- [파일 관리 기능](#-파일-관리-기능)
- [설정](#-설정)
- [반응형 디자인](#-반응형-디자인)
- [보안](#-보안)

## 🎯 시스템 개요

이 NAS 시스템은 웹 브라우저를 통해 접근할 수 있는 파일 관리 시스템입니다. 사용자는 웹 인터페이스를 통해 파일을 업로드, 다운로드, 편집, 삭제할 수 있으며, 관리자는 사용자 권한을 관리할 수 있습니다.

### 주요 특징
- **웹 기반 인터페이스**: 브라우저에서 직접 파일 관리
- **실시간 시스템 모니터링**: CPU, 메모리, 디스크 사용량 실시간 확인
- **다중 사용자 지원**: 사용자별 권한 관리
- **반응형 디자인**: PC와 모바일 환경 모두 지원
- **다양한 파일 형식 지원**: 텍스트, 이미지, 비디오, 오디오 파일 처리

## 🚀 주요 기능

### 파일 관리
- **파일 업로드/다운로드**: 드래그 앤 드롭 지원
- **폴더 생성/삭제**: 계층적 디렉토리 구조
- **파일 편집**: Monaco Editor를 통한 텍스트 파일 편집
- **파일 검색**: 전체 파일 시스템에서 검색
- **파일 압축/해제**: ZIP 파일 생성 및 압축 해제
- **파일 복사/이동/이름변경**: 기본적인 파일 조작

### 미디어 지원
- **이미지 뷰어**: JPG, PNG, SVG 등 이미지 파일 표시
- **비디오 플레이어**: MP4 파일 스트리밍 재생
- **오디오 플레이어**: MP3 파일 스트리밍 재생

### 시스템 모니터링
- **실시간 시스템 정보**: CPU, 메모리, 디스크 사용량
- **시스템 업타임**: 서버 가동 시간 표시
- **활동 로그**: 사용자 활동 기록 및 조회

### 사용자 관리
- **소셜 로그인**: Discord, Kakao 로그인 지원
- **권한 관리**: 세분화된 권한 시스템
- **관리자 기능**: 사용자 권한 부여/해제

## 🏗️ 아키텍처

### 백엔드 (Node.js + Express)
```
backend/
├── src/
│   ├── index.ts          # 메인 서버 파일
│   ├── sqlite.ts         # 데이터베이스 연결
│   ├── config/           # 설정 파일
│   ├── functions/        # 비즈니스 로직
│   ├── entity/           # 데이터베이스 엔티티
│   └── db/              # 데이터베이스 초기화
```

### 프론트엔드 (Svelte + Vite)
```
frontend/
├── src/
│   ├── App.svelte        # 메인 앱 컴포넌트
│   └── lib/             # UI 컴포넌트들
│       ├── Explorer.svelte      # 파일 탐색기
│       ├── FileManager.svelte   # 파일 관리자
│       ├── SystemInfo.svelte    # 시스템 정보
│       └── ...
```

## 🛠️ 기술 스택

### 백엔드
- **Node.js**: 서버 런타임
- **Express.js**: 웹 프레임워크
- **TypeScript**: 타입 안전성
- **Better-SQLite3**: 데이터베이스
- **JWT**: 인증 토큰
- **Archiver**: 파일 압축
- **Monaco Editor**: 코드 에디터

### 프론트엔드
- **Svelte**: 반응형 UI 프레임워크
- **Vite**: 빌드 도구
- **TypeScript**: 타입 안전성
- **Sass**: 스타일링
- **Monaco Editor**: 웹 기반 코드 에디터

## 📦 설치 및 실행

### 사전 요구사항
- **Node.js** (v18 이상)
- **npm** 또는 **yarn**
- **Git**

### 1. 저장소 클론
```bash
git clone https://github.com/JMC50/nas
cd nas
```

### 2. 의존성 설치
```bash
# 루트 디렉토리에서 전체 의존성 설치
npm install

# 또는 개별 설치
cd backend && npm install
cd ../frontend && npm install
```

### 3. 환경 설정
백엔드 설정 파일을 수정합니다:
```bash
# backend/src/config/config.ts 파일 수정
export const private_key = "your-secret-key"; # jwt secret key
export const admin_password = "your-admin-password";
export const PORT = your-port; # server port
export const KAKAO_REST_API_KEY = "your-kakao-api-key";
export const KAKAO_REDIRECT_URL = "https://your-frontend-url/kakaoLogin";
export const KAKAO_CLIENT_SECRET = "your-kakao-client-secret";
```

프론트 설정 파일을 수정합니다:
```bash
# frontend/config.local.json 파일 수정
{
    "loginURL": "https://discord.com/oauth2/authorize?client_id=[your-client-id]&response_type=token&redirect_uri=[your-discord-redirection-url]&scope=identify", 
    # note : 디스코드 로그인 url의 redirection url은 https://프론트url/login 이어야하며, respone_type=token 으로 설정해주세요. scope는 identify만 있으면 됩니다.
    "kakaoLoginAPIKEY": "your-kakao-api-key",
    "kakaoLoginRediectURL": "your-kakao-login-redirection-url",
    "serverURL": "https://your-server-url" # note : url 끝에 "/" 를 붙이지 마세요.
}
```

### 4. 개발 모드 실행
```bash
# 전체 시스템 실행 (백엔드 + 프론트엔드)
npm test

# 또는 백엔드만 실행
npm start
```

### 5. 프로덕션 빌드
```bash
# 프론트엔드 빌드
cd frontend
npm run build
```

### 6. 프로덕션 실행
```bash
# 백엔드 서버 실행
cd backend
node dist/index.js

# note : 단순하게 npm start를 사용하셔도 됩니다.
```

### 7. 접속
- **백엔드 API**: `http://localhost:7777`
- **프론트엔드**: `http://localhost:5050` (개발 모드)

## 📡 API 문서

### 인증 관련
- `GET /login` - Discord 로그인
- `GET /kakaoLogin` - Kakao 로그인
- `POST /register` - 사용자 등록
- `GET /checkIntent` - 권한 확인

### 파일 관리
- `GET /readFolder` - 폴더 내용 조회
- `POST /input` - 파일 업로드
- `GET /download` - 파일 다운로드
- `GET /getTextFile` - 텍스트 파일 조회
- `POST /saveTextFile` - 텍스트 파일 저장
- `GET /getImageData` - 이미지 파일 조회
- `GET /getVideoData` - 비디오 파일 스트리밍
- `GET /getAudioData` - 오디오 파일 스트리밍

### 파일 조작
- `GET /forceDelete` - 파일/폴더 삭제
- `GET /copy` - 파일 복사
- `GET /move` - 파일 이동
- `GET /rename` - 파일 이름 변경
- `GET /makedir` - 폴더 생성

### 압축/해제
- `POST /zipFiles` - 파일 압축
- `POST /unzipFile` - 파일 압축 해제
- `GET /progress` - 압축/해제 진행률

### 시스템 정보
- `GET /getSystemInfo` - 시스템 정보 조회
- `GET /stat` - 파일 정보 조회
- `GET /searchInAllFiles` - 파일 검색

### 관리자 기능
- `GET /getAllUsers` - 모든 사용자 조회
- `GET /getActivityLog` - 활동 로그 조회
- `GET /authorize` - 권한 부여
- `GET /unauthorize` - 권한 해제

## 🗄️ 데이터베이스 스키마

### users 테이블
```sql
CREATE TABLE users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    userId TEXT UNIQUE NOT NULL,
    username TEXT NOT NULL,
    global_name TEXT,
    krname TEXT
);
```

### user_intents 테이블
```sql
CREATE TABLE user_intents (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    intent TEXT NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);
```

### log 테이블
```sql
CREATE TABLE log (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    activity TEXT NOT NULL,
    description TEXT,
    user_id INTEGER NOT NULL,
    time INTEGER NOT NULL,
    loc TEXT,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL
);
```

## 🔐 권한 시스템

시스템은 세분화된 권한 시스템을 제공합니다:

### 권한 종류
- **ADMIN**: 모든 권한 (관리자)
- **VIEW**: 파일/폴더 조회
- **OPEN**: 파일 열기 (텍스트 편집, 미디어 재생)
- **DOWNLOAD**: 파일 다운로드
- **UPLOAD**: 파일 업로드, 폴더 생성
- **COPY**: 파일 복사
- **DELETE**: 파일/폴더 삭제
- **RENAME**: 파일/폴더 이름 변경

### 권한 우선순위
1. **ADMIN** 권한이 있으면 모든 작업 가능
2. 각 작업별로 해당 권한 필요
3. 권한이 없으면 작업 거부

## 📁 파일 관리 기능

### 지원 파일 형식
- **텍스트 파일**: Monaco Editor로 편집 가능
- **이미지**: JPG, PNG, SVG 등
- **비디오**: MP4 스트리밍 재생
- **오디오**: MP3 스트리밍 재생
- **압축 파일**: ZIP 압축/해제

### 파일 조작 기능
- **드래그 앤 드롭**: 직관적인 파일 업로드
- **다중 선택**: 여러 파일 동시 선택
- **진행률 표시**: 대용량 파일 처리 시 진행률 표시
- **실시간 미리보기**: 이미지, 텍스트 파일 미리보기

### 검색 기능
- **전체 검색**: 모든 파일에서 검색
- **실시간 결과**: 검색 결과 실시간 표시
- **경로 표시**: 검색된 파일의 전체 경로 표시

## 🔧 설정

### 환경 변수
- `PORT`: 서버 포트 (기본값: 7777)
- `private_key`: JWT 서명 키
- `admin_password`: 관리자 권한 요청 시 비밀번호
- `KAKAO_REST_API_KEY`: Kakao 로그인 API 키
- `KAKAO_REDIRECT_URL`: Kakao 로그인 리다이렉트 URL
- `KAKAO_CLIENT_SECRET`: Kakao 클라이언트 시크릿

### 데이터 디렉토리
- `nas-data/`: 사용자 파일 저장소
- `nas-data-admin/`: 관리자 전용 파일 저장소
- `db/`: SQLite 데이터베이스 파일

## 📱 반응형 디자인

시스템은 다양한 디바이스에서 사용할 수 있도록 반응형으로 설계되었습니다:

- **PC 환경**: 사이드바 + 메인 영역 + 파일 관리자 3단 레이아웃
- **모바일 환경**: 하단 메뉴 + 메인 영역 2단 레이아웃
- **자동 감지**: 화면 크기에 따라 자동으로 레이아웃 변경

## 🔒 보안

- **JWT 토큰**: 사용자 인증 및 세션 관리
- **권한 검증**: 모든 API 요청에 권한 검증
- **파일 경로 검증**: 경로 순회 공격 방지
- **입력 검증**: 사용자 입력 데이터 검증