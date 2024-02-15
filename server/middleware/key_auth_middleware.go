package middleware

import (
	"errors"
	"github.com/ArkamFahry/storage/server/config"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/keyauth"
)

func KeyAuth(config *config.Config) fiber.Handler {
	return keyauth.New(keyauth.Config{
		ErrorHandler: func(ctx *fiber.Ctx, err error) error {
			if errors.Is(err, keyauth.ErrMissingOrMalformedAPIKey) {
				return ctx.Status(fiber.StatusUnauthorized).JSON(&HttpError{
					StatusCode: fiber.StatusUnauthorized,
					Message:    "missing or malformed api key access denied. please provide a valid api key",
					Path:       ctx.Path(),
					RequestId:  ctx.Context().Value("request_id").(string),
				})
			}
			return ctx.Status(fiber.StatusUnauthorized).JSON(&HttpError{
				StatusCode: fiber.StatusUnauthorized,
				Message:    "invalid api key access denied. please provide a valid api key",
				Path:       ctx.Path(),
				RequestId:  ctx.Context().Value("request_id").(string),
			})
		},
		KeyLookup: "header:X-STORAGE-API-KEY",
		Validator: func(ctx *fiber.Ctx, apiKey string) (bool, error) {
			if apiKey == config.ServerApiKey {
				return true, nil
			} else {
				return false, ctx.Status(fiber.StatusUnauthorized).JSON(&HttpError{
					StatusCode: fiber.StatusUnauthorized,
					Message:    "invalid api key access denied. please provide a valid api key",
					Path:       ctx.Path(),
					RequestId:  ctx.Context().Value("request_id").(string),
				})
			}
		},
	})
}
