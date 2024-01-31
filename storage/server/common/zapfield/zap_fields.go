package zapfield

import "go.uber.org/zap"

func Operation(operation string) zap.Field {
	return zap.String("operation", operation)
}
