package server

import (
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/go-chi/chi/v5/middleware"
)

// 공개 다운로드 토큰과 token 쿼리 값은 그것만으로 파일을 받거나 로그인할 수 있으므로 접근 기록에 남기지 않음.
// 요청 줄(RequestURI)은 절대 주소 형식이나 퍼센트 인코딩으로 모양을 바꿀 수 있어서, 해석된 경로와 쿼리로 기록할 주소를 다시 만듦
var shareTokenInPath = regexp.MustCompile(`(?i)(^|/)d/[^/]+`)

type redactingLogFormatter struct {
	middleware.DefaultLogFormatter
}

func (f *redactingLogFormatter) NewLogEntry(r *http.Request) middleware.LogEntry {
	shown := r.WithContext(r.Context())
	shown.RequestURI = loggedURI(r)
	return f.DefaultLogFormatter.NewLogEntry(shown)
}

func loggedURI(r *http.Request) string {
	uri := shareTokenInPath.ReplaceAllString(r.URL.EscapedPath(), "${1}d/REDACTED")
	query := r.URL.Query()
	if len(query) == 0 {
		return uri
	}
	for key := range query {
		if strings.EqualFold(key, "token") {
			query[key] = []string{"REDACTED"}
		}
	}
	return uri + "?" + query.Encode()
}

// accessLog는 chi middleware.Logger와 같은 형식으로 쓰되 토큰만 가림
func accessLog() func(http.Handler) http.Handler {
	return middleware.RequestLogger(&redactingLogFormatter{
		DefaultLogFormatter: middleware.DefaultLogFormatter{Logger: log.New(os.Stdout, "", log.LstdFlags)},
	})
}
