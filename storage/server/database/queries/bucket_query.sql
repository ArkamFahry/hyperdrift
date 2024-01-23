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

-- name: AddAllowedMimeTypesToBucket :exec
update storage.buckets
set allowed_content_types = array_append(allowed_content_types, sqlc.arg('mime_type')::text[])
where id = sqlc.arg('id');

-- name: RemoveAllowedMimeTypesFromBucket :exec
update storage.buckets
set allowed_content_types = array_remove(allowed_content_types, sqlc.arg('mime_type')::text[])
where id = sqlc.arg('id');

-- name: UpdateBucketMaxAllowedObjectSize :exec
update storage.buckets
set max_allowed_object_size = sqlc.arg('max_allowed_object_size')
where id = sqlc.arg('id');

-- name: DisableBucket :exec
update storage.buckets
set disabled = true
where id = sqlc.arg('id');

-- name: EnableBucket :exec
update storage.buckets
set disabled = false
where id = sqlc.arg('id');

-- name: MakeBucketPublic :exec
update storage.buckets
set public = true
where id = sqlc.arg('id');

-- name: MakeBucketPrivate :exec
update storage.buckets
set public = false
where id = sqlc.arg('id');

-- name: LockBucket :exec
update storage.buckets
set locked      = true,
    lock_reason = sqlc.arg('lock_reason'),
    locked_at   = now()
where id = sqlc.arg('id');

-- name: UnlockBucket :exec
update storage.buckets
set locked      = false,
    lock_reason = null,
    locked_at   = null
where id = sqlc.arg('id');

-- name: DeleteBucket :exec
delete
from storage.buckets
where id = sqlc.arg('id');

-- name: GetBucketById :one
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
where id = sqlc.arg('id')
limit 1;

-- name: GetBucketByName :one
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
where name = sqlc.arg('name')
limit 1;

-- name: ListAllBuckets :many
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
from storage.buckets;

-- name: ListBucketsPaged :many
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
limit sqlc.narg('limit') offset sqlc.narg('offset');

-- name: SearchBucketsPaged :many
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

-- name: GetBucketSizeByName :one
select sum(size) as size
from storage.objects
where bucket_id = (select id from storage.buckets where storage.buckets.name = sqlc.arg('name'));

-- name: GetBucketObjectCountById :one
select count(1) as count
from storage.objects
where bucket_id = sqlc.arg('id');

-- name: GetBucketObjectCountByName :one
select count(1) as count
from storage.objects
where bucket_id = (select id from storage.buckets where storage.buckets.name = sqlc.arg('name'));