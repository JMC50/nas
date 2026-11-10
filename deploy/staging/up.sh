#!/usr/bin/env bash
# 시험용 NAS를 127.0.0.1에만 띄움. 운영 데이터와 떨어진 이 폴더의 data/ 를 씀
# 사용: ./up.sh <이미지>   예: ./up.sh nas:abc1234
set -euo pipefail
cd "$(dirname "$0")"
IMAGE=${1:?usage: up.sh <image>}
NAME=${STAGING_NAME:-nas-staging}
PORT=${STAGING_PORT:-7778}
RUN=$PWD/run
umask 077
mkdir -p "$RUN" data/nas data/nas-admin data/db

# 시험 값만 씀. 실제 OAuth 앱이 아니므로 Discord, Google 버튼은 모양 확인용
if [ ! -f "$RUN/staging.env" ]; then
  {
    echo "PRIVATE_KEY=$(openssl rand -hex 32)"
    echo "ADMIN_PASSWORD=$(openssl rand -hex 16)"
    echo "AUTH_TYPE=both"
    echo "PASSWORD_MIN_LENGTH=8"
    echo "DISCORD_CLIENT_ID=staging-discord-client"
    echo "DISCORD_REDIRECT_URI=http://127.0.0.1:$PORT/login"
    echo "GOOGLE_CLIENT_ID=staging-google-client"
    echo "GOOGLE_CLIENT_SECRET=staging-google-secret"
    echo "GOOGLE_REDIRECT_URI=http://127.0.0.1:$PORT/googleLogin"
    echo "TZ=Asia/Seoul"
  } > "$RUN/staging.env"
fi

docker rm -f "$NAME" >/dev/null 2>&1 || true
docker run -d --name "$NAME" -p "127.0.0.1:$PORT:7777" --env-file "$RUN/staging.env" \
  -v "$PWD/data/nas:/data/nas" -v "$PWD/data/nas-admin:/data/nas-admin" -v "$PWD/data/db:/data/db" \
  "$IMAGE" >/dev/null
for _ in $(seq 1 30); do
  curl -fs -o /dev/null "http://127.0.0.1:$PORT/healthz" && break
  sleep 1
done
curl -fsS "http://127.0.0.1:$PORT/healthz"
echo
