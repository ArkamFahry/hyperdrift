-- name: CreateBucket :one
insert into storage.buckets
    (name, allowed_mime_types, max_allowed_object_size, public)
values (sqlc.arg('name'),
        sqlc.narg('allowed_mime_types'),
        sqlc.narg('max_allowed_object_size'),
        sqlc.arg('public'))
returning id;

-- name: UpdateBucket :exec
update storage.buckets
set max_allowed_object_size = coalesce(sqlc.narg('max_allowed_object_size'), max_allowed_object_size),
    public                  = coalesce(sqlc.narg('public'), public),
    allowed_mime_types      = coalesce(sqlc.narg('allowed_mime_types'), allowed_mime_types)
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
    lock_reason = sqlc.arg('lock_reason')::text,
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
       version,
       name,
       allowed_mime_types,
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
       allowed_mime_types,
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
       allowed_mime_types,
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
       allowed_mime_types,
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
limit sqlc.arg('limit');

-- name: SearchBucketsPaginated :many
select id,
       name,
       allowed_mime_types,
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
select o.bucket_id as id, b.name as name, SUM(o.size) as size
from storage.objects as o
         join storage.buckets as b on o.bucket_id = b.id
where o.bucket_id = sqlc.arg('id')
group by o.bucket_id, b.name;

-- name: GetBucketObjectCountById :one
select o.bucket_id as id, count(1) as count
from storage.objects as o
where o.bucket_id = sqlc.arg('id')
group by o.bucket_id;