package validators

import "fmt"

func ValidateMaxAllowedObjectSize(maxAllowedObjectSize int64) error {
	if maxAllowedObjectSize < 0 {
		return fmt.Errorf("max allowed object size must be 0 or greater than 0")
	}

	return nil
}
