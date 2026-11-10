package share

import (
	"archive/zip"
	"io"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/JMC50/nas/internal/files"
)

// 이미 압축된 형식은 다시 압축해도 줄지 않으므로 저장만 함
var storedExtensions = map[string]bool{
	".zip": true, ".gz": true, ".7z": true, ".rar": true,
	".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".webp": true, ".avif": true,
	".mp4": true, ".mkv": true, ".mov": true, ".webm": true, ".avi": true,
	".mp3": true, ".m4a": true, ".ogg": true, ".flac": true,
	".pdf": true, ".docx": true, ".xlsx": true, ".pptx": true,
}

// streamFolder는 폴더를 임시 파일 없이 zip으로 바로 보냅니다.
// 심볼릭 링크는 실제 경로가 데이터 폴더 안의 보통 파일일 때만 넣고, 폴더를 가리키는 링크와 특수 파일은 뺍니다.
func streamFolder(w http.ResponseWriter, target resolvedTarget, name string) {
	folder, root := target.path, target.root
	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", files.ContentDispositionAttachment(name+".zip"))
	archive := zip.NewWriter(w)

	walkErr := filepath.WalkDir(folder, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			if entry != nil && entry.IsDir() {
				return fs.SkipDir
			}
			return nil
		}
		rel, err := filepath.Rel(folder, path)
		if err != nil {
			return nil
		}
		entryName := name + "/"
		if rel != "." {
			entryName += filepath.ToSlash(rel)
		}
		switch {
		case entry.Type()&fs.ModeSymlink != 0:
			resolved, err := filepath.EvalSymlinks(path)
			if err != nil || !insideRoot(root, resolved) {
				return nil
			}
			info, err := os.Stat(resolved)
			if err != nil || !info.Mode().IsRegular() {
				return nil
			}
			return addFile(archive, entryName, resolved, info)
		case entry.IsDir():
			info, err := entry.Info()
			if err != nil {
				return fs.SkipDir
			}
			_, err = archive.CreateHeader(&zip.FileHeader{Name: strings.TrimSuffix(entryName, "/") + "/", Modified: info.ModTime()})
			return err
		case entry.Type().IsRegular():
			info, err := entry.Info()
			if err != nil {
				return nil
			}
			return addFile(archive, entryName, path, info)
		default:
			return nil
		}
	})
	if walkErr != nil {
		slog.Warn("share folder zip stopped", "folder", folder, "err", walkErr)
	}
	if err := archive.Close(); err != nil {
		slog.Warn("share folder zip close failed", "folder", folder, "err", err)
	}
}

// addFile은 열 수 없는 파일을 건너뛰고, 쓰는 중 오류(받는 쪽 연결 끊김 등)는 돌려줘서 폴더 순회를 멈춥니다.
func addFile(archive *zip.Writer, entryName, path string, info os.FileInfo) error {
	source, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer source.Close()
	method := zip.Deflate
	if storedExtensions[strings.ToLower(filepath.Ext(path))] {
		method = zip.Store
	}
	writer, err := archive.CreateHeader(&zip.FileHeader{Name: entryName, Method: method, Modified: info.ModTime()})
	if err != nil {
		return err
	}
	_, err = io.Copy(writer, source)
	return err
}
