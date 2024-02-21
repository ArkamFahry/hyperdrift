package middleware

import (
	"encoding/json"
	"errors"
	"github.com/ArkamFahry/storage/server/srverr"
	"github.com/ArkamFahry/storage/server/utils"
	"github.com/gofiber/fiber/v2"
)

type HttpError struct {
	// status_code return a http status code depending on the error
	StatusCode int `json:"status_code" example:"404"`
	/*
		message is a human-readable displayable safe to use error
		message on what the error is reason for the error and how to resolve the error
	*/
	Message string `json:"message" example:"bucket not found"`
	//	path the error happened on means the http endpoint where the error happened
	Path string `json:"path" example:"/api/v1/buckets"`
	/*
			request_id is an id used to track the request from start to response return.
		 	this can be used for finding out the starting point of the error how it happened in the system
	*/
	RequestId string `json:"request_id" example:"01HPG4GN5JY2Z6S0638ERSG375"`
}

func ErrorHandler(ctx *fiber.Ctx, err error) error {
	httpError := &HttpError{
		StatusCode: fiber.StatusInternalServerError,
		Message:    "internal server error",
		Path:       ctx.Path(),
		RequestId:  utils.RequestId(ctx.Context()),
	}

	var srvError srverr.ServiceError
	var fiberError *fiber.Error
	var jsonSyntaxError *json.SyntaxError

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
			httpError.RequestId = utils.RequestId(ctx.Context())
		}

		if errors.As(err, &jsonSyntaxError) {
			httpError.StatusCode = fiber.StatusBadRequest
			httpError.Message = jsonSyntaxError.Error()
			httpError.Path = ctx.Path()
			httpError.RequestId = utils.RequestId(ctx.Context())
		}

		return ctx.Status(httpError.StatusCode).JSON(httpError)
	}

	return nil
}
