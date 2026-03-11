//go:build bundle

package payload

import (
	_ "embed"
)

//go:embed assets/ts3-client-win64.zip
var bundledPayload []byte

func Load(_ string) ([]byte, error) {
	if err := validateZip(bundledPayload); err != nil {
		return nil, err
	}
	return bundledPayload, nil
}
