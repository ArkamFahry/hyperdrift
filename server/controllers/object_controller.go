package controllers

import (
	"github.com/ArkamFahry/storage/server/models"
	"github.com/ArkamFahry/storage/server/services"
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

	routesV1.Post("/objects/pre-signed/upload/:bucket_id", oc.CreatePreSignedUploadSession)
	routesV1.Post("/objects/pre-signed/upload/:bucket_id/:object_id/complete", oc.CompletePreSignedUploadSession)
	routesV1.Get("/objects/pre-signed/download/:bucket_id/:object_id", oc.CreatePreSignedDownloadSession)
	routesV1.Delete("/objects/:bucket_id/:object_id", oc.DeleteObject)
	routesV1.Get("/objects/search/:bucket_id", oc.SearchObjects)
	routesV1.Get("/objects/:bucket_id/:object_id", oc.GetObject)
}

// CreatePreSignedUploadSession is used to create a pre signed upload session
// @Summary Create a pre signed upload session
// @Description Create a pre signed upload session
// @Tags objects
// @Accept json
// @Produce json
// @Param bucket_id path string true "Bucket ID"
// @Param bucket body models.PreSignedUploadSessionCreate true "Pre Signed Upload Session Create"
// @Success 201 {object} models.PreSignedUploadSession
// @Failure 400 {object} middleware.HttpError
// @Failure 500 {object} middleware.HttpError
// @Router /api/v1/objects/pre-signed/upload/{bucket_id} [post]
func (oc *ObjectController) CreatePreSignedUploadSession(ctx *fiber.Ctx) error {
	var preSignedUploadObjectCreate models.PreSignedUploadSessionCreate

	preSignedUploadObjectCreate.BucketId = ctx.Params("bucket_id")

	err := ctx.BodyParser(&preSignedUploadObjectCreate)
	if err != nil {
		return err
	}

	preSignedUploadObject, err := oc.objectService.CreatePreSignedUploadSession(ctx.Context(), &preSignedUploadObjectCreate)
	if err != nil {
		return err
	}

	return ctx.Status(fiber.StatusCreated).JSON(preSignedUploadObject)
}

// CompletePreSignedUploadSession is used to complete a pre signed upload session
// @Summary Complete a pre signed upload session
// @Description Complete a pre signed upload session
// @Tags objects
// @Accept json
// @Produce json
// @Param bucket_id path string true "Bucket ID"
// @Param object_id path string true "Object ID"
// @Success 200
// @Failure 400 {object} middleware.HttpError
// @Failure 500 {object} middleware.HttpError
// @Router /api/v1/objects/pre-signed/upload/{bucket_id}/{object_id}/complete [post]
func (oc *ObjectController) CompletePreSignedUploadSession(ctx *fiber.Ctx) error {
	bucketId := ctx.Params("bucket_id")
	objectId := ctx.Params("object_id")

	err := oc.objectService.CompletePreSignedUploadSession(ctx.Context(), bucketId, objectId)
	if err != nil {
		return err
	}

	return ctx.SendStatus(fiber.StatusOK)
}

// CreatePreSignedDownloadSession is used to create a pre signed download session
// @Summary Create a pre signed download session
// @Description Create a pre signed download session
// @Tags objects
// @Accept json
// @Produce json
// @Param bucket_id path string true "Bucket ID"
// @Param object_id path string true "Object ID"
// @Param expires_in query int true "Expires In"
// @Success 200 {object} models.PreSignedDownloadSession
// @Failure 400 {object} middleware.HttpError
// @Failure 500 {object} middleware.HttpError
// @Router /api/v1/objects/pre-signed/download/{bucket_id}/{object_id} [post]
func (oc *ObjectController) CreatePreSignedDownloadSession(ctx *fiber.Ctx) error {
	bucketId := ctx.Params("bucket_id")
	objectId := ctx.Params("object_id")

	expiresIn := ctx.QueryInt("expires_in")

	preSignedDownloadObject, err := oc.objectService.CreatePreSignedDownloadSession(ctx.Context(), bucketId, objectId, int64(expiresIn))
	if err != nil {
		return err
	}

	return ctx.Status(fiber.StatusOK).JSON(preSignedDownloadObject)
}

// DeleteObject is used to delete an object
// @Summary Delete an object
// @Description Delete an object
// @Tags objects
// @Accept json
// @Produce json
// @Param bucket_id path string true "Bucket ID"
// @Param object_id path string true "Object ID"
// @Success 204
// @Failure 400 {object} middleware.HttpError
// @Failure 500 {object} middleware.HttpError
// @Router /api/v1/objects/{bucket_id}/{object_id} [delete]
func (oc *ObjectController) DeleteObject(ctx *fiber.Ctx) error {
	bucketId := ctx.Params("bucket_id")
	objectId := ctx.Params("object_id")

	err := oc.objectService.DeleteObject(ctx.Context(), bucketId, objectId)
	if err != nil {
		return err
	}

	return ctx.SendStatus(fiber.StatusNoContent)
}

// SearchObjects is used to search objects
// @Summary Search objects by path
// @Description Search objects by path
// @Tags objects
// @Accept json
// @Produce json
// @Param bucket_id path string true "Bucket ID"
// @Param object_path query string true "Object Path"
// @Param limit query int true "Limit"
// @Param offset query int true "Offset"
// @Success 200 {array} models.Object
// @Failure 400 {object} middleware.HttpError
// @Failure 500 {object} middleware.HttpError
// @Router /api/v1/objects/search/{bucket_id} [get]
func (oc *ObjectController) SearchObjects(ctx *fiber.Ctx) error {
	bucketId := ctx.Params("bucket_id")

	objectPath := ctx.Query("object_path")
	limit := ctx.QueryInt("limit")
	offset := ctx.QueryInt("offset")

	objects, err := oc.objectService.SearchObjects(ctx.Context(), bucketId, objectPath, int32(limit), int32(offset))
	if err != nil {
		return err
	}

	return ctx.Status(fiber.StatusOK).JSON(objects)
}

// GetObject is used to get an object
// @Summary Get an object
// @Description Get an object
// @Tags objects
// @Accept json
// @Produce json
// @Param bucket_id path string true "Bucket ID"
// @Param object_id path string true "Object ID"
// @Success 200 {object} models.Object
// @Failure 400 {object} middleware.HttpError
// @Failure 500 {object} middleware.HttpError
// @Router /api/v1/objects/{bucket_id}/{object_id} [get]
func (oc *ObjectController) GetObject(ctx *fiber.Ctx) error {
	bucketId := ctx.Params("bucket_id")
	objectId := ctx.Params("object_id")

	object, err := oc.objectService.GetObject(ctx.Context(), bucketId, objectId)
	if err != nil {
		return err
	}

	return ctx.Status(fiber.StatusOK).JSON(object)
}
