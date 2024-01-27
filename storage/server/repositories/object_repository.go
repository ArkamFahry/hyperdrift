package repositories

import (
	"context"
	"github.com/ArkamFahry/hyperdrift/storage/server/database"
	"github.com/ArkamFahry/hyperdrift/storage/server/models"
)

type IObjectRepository interface {
	CreateObject(ctx context.Context, createObject *models.CreateObject) error
	RenameObject(ctx context.Context, renameObject *models.RenameObject) error
	CopyObject(ctx context.Context, copyObject *models.CopyObject) error
	MoveObject(ctx context.Context, moveObject *models.MoveObject) error
	GetObjectByBucketIdAndId(ctx context.Context, bucketId string, id string) (*models.Object, error)
	SearchObjectByBucketIdAndName(ctx context.Context, bucketId string, name string) ([]*models.Object, error)
}

type ObjectRepository struct {
	db *database.Queries
}

func NewObjectRepository(db *database.Queries) IObjectRepository {
	return &ObjectRepository{
		db: db,
	}
}

func (or *ObjectRepository) CreateObject(ctx context.Context, createObject *models.CreateObject) error {
	err := or.db.CreateObject(ctx, &database.CreateObjectParams{
		ID:          createObject.Id,
		BucketID:    createObject.BucketId,
		Name:        createObject.Name,
		ContentType: createObject.ContentType,
		Size:        createObject.Size,
		Public:      createObject.Public,
		Metadata:    createObject.Metadata,
	})
	if err != nil {
		return err
	}

	return nil
}
