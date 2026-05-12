package integration

import "os"

// _osMkdirAll wraps os.MkdirAll so tests in this package can stay tidy.
func _osMkdirAll(path string) error {
	return os.MkdirAll(path, 0o755)
}
