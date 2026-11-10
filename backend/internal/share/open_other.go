//go:build !unix

package share

import "os"

const readFlags = os.O_RDONLY
