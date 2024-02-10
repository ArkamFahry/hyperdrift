package middleware

import (
	"errors"
	"github.com/ArkamFahry/hyperdrift/storage/server/srverr"
	"github.com/gofiber/fiber/v2"
)

type HttpError struct {
	StatusCode int    `json:"status_code"`
	Message    string `json:"message"`
	Path       string `json:"path"`
	RequestId  string `json:"request_id"`
}

func ErrorHandler(ctx *fiber.Ctx, err error) error {
	httpError := &HttpError{
		StatusCode: fiber.StatusInternalServerError,
		Message:    "internal server error",
		Path:       ctx.Path(),
		RequestId:  ctx.Get("X-Request-Id"),
	}

	var srvError *srverr.ServiceError
	var fiberError *fiber.Error

	if err != nil {
		if errors.As(err, &srvError) {
			switch {
			case errors.Is(srvError.ErrorCode, srverr.NotFoundError):
				httpError.StatusCode = fiber.StatusNotFound
			case errors.Is(srvError.ErrorCode, srverr.ConflictError):
				httpError.StatusCode = fiber.StatusConflict
			case errors.Is(srvError.ErrorCode, srverr.InvalidInputError):
				httpError.StatusCode = fiber.StatusUnprocessableEntity
			case errors.Is(srvError.ErrorCode, srverr.BadRequestError):
				httpError.StatusCode = fiber.StatusBadRequest
			case errors.Is(srvError.ErrorCode, srverr.ForbiddenError):
				httpError.StatusCode = fiber.StatusForbidden
			case errors.Is(srvError.ErrorCode, srverr.UnknownError):
				httpError.StatusCode = fiber.StatusInternalServerError
			}
			httpError.Message = srvError.Message
			httpError.Path = ctx.Path()
			httpError.RequestId = srvError.RequestId
		}

		if errors.As(err, &fiberError) {
			httpError.StatusCode = fiberError.Code
			httpError.Message = fiberError.Message
			httpError.Path = ctx.Path()
			httpError.RequestId = ctx.Get("X-Request-Id")
		}

		return ctx.Status(httpError.StatusCode).JSON(httpError)
	}

	return nil
}
