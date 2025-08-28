# 🚀 오픈소스 NAS 시스템

완전 자동화된 Docker 기반 NAS 파일 관리 시스템

## ⚡ 주요 특징

- **🔥 원클릭 설치**: `docker-compose up -d` 한 번으로 완료
- **🔄 자동 업데이트**: 새 버전 릴리스 시 자동 배포
- **🐋 경량 Alpine**: 250MB 이하 초경량 이미지
- **🔒 보안 강화**: JWT 인증 + non-root 실행
- **📱 반응형 웹UI**: 모든 디바이스에서 접근 가능

## 🚨 중요: Fork 필수!

이 프로젝트를 사용하려면 **반드시 본인 계정으로 Fork**해야 합니다.

### 왜 Fork가 필요한가요?

- 🔧 **본인만의 이미지**: 각자의 GitHub Container Registry 사용
- 💰 **비용 절약**: 원본 저장소 대역폭 비용 방지
- 🎛️ **자유로운 커스터마이징**: 개인 요구사항에 맞게 수정 가능
- 🔄 **독립적인 업데이트**: 본인 일정에 맞춰 업데이트 관리

## 📋 설치 가이드

### 1단계: Repository Fork
```bash
# GitHub에서 이 저장소를 본인 계정으로 Fork
# https://github.com/original-author/nas → Fork 버튼 클릭
```

### 2단계: Fork된 저장소 클론
```bash
git clone https://github.com/YOUR-USERNAME/nas.git
cd nas
```

### 3단계: 환경 설정
```bash
# .env 파일 생성
cp .env.example .env

# 필수 설정 수정
vim .env
```

**수정 필요한 항목들:**
```bash
# 본인의 GitHub 저장소로 변경 (중요!)
GITHUB_REPOSITORY=YOUR-USERNAME/nas

# 시크릿 키 변경 (보안 필수!)
JWT_SECRET=your-random-64-character-string

# 관리자 비밀번호 변경
ADMIN_PASSWORD=your-secure-password

# 데이터 저장 경로
DATA_PATH=./data
```

### 4단계: 원클릭 설치
```bash
# 자동 설치 스크립트 실행
chmod +x scripts/setup.sh
./scripts/setup.sh

# 또는 직접 실행
docker-compose up -d
```

### 5단계: 접속 확인
```bash
# 웹 인터페이스 접속
http://localhost:7777

# 상태 확인
docker-compose ps
```

## 🔄 자동 업데이트 시스템

### 작동 원리
1. **본인이 Fork한 저장소**에 코드 push
2. **GitHub Actions**가 자동으로 이미지 빌드
3. **본인의 GHCR**에 이미지 저장
4. **Watchtower**가 5분마다 체크하여 자동 업데이트

### 업데이트 흐름
```bash
# 개발자 (본인)
git add . && git commit -m "feature: 새 기능 추가"
git push origin main

# 5분 후 자동으로...
# 1. GitHub Actions 빌드 시작
# 2. 새 이미지가 ghcr.io/YOUR-USERNAME/nas:latest로 푸시
# 3. 운영 중인 모든 서버에서 Watchtower가 감지
# 4. 자동으로 무중단 업데이트 완료 ✨
```

## 🛠️ 관리 명령어

```bash
# 로그 확인
docker-compose logs -f

# 서비스 재시작
docker-compose restart

# 수동 업데이트
docker-compose pull && docker-compose up -d

# 서비스 중지
docker-compose down

# 전체 업그레이드 (스크립트 사용)
./scripts/setup.sh --upgrade
```

## 📊 시스템 요구사항

- **OS**: Linux, macOS, Windows (Docker 지원 환경)
- **RAM**: 최소 512MB, 권장 1GB
- **Storage**: 최소 1GB (데이터 별도)
- **Docker**: 20.10+ 
- **Docker Compose**: 2.0+

## 🔧 고급 설정

### 포트 변경
```bash
# .env 파일에서
PORT=8080
```

### 데이터 경로 변경
```bash
# .env 파일에서
DATA_PATH=/mnt/nas-storage
```

### 업데이트 주기 변경
```bash
# .env 파일에서 (초 단위)
WATCHTOWER_POLL_INTERVAL=1800  # 30분마다
```

### Watchtower 비활성화
```bash
# docker-compose.yml에서 watchtower 서비스 주석 처리
# watchtower:
#   image: containrrr/watchtower:latest
#   ...
```

## 🐛 문제 해결

### 이미지를 찾을 수 없음
```bash
# .env에서 GITHUB_REPOSITORY 확인
GITHUB_REPOSITORY=YOUR-USERNAME/nas  # 정확한 저장소명

# GitHub Container Registry가 public인지 확인
# GitHub → 본인 저장소 → Packages → nas → Package settings → Change visibility
```

### 자동 업데이트 안됨
```bash
# Watchtower 로그 확인
docker-compose logs watchtower

# 수동으로 업데이트 테스트
docker-compose pull
```

### 포트 충돌
```bash
# .env에서 다른 포트 사용
PORT=8080

# 재시작
docker-compose down && docker-compose up -d
```

## 🤝 기여 가이드

1. 이슈 등록 또는 기능 요청
2. 본인 Fork에서 개발 브랜치 생성
3. 기능 개발 및 테스트
4. 원본 저장소에 Pull Request

## 📄 라이선스

MIT License - 자세한 내용은 [LICENSE](LICENSE) 참조

## ⚠️ 주의사항

- **보안**: JWT_SECRET과 ADMIN_PASSWORD를 반드시 변경하세요
- **백업**: 정기적으로 데이터 백업을 수행하세요
- **모니터링**: 시스템 리소스 사용량을 주기적으로 확인하세요
- **업데이트**: 주요 업데이트 전에는 데이터 백업을 권장합니다

---

## 📚 고급 기능 및 개발자 문서

이 간단한 설치 가이드 외에도 다음과 같은 고급 기능들이 제공됩니다:

- **🔐 OAuth 인증**: Discord, Kakao 등 소셜 로그인 연동
- **👥 사용자 관리**: 권한 기반 접근 제어 시스템  
- **🎨 프론트엔드 개발**: Svelte 5 + TypeScript 아키텍처
- **🛠️ 백엔드 API**: Express.js + SQLite 완전한 REST API
- **🚀 다양한 배포 방식**: PM2, systemd, 수동 설치 옵션

자세한 내용은 **[📖 완전한 문서](Docs/README.md)**를 참조하세요.

## 💡 도움이 필요하신가요?

- 📚 **완전한 문서**: [Docs 폴더](Docs/README.md) - 모든 기능과 설정 가이드
- 🐛 **버그 리포트**: [Issues](../../issues) 등록
- 💬 **질문**: [Discussions](../../discussions) 활용
- 🌐 **English Version**: [README_EN.md](README_EN.md)

**즐거운 NAS 라이프 되세요! 🎉**