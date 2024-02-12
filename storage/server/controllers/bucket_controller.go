package controllers

import (
	"github.com/ArkamFahry/hyperdrift/storage/server/dto"
	"github.com/ArkamFahry/hyperdrift/storage/server/services"
	"github.com/gofiber/fiber/v2"
)

type BucketController struct {
	bucketService *services.BucketService
}

func NewBucketController(bucketService *services.BucketService) *BucketController {
	return &BucketController{
		bucketService: bucketService,
	}
}

func (bc *BucketController) RegisterBucketRoutes(app *fiber.App) {
	routes := app.Group("/api")

	routesV1 := routes.Group("/v1")

	routesV1.Post("/buckets", bc.CreateBucket)
	routesV1.Patch("/buckets/:id", bc.UpdateBucket)
	routesV1.Post("/buckets/:id/empty", bc.EmptyBucket)
	routesV1.Post("/buckets/:id/disable", bc.DisableBucket)
	routesV1.Post("/buckets/:id/enable", bc.EnableBucket)
	routesV1.Delete("/buckets/:id", bc.DeleteBucket)
	routesV1.Get("/buckets/:id", bc.GetBucket)
	routesV1.Get("/buckets/:id/size", bc.GetBucketSize)
	routesV1.Get("/buckets", bc.ListAllBuckets)
}

// CreateBucket is used to create a bucket
// @Summary Create a bucket
// @Description Create a bucket
// @Tags buckets
// @Accept json
// @Produce json
// @Param bucket body dto.BucketCreate true "Bucket Create"
// @Success 201 {object} entities.Bucket
// @Failure 400 {object} middleware.HttpError
// @Failure 500 {object} middleware.HttpError
// @Router /api/v1/buckets [post]
func (bc *BucketController) CreateBucket(ctx *fiber.Ctx) error {
	var bucketCreate dto.BucketCreate

	err := ctx.BodyParser(&bucketCreate)
	if err != nil {
		return err
	}

	createdBucket, err := bc.bucketService.CreateBucket(ctx.Context(), &bucketCreate)
	if err != nil {
		return err
	}

	return ctx.Status(fiber.StatusCreated).JSON(createdBucket)
}

// UpdateBucket is used to update a bucket
// @Summary Update a bucket by id
// @Description Update a bucket by id
// @Tags buckets
// @Accept json
// @Produce json
// @Param id path string true "Bucket ID"
// @Param bucket body dto.BucketUpdate true "Bucket Update"
// @Success 200 {object} entities.Bucket
// @Failure 400 {object} middleware.HttpError
// @Failure 500 {object} middleware.HttpError
// @Router /api/v1/buckets/{id} [patch]
func (bc *BucketController) UpdateBucket(ctx *fiber.Ctx) error {
	var bucketUpdate dto.BucketUpdate

	id := ctx.Params("id")

	err := ctx.BodyParser(&bucketUpdate)
	if err != nil {
		return err
	}

	updatedBucket, err := bc.bucketService.UpdateBucket(ctx.Context(), id, &bucketUpdate)
	if err != nil {
		return err
	}

	return ctx.Status(fiber.StatusOK).JSON(updatedBucket)
}

// EmptyBucket is used to empty a bucket
// @Summary Empty a bucket by id
// @Description Empty a bucket by id
// @Tags buckets
// @Accept json
// @Produce json
// @Param id path string true "Bucket ID"
// @Success 202
// @Failure 400 {object} middleware.HttpError
// @Failure 500 {object} middleware.HttpError
// @Router /api/v1/buckets/{id}/empty [post]
func (bc *BucketController) EmptyBucket(ctx *fiber.Ctx) error {
	id := ctx.Params("id")

	err := bc.bucketService.EmptyBucket(ctx.Context(), id)
	if err != nil {
		return err
	}

	return ctx.SendStatus(fiber.StatusAccepted)
}

// DisableBucket is used to disable an enabled bucket
// @Summary Disable an enabled bucket by id
// @Description Disable an enabled bucket by id
// @Tags buckets
// @Accept json
// @Produce json
// @Param id path string true "Bucket ID"
// @Success 200 {object} entities.Bucket
// @Failure 400 {object} middleware.HttpError
// @Failure 500 {object} middleware.HttpError
// @Router /api/v1/buckets/{id}/disable [post]
func (bc *BucketController) DisableBucket(ctx *fiber.Ctx) error {
	id := ctx.Params("id")

	disabledBucket, err := bc.bucketService.DisableBucket(ctx.Context(), id)
	if err != nil {
		return err
	}

	return ctx.Status(fiber.StatusOK).JSON(disabledBucket)
}

// EnableBucket is used to enable a disabled bucket
// @Summary Enable a disabled bucket by id
// @Description Enable a disabled bucket by id
// @Tags buckets
// @Accept json
// @Produce json
// @Param id path string true "Bucket ID"
// @Success 200 {object} entities.Bucket
// @Failure 400 {object} middleware.HttpError
// @Failure 500 {object} middleware.HttpError
// @Router /api/v1/buckets/{id}/enable [post]
func (bc *BucketController) EnableBucket(ctx *fiber.Ctx) error {
	id := ctx.Params("id")

	enabledBucket, err := bc.bucketService.EnableBucket(ctx.Context(), id)
	if err != nil {
		return err
	}

	return ctx.Status(fiber.StatusOK).JSON(enabledBucket)
}

// DeleteBucket is used to delete a bucket
// @Summary Delete a bucket by id
// @Description Delete a bucket by id
// @Tags buckets
// @Accept json
// @Produce json
// @Param id path string true "Bucket ID"
// @Success 204
// @Failure 400 {object} middleware.HttpError
// @Failure 500 {object} middleware.HttpError
// @Router /api/v1/buckets/{id} [delete]
func (bc *BucketController) DeleteBucket(ctx *fiber.Ctx) error {
	id := ctx.Params("id")

	err := bc.bucketService.DeleteBucket(ctx.Context(), id)
	if err != nil {
		return err
	}

	return ctx.SendStatus(fiber.StatusNoContent)
}

// GetBucket is used to get a bucket
// @Summary Get a bucket by id
// @Description Get a bucket by id
// @Tags buckets
// @Accept json
// @Produce json
// @Param id path string true "Bucket ID"
// @Success 200 {object} entities.Bucket
// @Failure 400 {object} middleware.HttpError
// @Failure 500 {object} middleware.HttpError
// @Router /api/v1/buckets/{id} [get]
func (bc *BucketController) GetBucket(ctx *fiber.Ctx) error {
	id := ctx.Params("id")

	bucket, err := bc.bucketService.GetBucketById(ctx.Context(), id)
	if err != nil {
		return err
	}

	return ctx.Status(fiber.StatusOK).JSON(bucket)
}

func (bc *BucketController) GetBucketSize(ctx *fiber.Ctx) error {
	id := ctx.Params("id")

	bucketSize, err := bc.bucketService.GetBucketSize(ctx.Context(), id)
	if err != nil {
		return err
	}

	return ctx.Status(fiber.StatusOK).JSON(bucketSize)
}

// ListAllBuckets is used to list all available buckets
// @Summary List all available buckets
// @Description List all available buckets
// @Tags buckets
// @Accept json
// @Produce json
// @Success 200 {array} entities.Bucket
// @Failure 400 {object} middleware.HttpError
// @Failure 500 {object} middleware.HttpError
// @Router /api/v1/buckets [get]
func (bc *BucketController) ListAllBuckets(ctx *fiber.Ctx) error {
	buckets, err := bc.bucketService.ListAllBuckets(ctx.Context())
	if err != nil {
		return err
	}

	return ctx.Status(fiber.StatusOK).JSON(buckets)
}
