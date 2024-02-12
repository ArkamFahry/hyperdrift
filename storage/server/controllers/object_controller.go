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

	routesV1.Post("/objects/pre-signed/upload", oc.CreatePreSignedUploadObject)
	routesV1.Post("/objects/pre-signed/upload/:id/complete", oc.CompletePreSignedObjectUpload)
	routesV1.Get("/objects/pre-signed/download/:id", oc.CreatePreSignedDownloadObject)
	routesV1.Delete("/objects/:id", oc.DeleteObject)
	routesV1.Get("/objects/:id", oc.GetObjectById)
	routesV1.Get("/objects/:bucket_name/:object_path", oc.SearchObjectsByBucketNameAndObjectPath)
}

func (oc *ObjectController) CreatePreSignedUploadObject(ctx *fiber.Ctx) error {
	var preSignedUploadObjectCreate dto.PreSignedUploadObjectCreate

	err := ctx.BodyParser(&preSignedUploadObjectCreate)
	if err != nil {
		return err
	}

	preSignedUploadObject, err := oc.objectService.CreatePreSignedUploadObject(ctx.Context(), &preSignedUploadObjectCreate)
	if err != nil {
		return err
	}

	return ctx.Status(fiber.StatusCreated).JSON(preSignedUploadObject)
}

func (oc *ObjectController) CompletePreSignedObjectUpload(ctx *fiber.Ctx) error {
	id := ctx.Params("id")

	err := oc.objectService.CompletePreSignedObjectUpload(ctx.Context(), id)
	if err != nil {
		return err
	}

	return ctx.SendStatus(fiber.StatusAccepted)
}

func (oc *ObjectController) CreatePreSignedDownloadObject(ctx *fiber.Ctx) error {
	id := ctx.Params("id")

	preSignedDownloadObject, err := oc.objectService.CreatePreSignedDownloadObject(ctx.Context(), id)
	if err != nil {
		return err
	}

	return ctx.Status(fiber.StatusOK).JSON(preSignedDownloadObject)
}

func (oc *ObjectController) DeleteObject(ctx *fiber.Ctx) error {
	id := ctx.Params("id")

	err := oc.objectService.DeleteObject(ctx.Context(), id)
	if err != nil {
		return err
	}

	return ctx.SendStatus(fiber.StatusNoContent)
}

func (oc *ObjectController) GetObjectById(ctx *fiber.Ctx) error {
	id := ctx.Params("id")

	object, err := oc.objectService.GetObjectById(ctx.Context(), id)
	if err != nil {
		return err
	}

	return ctx.Status(fiber.StatusOK).JSON(object)
}

func (oc *ObjectController) SearchObjectsByBucketNameAndObjectPath(ctx *fiber.Ctx) error {
	bucketName := ctx.Params("bucket_name")
	objectPath := ctx.Params("object_path")

	levels := ctx.QueryInt("levels")
	limit := ctx.QueryInt("limit")
	offset := ctx.QueryInt("offset")

	objects, err := oc.objectService.SearchObjectsByBucketNameAndObjectPath(ctx.Context(), bucketName, objectPath, int32(levels), int32(limit), int32(offset))
	if err != nil {
		return err
	}

	return ctx.Status(fiber.StatusOK).JSON(objects)
}
