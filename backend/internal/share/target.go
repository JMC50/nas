package share

import (
	"errors"
	"os"
	"path"
	"strings"
)

var (
	errBadName      = errors.New("name must not be empty, '.', '..', or contain a slash or backslash")
	errNotShareable = errors.New("not a regular file or folder")
)

// 만들 때와 받을 때 모두 이 정리된 상대 경로로 대상을 찾음
func targetPath(loc, name string) (string, error) {
	if name == "" || name == "." || name == ".." || strings.ContainsAny(name, `/\`) {
		return "", errBadName
	}
	return strings.TrimPrefix(path.Clean("/"+loc+"/"+name), "/"), nil
}

func splitTarget(rel string) (loc, name string) {
	dir := path.Dir(rel)
	if dir == "." {
		return "/", rel
	}
	return "/" + dir, path.Base(rel)
}

// os.Root는 데이터 폴더 안에서만 심볼릭 링크를 따라가므로 밖을 가리키는 링크와 절대 경로 링크는 오류가 됨
func statTarget(root *os.Root, rel string) (os.FileInfo, error) {
	info, err := root.Stat(rel)
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() && !info.IsDir() {
		return nil, errNotShareable
	}
	return info, nil
}

// 열기 전과 연 뒤에 모두 보통 파일인지 봄. 그 사이 FIFO나 장치로 바뀌어도 멈추지 않고 거절됨
func openRegular(root *os.Root, rel string) (*os.File, os.FileInfo, error) {
	before, err := root.Stat(rel)
	if err != nil {
		return nil, nil, err
	}
	if !before.Mode().IsRegular() {
		return nil, nil, errNotShareable
	}
	file, err := root.OpenFile(rel, readFlags, 0)
	if err != nil {
		return nil, nil, err
	}
	info, err := file.Stat()
	if err != nil || !info.Mode().IsRegular() {
		file.Close()
		return nil, nil, errNotShareable
	}
	return file, info, nil
}
