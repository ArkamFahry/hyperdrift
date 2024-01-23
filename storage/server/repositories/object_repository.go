package repositories

import (
	"context"
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
