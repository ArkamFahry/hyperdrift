-- name: CreateBucket :exec
insert into storage.buckets
(id, name, allowed_content_types, max_allowed_object_size, public, disabled)
values (sqlc.arg('id'),
        sqlc.arg('name'),
        sqlc.narg('allowed_content_types'),
        sqlc.narg('max_allowed_object_size'),
        sqlc.arg('public'),
        sqlc.arg('disabled'))
returning *;

-- name: UpdateBucket :exec
update storage.buckets
set max_allowed_object_size = coalesce(sqlc.narg('max_allowed_object_size'), max_allowed_object_size),
    public                  = coalesce(sqlc.narg('public'), public)
where id = sqlc.arg('id') and version = sqlc.arg('version');

-- name: AddAllowedContentTypesToBucket :exec
update storage.buckets
set allowed_content_types = array_append(allowed_content_types, sqlc.arg('allowed_content_types')::text[])
where id = sqlc.arg('id') and version = sqlc.arg('version');

-- name: RemoveAllowedContentTypesFromBucket :exec
update storage.buckets
set allowed_content_types = array_remove(allowed_content_types, sqlc.arg('allowed_content_types')::text[])
where id = sqlc.arg('id') and version = sqlc.arg('version');

-- name: DisableBucket :exec
update storage.buckets
set disabled = true
where id = sqlc.arg('id') and version = sqlc.arg('version');

-- name: EnableBucket :exec
update storage.buckets
set disabled = false
where id = sqlc.arg('id') and version = sqlc.arg('version');

-- name: MakeBucketPublic :exec
update storage.buckets
set public = true
where id = sqlc.arg('id') and version = sqlc.arg('version');

-- name: MakeBucketPrivate :exec
update storage.buckets
set public = false
where id = sqlc.arg('id') and version = sqlc.arg('version');

-- name: LockBucket :exec
update storage.buckets
set locked      = true,
    lock_reason = sqlc.arg('lock_reason')::text,
    locked_at   = now()
where id = sqlc.arg('id') and version = sqlc.arg('version');

-- name: UnlockBucket :exec
update storage.buckets
set locked      = false,
    lock_reason = null,
    locked_at   = null
where id = sqlc.arg('id') and version = sqlc.arg('version');

-- name: DeleteBucket :exec
delete
from storage.buckets
where id = sqlc.arg('id');

-- name: GetBucketById :one
select id,
       version,
       name,
       allowed_content_types,
       max_allowed_object_size,
       public,
       disabled,
       locked,
       lock_reason,
       locked_at,
       created_at,
       updated_at
from storage.buckets
where id = sqlc.arg('id')
limit 1;

-- name: GetBucketByName :one
select id,
       version,
       name,
       allowed_content_types,
       max_allowed_object_size,
       public,
       disabled,
       locked,
       lock_reason,
       locked_at,
       created_at,
       updated_at
from storage.buckets
where name = sqlc.arg('name')
limit 1;

-- name: ListAllBuckets :many
select id,
       version,
       name,
       allowed_content_types,
       max_allowed_object_size,
       public,
       disabled,
       locked,
       lock_reason,
       locked_at,
       created_at,
       updated_at
from storage.buckets;

-- name: ListBucketsPaginated :many
select id,
       name,
       allowed_content_types,
       max_allowed_object_size,
       public,
       disabled,
       locked,
       lock_reason,
       locked_at,
       created_at,
       updated_at
from storage.buckets
where id >= sqlc.arg('cursor')
limit sqlc.narg('limit');

-- name: SearchBucketsPaginated :many
select id,
       name,
       allowed_content_types,
       max_allowed_object_size,
       public,
       disabled,
       locked,
       lock_reason,
       locked_at,
       created_at,
       updated_at
from storage.buckets
where name ilike sqlc.narg('name')
limit sqlc.narg('limit') offset sqlc.narg('offset');

-- name: CountBuckets :one
select count(1) as count
from storage.buckets;

-- name: GetBucketSizeById :one
select sum(size) as size
from storage.objects
where bucket_id = sqlc.arg('id');

-- name: GetBucketObjectCountById :one
select count(1) as count
from storage.objects
where bucket_id = sqlc.arg('id');