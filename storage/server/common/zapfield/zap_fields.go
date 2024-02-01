package zapfield

import "go.uber.org/zap"

func Operation(operation string) zap.Field {
	return zap.String("operation", operation)
}

func RequestId(requestId string) zap.Field {
	return zap.String("request_id", requestId)
}
