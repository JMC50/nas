package server

import (
	"log"
	"net/http"
	"os"
	"regexp"

	"github.com/go-chi/chi/v5/middleware"
)

// 공개 다운로드 토큰과 쿼리의 token 값은 그것만으로 파일을 받거나 로그인할 수 있으므로 접근 기록에 남기지 않음
var (
	shareTokenInPath = regexp.MustCompile(`^((?:/+server)?/+d/)[^/?#]+`)
	tokenInQuery     = regexp.MustCompile(`([?&]token=)[^&#]*`)
)

type redactingLogFormatter struct {
	middleware.DefaultLogFormatter
}

func (f *redactingLogFormatter) NewLogEntry(r *http.Request) middleware.LogEntry {
	shown := r.WithContext(r.Context())
	shown.RequestURI = redactURI(r.RequestURI)
	return f.DefaultLogFormatter.NewLogEntry(shown)
}

func redactURI(uri string) string {
	uri = shareTokenInPath.ReplaceAllString(uri, "${1}REDACTED")
	return tokenInQuery.ReplaceAllString(uri, "${1}REDACTED")
}

// accessLog는 chi middleware.Logger와 같은 형식으로 쓰되 토큰만 가림
func accessLog() func(http.Handler) http.Handler {
	return middleware.RequestLogger(&redactingLogFormatter{
		DefaultLogFormatter: middleware.DefaultLogFormatter{Logger: log.New(os.Stdout, "", log.LstdFlags)},
	})
}
