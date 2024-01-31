package bucket

import (
	"github.com/gofiber/fiber/v2"
)

type BucketController struct {
	bucketService *BucketService
}

func NewBucketController(bucketService *BucketService) *BucketController {
	return &BucketController{
		bucketService: bucketService,
	}
}

func (bc *BucketController) RegisterBucketRoutes(app *fiber.App) {
	routes := app.Group("/api")

	routesV1 := routes.Group("/v1")

	routesV1.Post("/buckets", bc.CreateBucket)
	routesV1.Post("/buckets/:id/empty", bc.EmptyBucket)
	routesV1.Post("/buckets/:id/rename", bc.RenameBucket)
	routesV1.Delete("/buckets/:id/disable", bc.DisableBucket)
	routesV1.Delete("/buckets/:id/enable", bc.EnableBucket)
	routesV1.Post("/buckets/:id/add-allowed-content-types", bc.AddAllowedContentTypesToBucket)
	routesV1.Patch("/buckets/:id/remove-allowed-content-types", bc.RemoveAllowedContentTypesFromBucket)
	routesV1.Patch("/buckets/:id", bc.UpdateBucket)
	routesV1.Delete("/buckets/:id", bc.DeleteBucket)
	routesV1.Get("/buckets/:id", bc.GetBucket)
	routesV1.Get("/buckets/:id/size", bc.GetBucketSize)
	routesV1.Get("/buckets", bc.ListBuckets)
}

func (bc *BucketController) CreateBucket(ctx *fiber.Ctx) error {
	return nil
}

func (bc *BucketController) UpdateBucket(ctx *fiber.Ctx) error {
	return nil
}

func (bc *BucketController) AddAllowedContentTypesToBucket(ctx *fiber.Ctx) error {
	return nil
}

func (bc *BucketController) RemoveAllowedContentTypesFromBucket(ctx *fiber.Ctx) error {
	return nil
}

func (bc *BucketController) EmptyBucket(ctx *fiber.Ctx) error {
	return nil
}

func (bc *BucketController) RenameBucket(ctx *fiber.Ctx) error {
	return nil
}

func (bc *BucketController) DisableBucket(ctx *fiber.Ctx) error {
	return nil
}

func (bc *BucketController) EnableBucket(ctx *fiber.Ctx) error {
	return nil
}

func (bc *BucketController) DeleteBucket(ctx *fiber.Ctx) error {
	return nil
}

func (bc *BucketController) GetBucket(ctx *fiber.Ctx) error {
	return nil
}

func (bc *BucketController) GetBucketSize(ctx *fiber.Ctx) error { return nil }

func (bc *BucketController) ListBuckets(ctx *fiber.Ctx) error {
	return nil
}
