-- name: BucketCreate :one
insert into storage.buckets
    (name, allowed_mime_types, max_allowed_object_size, public)
values (sqlc.arg('name'),
        sqlc.narg('allowed_mime_types'),
        sqlc.narg('max_allowed_object_size'),
        sqlc.arg('public'))
returning id;

-- name: BucketUpdate :exec
update storage.buckets
set max_allowed_object_size = coalesce(sqlc.narg('max_allowed_object_size'), max_allowed_object_size),
    public                  = coalesce(sqlc.narg('public'), public),
    allowed_mime_types      = coalesce(sqlc.narg('allowed_mime_types'), allowed_mime_types)
where id = sqlc.arg('id');

-- name: BucketDisable :exec
update storage.buckets
set disabled = true
where id = sqlc.arg('id');

-- name: BucketEnable :exec
update storage.buckets
set disabled = false
where id = sqlc.arg('id');


-- name: BucketLock :exec
update storage.buckets
set locked      = true,
    lock_reason = sqlc.arg('lock_reason')::text,
    locked_at   = now()
where id = sqlc.arg('id');

-- name: BucketUnlock :exec
update storage.buckets
set locked      = false,
    lock_reason = null,
    locked_at   = null
where id = sqlc.arg('id');

-- name: BucketDelete :exec
delete
from storage.buckets
where id = sqlc.arg('id');

-- name: BucketGetById :one
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

-- name: BucketGetByName :one
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

-- name: BucketListAll :many
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

-- name: BucketListPaginated :many
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
where id >= sqlc.arg('cursor')
limit sqlc.arg('limit');

-- name: BucketSearch :many
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
where name ilike '%' || sqlc.arg('name')::text || '%';

-- name: BucketCount :one
select count(1) as count
from storage.buckets;

-- name: BucketGetSizeById :one
select object.bucket_id as id, bucket.name as name, SUM(object.size) as size
from storage.objects as object
         join storage.buckets as bucket on object.bucket_id = bucket.id
where object.bucket_id = sqlc.arg('id')
group by object.bucket_id, bucket.name;

-- name: BucketGetObjectCountById :one
select bucket_id as id, count(1) as count
from storage.objects
where bucket_id = sqlc.arg('id')
group by bucket_id;