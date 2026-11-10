package share

import (
	"archive/zip"
	"bufio"
	"io"
	"io/fs"
	"os"
	"path"
	"strings"

	"github.com/JMC50/nas/internal/files"
)

const zipBufferBytes = 64 << 10

// 이미 압축된 형식은 다시 압축해도 줄지 않으므로 저장만 함
var storedExtensions = map[string]bool{
	".zip": true, ".gz": true, ".7z": true, ".rar": true,
	".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".webp": true, ".avif": true,
	".mp4": true, ".mkv": true, ".mov": true, ".webm": true, ".avi": true,
	".mp3": true, ".m4a": true, ".ogg": true, ".flac": true,
	".pdf": true, ".docx": true, ".xlsx": true, ".pptx": true,
}

type folderZip struct {
	guard   *streamGuard
	root    *os.Root
	rel     string
	name    string
	archive *zip.Writer
}

// streamFolder는 os.Root 안에서만 폴더를 돌며 zip을 바로 씀. 오류를 돌려주면 zip은 끝맺지 않은 상태임
func streamFolder(guard *streamGuard, root *os.Root, rel, name string) error {
	guard.Header().Set("Content-Type", "application/zip")
	guard.Header().Set("Content-Disposition", files.ContentDispositionAttachment(name+".zip"))
	z := &folderZip{guard: guard, root: root, rel: rel, name: name, archive: zip.NewWriter(bufio.NewWriterSize(guard, zipBufferBytes))}
	if err := fs.WalkDir(root.FS(), rel, z.visit); err != nil {
		return err
	}
	return z.archive.Close()
}

func (z *folderZip) visit(entryPath string, entry fs.DirEntry, err error) error {
	if err != nil {
		return nil
	}
	inner := strings.TrimPrefix(strings.TrimPrefix(entryPath, z.rel), "/")
	// Windows 압축 해제 프로그램은 \를 폴더 구분자로 읽어 폴더 밖에 쓸 수 있음
	if strings.Contains(inner, `\`) {
		if entry.IsDir() {
			return fs.SkipDir
		}
		return nil
	}
	if err := z.guard.checkpoint(); err != nil {
		return err
	}
	entryName := path.Join(z.name, inner)
	switch {
	case entry.IsDir():
		return z.addDir(entry, entryName)
	case entry.Type().IsRegular(), entry.Type()&fs.ModeSymlink != 0:
		return z.addFile(entryPath, entryName)
	default:
		return nil
	}
}

func (z *folderZip) addDir(entry fs.DirEntry, entryName string) error {
	info, err := entry.Info()
	if err != nil {
		return fs.SkipDir
	}
	_, err = z.archive.CreateHeader(&zip.FileHeader{Name: entryName + "/", Modified: info.ModTime()})
	return err
}

// 열 수 없거나 보통 파일이 아닌 항목은 건너뛰고, 쓰기 오류(받는 쪽 연결 끊김, 링크 끊김)는 돌려줘서 순회를 멈춤
func (z *folderZip) addFile(entryPath, entryName string) error {
	file, info, err := openRegular(z.root, entryPath)
	if err != nil {
		return nil
	}
	defer file.Close()
	method := zip.Deflate
	if storedExtensions[strings.ToLower(path.Ext(entryPath))] {
		method = zip.Store
	}
	writer, err := z.archive.CreateHeader(&zip.FileHeader{Name: entryName, Method: method, Modified: info.ModTime()})
	if err != nil {
		return err
	}
	_, err = io.Copy(writer, file)
	return err
}
