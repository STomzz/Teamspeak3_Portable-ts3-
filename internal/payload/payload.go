package payload

import "fmt"

func validateZip(data []byte) error {
	if len(data) < 4 {
		return fmt.Errorf("payload too small")
	}
	if string(data[:2]) != "PK" {
		return fmt.Errorf("payload is not a zip archive")
	}
	return nil
}
