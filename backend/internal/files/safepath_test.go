package files

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSafeJoin_AllowsValidPath(t *testing.T) {
	base := t.TempDir()
	result, err := SafeJoin(base, "subdir", "file.txt")
	require.NoError(t, err)
	expected, _ := filepath.Abs(filepath.Join(base, "subdir", "file.txt"))
	require.Equal(t, expected, result)
}

func TestSafeJoin_BlocksDotDotTraversal(t *testing.T) {
	base := t.TempDir()
	_, err := SafeJoin(base, "..", "etc", "passwd")
	require.ErrorIs(t, err, ErrUnsafePath)
}

func TestSafeJoin_BlocksDeepTraversal(t *testing.T) {
	base := t.TempDir()
	_, err := SafeJoin(base, "ok", "..", "..", "..", "etc", "passwd")
	require.ErrorIs(t, err, ErrUnsafePath)
}

func TestSafeJoin_BlocksAbsolutePath(t *testing.T) {
	base := t.TempDir()
	var attempt string
	if runtime.GOOS == "windows" {
		attempt = `C:\Windows\System32\config\SAM`
	} else {
		attempt = "/etc/passwd"
	}
	_, err := SafeJoin(base, attempt)
	require.ErrorIs(t, err, ErrUnsafePath)
}

func TestSafeJoin_BlocksNullByte(t *testing.T) {
	base := t.TempDir()
	_, err := SafeJoin(base, "ok\x00../etc/passwd")
	require.ErrorIs(t, err, ErrUnsafePath)
}

func TestSafeJoin_AllowsBaseItself(t *testing.T) {
	base := t.TempDir()
	result, err := SafeJoin(base)
	require.NoError(t, err)
	expected, _ := filepath.Abs(base)
	require.Equal(t, expected, result)
}

func TestSafeJoin_NormalizesRedundantSeparators(t *testing.T) {
	base := t.TempDir()
	result, err := SafeJoin(base, "/", "/sub//", "file.txt")
	require.NoError(t, err)
	expected, _ := filepath.Abs(filepath.Join(base, "sub", "file.txt"))
	require.Equal(t, expected, result)
}

// TestSafeJoin_AdversarialMatrix covers a broad sweep of attack patterns
// in one table-driven test so future additions stay easy.
func TestSafeJoin_AdversarialMatrix(t *testing.T) {
	base := t.TempDir()
	parent := filepath.Dir(base)
	_ = os.WriteFile(filepath.Join(parent, "leak.txt"), []byte("secret"), 0o644)
	defer os.Remove(filepath.Join(parent, "leak.txt"))

	attacks := []struct {
		name     string
		segments []string
	}{
		{"dot-dot prefix", []string{"..", "leak.txt"}},
		{"dot-dot middle", []string{"sub", "..", "..", "leak.txt"}},
		{"hidden dot-dot in single segment", []string{"sub/../../leak.txt"}},
		{"backslash dot-dot on Windows", []string{`sub\..\..\leak.txt`}},
		{"null byte", []string{"sub\x00leak.txt"}},
	}
	for _, attack := range attacks {
		t.Run(attack.name, func(t *testing.T) {
			result, err := SafeJoin(base, attack.segments...)
			if err == nil {
				absBase, _ := filepath.Abs(base)
				require.True(t,
					result == absBase || filepath.HasPrefix(result, absBase+string(filepath.Separator)),
					"escaped base: %q vs %q", result, absBase)
			}
		})
	}
}

func TestTrimLeadingSlash(t *testing.T) {
	require.Equal(t, "foo", TrimLeadingSlash("/foo"))
	require.Equal(t, "foo", TrimLeadingSlash("///foo"))
	require.Equal(t, "foo", TrimLeadingSlash(`\\foo`))
	require.Equal(t, "foo/bar", TrimLeadingSlash("/foo/bar"))
}
