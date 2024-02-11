package controllers

import (
	"github.com/ArkamFahry/hyperdrift/storage/server/dto"
	"github.com/ArkamFahry/hyperdrift/storage/server/services"
	"github.com/gofiber/fiber/v2"
)

type ObjectController struct {
	objectService *services.ObjectService
}

func NewObjectController(objectService *services.ObjectService) *ObjectController {
	return &ObjectController{
		objectService: objectService,
	}
}

func (oc *ObjectController) RegisterObjectRoutes(app *fiber.App) {
	routes := app.Group("/api")

	routesV1 := routes.Group("/v1")

	routesV1.Post("/pre-signed-upload-object", oc.CreatePreSignedUploadObject)
	routesV1.Post("/pre-signed-upload-object/:id/complete", oc.CompletePreSignedObjectUpload)
}

func (oc *ObjectController) CreatePreSignedUploadObject(ctx *fiber.Ctx) error {
	var preSignedUploadObjectCreate dto.PreSignedUploadObjectCreate

	err := ctx.BodyParser(&preSignedUploadObjectCreate)
	if err != nil {
		return err
	}

	preSignedObject, err := oc.objectService.CreatePreSignedUploadObject(ctx.Context(), &preSignedUploadObjectCreate)
	if err != nil {
		return err
	}

	return ctx.Status(fiber.StatusCreated).JSON(preSignedObject)
}

func (oc *ObjectController) CompletePreSignedObjectUpload(ctx *fiber.Ctx) error {
	id := ctx.Params("id")

	err := oc.objectService.CompletePreSignedObjectUpload(ctx.Context(), id)
	if err != nil {
		return err
	}

	return ctx.SendStatus(fiber.StatusAccepted)
}
