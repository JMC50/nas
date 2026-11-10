//go:build unix

package share

import (
	"os"
	"syscall"
)

// FIFO를 열 때 쓰는 쪽을 기다리며 멈추지 않게 비차단으로 엶. 보통 파일 읽기에는 영향 없음
const readFlags = os.O_RDONLY | syscall.O_NONBLOCK
