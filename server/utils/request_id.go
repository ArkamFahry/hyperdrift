package utils

import "context"

func RequestId(ctx context.Context) string {
	var requestId string
	if reqId, ok := ctx.Value("request_id").(string); ok {
		requestId = reqId
	}

	return requestId
}
