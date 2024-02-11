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
	routesV1.Post("/buckets/:id/empty", bc.EmptyBucket)
	routesV1.Delete("/buckets/:id/disable", bc.DisableBucket)
	routesV1.Delete("/buckets/:id/enable", bc.EnableBucket)
	routesV1.Patch("/buckets/:id", bc.UpdateBucket)
	routesV1.Delete("/buckets/:id", bc.DeleteBucket)
	routesV1.Get("/buckets/:id", bc.GetBucket)
	routesV1.Get("/buckets/:id/size", bc.GetBucketSize)
	routesV1.Get("/buckets", bc.ListAllBuckets)
}

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

func (bc *BucketController) EmptyBucket(ctx *fiber.Ctx) error {
	id := ctx.Params("id")

	err := bc.bucketService.EmptyBucket(ctx.Context(), id)
	if err != nil {
		return err
	}

	return ctx.SendStatus(fiber.StatusAccepted)
}

func (bc *BucketController) DisableBucket(ctx *fiber.Ctx) error {
	id := ctx.Params("id")

	disabledBucket, err := bc.bucketService.DisableBucket(ctx.Context(), id)
	if err != nil {
		return err
	}

	return ctx.Status(fiber.StatusOK).JSON(disabledBucket)
}

func (bc *BucketController) EnableBucket(ctx *fiber.Ctx) error {
	id := ctx.Params("id")

	enabledBucket, err := bc.bucketService.EnableBucket(ctx.Context(), id)
	if err != nil {
		return err
	}

	return ctx.Status(fiber.StatusOK).JSON(enabledBucket)
}

func (bc *BucketController) DeleteBucket(ctx *fiber.Ctx) error {
	id := ctx.Params("id")

	err := bc.bucketService.DeleteBucket(ctx.Context(), id)
	if err != nil {
		return err
	}

	return ctx.SendStatus(fiber.StatusNoContent)
}

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

func (bc *BucketController) ListAllBuckets(ctx *fiber.Ctx) error {
	buckets, err := bc.bucketService.ListAllBuckets(ctx.Context())
	if err != nil {
		return err
	}

	return ctx.Status(fiber.StatusOK).JSON(buckets)
}
