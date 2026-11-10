# 시험용 NAS

이 폴더는 포크 이미지를 운영 NAS와 떨어진 곳에서 확인하는 시험용 NAS를 띄웁니다. 시험용 NAS는 `127.0.0.1`에만 묶이고, 이 폴더의 `data/`만 씁니다. 운영 NAS의 `/home/anxi/nas/data`는 읽지도 쓰지도 않습니다.

## 실행

```bash
cd deploy/staging
./up.sh nas-signal:<커밋>          # 127.0.0.1:7778 에 시험용 NAS 시작
python3 -I seed.py                  # 시험 계정, 시험 파일, 시험 활동 기록
docker rm -f nas-signal-staging     # 끄기
```

다른 포트를 쓰려면 `STAGING_PORT=7779 ./up.sh ...`처럼 환경 변수를 줍니다. 저장소 밖에서 돌릴 때는 `NAS_TESTDATA`로 `backend/tests/testdata` 위치를 알려 줍니다.

## 시험 계정과 값

모든 값은 시험 전용입니다. 운영 비밀번호나 실제 OAuth 앱 값이 아닙니다.

| 파일 | 내용 |
| --- | --- |
| `run/staging.env` | 처음 실행 때 만든 시험용 `PRIVATE_KEY`, `ADMIN_PASSWORD`, 가짜 Discord와 Google 클라이언트 값 |
| `run/test-accounts.env` | 시험 계정 4개(`owner`, `member`, `viewer`, `longname`)의 무작위 비밀번호 |

두 파일은 권한 600으로 만들어지고 저장소에 들어가지 않습니다(`.gitignore`). 가짜 Discord와 Google 값은 로그인 화면의 버튼 모양을 보려는 것이므로 버튼을 눌러도 로그인되지 않습니다.

| 계정 | 권한 | 쓰임 |
| --- | --- | --- |
| `owner` | 관리자와 모든 권한 | 관리 화면, 설정, 파일 작업 |
| `member` | VIEW, OPEN, DOWNLOAD, UPLOAD, RENAME | 일반 팀원 화면 |
| `viewer` | 없음 | 권한 없음 화면 |
| `longname` | VIEW | 긴 한글 이름 표시 확인 |
