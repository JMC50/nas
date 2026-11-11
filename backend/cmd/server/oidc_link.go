package main

import (
	"bufio"
	"bytes"
	"database/sql"
	"encoding/csv"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"slices"
	"strings"

	"github.com/JMC50/nas/internal/auth"
	"github.com/JMC50/nas/internal/db"
)

const oidcLinkUsage = `Usage: server oidc-link --map <file.csv> [--from google] [--db <path>] [--dry-run]

Adds the "oidc" sign-in identity to existing users, so they keep their account
after the server changes to AUTH_TYPE=oidc. Each line of the CSV file is

  <ID of the user at the old sign-in provider>,<sub of the same person at the OIDC issuer>

Lines that start with # are ignored. The command can run again with the same file.

  --from     old sign-in provider: google (default), discord or local
  --db       SQLite database file (default: the DB_PATH environment variable)
  --dry-run  report what would change and write nothing
`

var (
	linkSources = []string{"google", "discord", "local"}
	utf8BOM     = []byte{0xEF, 0xBB, 0xBF}
)

// runOIDCLink는 종료 코드를 돌려줍니다. 0은 정상, 1은 충돌이나 실행 오류, 2는 잘못된 인자입니다.
func runOIDCLink(args []string) int {
	flags := flag.NewFlagSet("oidc-link", flag.ContinueOnError)
	flags.Usage = func() { fmt.Fprint(os.Stderr, oidcLinkUsage) }
	mapPath := flags.String("map", "", "")
	from := flags.String("from", "google", "")
	dbPath := flags.String("db", os.Getenv("DB_PATH"), "")
	dryRun := flags.Bool("dry-run", false, "")
	if err := flags.Parse(args); err != nil {
		return 2
	}
	if *mapPath == "" || *dbPath == "" || flags.NArg() > 0 || !slices.Contains(linkSources, *from) {
		flags.Usage()
		return 2
	}
	counts, err := linkFromFile(*dbPath, *mapPath, *from, *dryRun)
	if err != nil {
		fmt.Fprintln(os.Stderr, "oidc-link:", err)
		return 1
	}
	fmt.Printf("%d linked, %d already linked, %d without a NAS user, %d conflicts\n",
		counts[auth.LinkAdded], counts[auth.LinkPresent], counts[auth.LinkNoUser], counts[auth.LinkConflict])
	if *dryRun {
		fmt.Println("Dry run: nothing was written.")
	}
	if counts[auth.LinkConflict] > 0 {
		return 1
	}
	return 0
}

func linkFromFile(dbPath, mapPath, from string, dryRun bool) (map[auth.LinkStatus]int, error) {
	// db.Open은 없는 파일을 새로 만들므로 경로를 잘못 주면 빈 데이터베이스가 생김
	if _, err := os.Stat(dbPath); err != nil {
		return nil, fmt.Errorf("database: %w", err)
	}
	file, err := os.Open(mapPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	conn, err := db.Open(dbPath)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	reader := csv.NewReader(withoutBOM(file))
	reader.Comment = '#'
	reader.FieldsPerRecord = 2
	reader.TrimLeadingSpace = true
	return linkRecords(conn, reader, from, dryRun)
}

// 엑셀의 "CSV UTF-8" 저장은 맨 앞에 BOM을 붙임. 그대로 두면 첫 값에 BOM이 붙고 첫 줄의 #도 주석으로 읽히지 않음
func withoutBOM(file io.Reader) io.Reader {
	buffered := bufio.NewReader(file)
	if start, err := buffered.Peek(len(utf8BOM)); err == nil && bytes.Equal(start, utf8BOM) {
		buffered.Discard(len(utf8BOM))
	}
	return buffered
}

func linkRecords(conn *sql.DB, reader *csv.Reader, from string, dryRun bool) (map[auth.LinkStatus]int, error) {
	counts := map[auth.LinkStatus]int{}
	for {
		record, err := reader.Read()
		if errors.Is(err, io.EOF) {
			return counts, nil
		}
		if err != nil {
			return nil, err
		}
		line, _ := reader.FieldPos(0)
		oldID, subject := strings.TrimSpace(record[0]), strings.TrimSpace(record[1])
		if oldID == "" || subject == "" {
			return nil, fmt.Errorf("line %d: both values are required", line)
		}
		result, err := auth.LinkOIDCIdentity(conn, from, oldID, subject, dryRun)
		if err != nil {
			return nil, fmt.Errorf("line %d: %w", line, err)
		}
		counts[result.Status]++
		fmt.Println(describeLink(line, result))
	}
}

// 외부 ID와 sub는 출력하지 않고 NAS 사용자 이름만 보임
func describeLink(line int, result auth.LinkResult) string {
	text := fmt.Sprintf("line %d: %s", line, result.Status)
	if result.Username != "" {
		text += fmt.Sprintf(" (%s)", result.Username)
	}
	if result.Detail != "" {
		text += ": " + result.Detail
	}
	return text
}
