package share

import (
	"archive/zip"
	"bufio"
	"context"
	"errors"
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

type zipItem struct {
	rel  string
	name string
}

// 0은 제한 없음
type zipLimits struct {
	entries int
	bytes   int64
}

var (
	errTooManyEntries = errors.New("too many entries for one zip")
	errTooManyBytes   = errors.New("too many bytes for one zip")
)

type zipUsage struct {
	limits  zipLimits
	entries int
	bytes   int64
}

func (u *zipUsage) addEntry() error {
	u.entries++
	if u.limits.entries > 0 && u.entries > u.limits.entries {
		return errTooManyEntries
	}
	return nil
}

func (u *zipUsage) addBytes(size int64) error {
	u.bytes += size
	if u.limits.bytes > 0 && u.bytes > u.limits.bytes {
		return errTooManyBytes
	}
	return nil
}

type zipStream struct {
	zipUsage
	guard   *streamGuard
	root    *os.Root
	archive *zip.Writer
}

// streamZip은 os.Root 안에서만 항목을 돌며 zip을 바로 씀. 오류를 돌려주면 zip은 끝맺지 않은 상태임
func streamZip(guard *streamGuard, root *os.Root, items []zipItem, filename string, limits zipLimits) error {
	guard.Header().Set("Content-Type", "application/zip")
	guard.Header().Set("Content-Disposition", files.ContentDispositionAttachment(filename))
	z := &zipStream{zipUsage: zipUsage{limits: limits}, guard: guard, root: root, archive: zip.NewWriter(bufio.NewWriterSize(guard, zipBufferBytes))}
	for _, item := range items {
		if err := walkZipEntries(root, item, z.add); err != nil {
			return err
		}
	}
	return z.archive.Close()
}

// measureZip은 보내기 전에 streamZip이 넣을 항목 수와 크기가 한도 안인지 봄
func measureZip(ctx context.Context, root *os.Root, items []zipItem, limits zipLimits) (zipUsage, error) {
	usage := zipUsage{limits: limits}
	measure := func(entryPath, _ string, entry fs.DirEntry) error {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := usage.addEntry(); err != nil || entry.IsDir() {
			return err
		}
		info, err := root.Stat(entryPath)
		if err != nil || !info.Mode().IsRegular() {
			return nil
		}
		return usage.addBytes(info.Size())
	}
	for _, item := range items {
		if err := walkZipEntries(root, item, measure); err != nil {
			return usage, err
		}
	}
	return usage, nil
}

// walkZipEntries는 zip에 넣을 수 있는 폴더, 보통 파일, 심볼릭 링크만 zip 안의 이름과 함께 넘김. 읽지 못한 항목은 건너뜀
func walkZipEntries(root *os.Root, item zipItem, visit func(entryPath, entryName string, entry fs.DirEntry) error) error {
	return fs.WalkDir(root.FS(), item.rel, func(entryPath string, entry fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		inner := strings.TrimPrefix(strings.TrimPrefix(entryPath, item.rel), "/")
		// Windows 압축 해제 프로그램은 \를 폴더 구분자로 읽어 폴더 밖에 쓸 수 있음
		if strings.Contains(inner, `\`) {
			if entry.IsDir() {
				return fs.SkipDir
			}
			return nil
		}
		if !entry.IsDir() && !entry.Type().IsRegular() && entry.Type()&fs.ModeSymlink == 0 {
			return nil
		}
		return visit(entryPath, path.Join(item.name, inner), entry)
	})
}

func (z *zipStream) add(entryPath, entryName string, entry fs.DirEntry) error {
	if err := z.guard.checkpoint(); err != nil {
		return err
	}
	if err := z.addEntry(); err != nil {
		return err
	}
	if entry.IsDir() {
		return z.addDir(entry, entryName)
	}
	return z.addFile(entryPath, entryName)
}

func (z *zipStream) addDir(entry fs.DirEntry, entryName string) error {
	info, err := entry.Info()
	if err != nil {
		return fs.SkipDir
	}
	_, err = z.archive.CreateHeader(&zip.FileHeader{Name: entryName + "/", Modified: info.ModTime()})
	return err
}

// 열 수 없거나 보통 파일이 아닌 항목은 건너뛰고, 쓰기 오류(받는 쪽 연결 끊김, 링크 끊김)와 한도 초과는 돌려줘서 순회를 멈춤
func (z *zipStream) addFile(entryPath, entryName string) error {
	file, info, err := openRegular(z.root, entryPath)
	if err != nil {
		return nil
	}
	defer file.Close()
	if err := z.addBytes(info.Size()); err != nil {
		return err
	}
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
