package file

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/airenas/go-app/pkg/goapp"
)

// Loader loads file by key from path
type Loader struct {
	baseDir string
}

// NewLoader inits new instance
func NewLoader(path string) (*Loader, error) {
	fi, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("can't check dir %s: %v", path, err)
	}
	if !fi.IsDir() {
		return nil, fmt.Errorf("not dir %s", path)
	}
	return &Loader{baseDir: path}, nil
}

// TakeWav loads file from path using the provided name
func (l *Loader) TakeWav(name string) ([]byte, error) {
	fn := getFileName(l.baseDir, name)
	goapp.Log.Infof("Loading suffix %s", fn)
	return os.ReadFile(fn)
}

func getFileName(b, name string) string {
	return filepath.Join(b, filepath.Base(name))
}

// Info returns info about loader
func (l *Loader) Info() string {
	return fmt.Sprintf("audioLoader(%s)", l.baseDir)
}
