package utils

import "context"

func RequestId(ctx context.Context) string {
	requestId, ok := ctx.Value("request_id").(string)
	if ok {
		return requestId
	} else {
		return ""
	}
}
