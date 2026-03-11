//go:build !bundle

package payload

import (
	"fmt"
	"os"
	"path/filepath"
)

const payloadRelativePath = "payload/ts3-client-win64.zip"

func Load(appDir string) ([]byte, error) {
	path := filepath.Join(appDir, payloadRelativePath)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read external payload %s: %w", path, err)
	}
	if err := validateZip(data); err != nil {
		return nil, err
	}
	return data, nil
}
