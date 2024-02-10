package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/oklog/ulid/v2"
)

func RequestId() fiber.Handler {
	return requestid.New(requestid.Config{
		Generator:  ulid.Make().String,
		ContextKey: "request_id",
	})
}
