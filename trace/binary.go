package trace

import (
	"os"
	"path/filepath"
)

// CurrentBinary gives the name of the executable running the current Go code
func CurrentBinary() string {
	return filepath.Base(os.Args[0])
}
