-- name: ObjectCreate :one
insert into storage.objects
    (bucket_id, name, mime_type, size, metadata, upload_status)
values (sqlc.arg('bucket_id'),
        sqlc.arg('name'),
        sqlc.narg('content_type'),
        sqlc.arg('size'),
        sqlc.arg('metadata'),
        sqlc.arg('upload_status'))
returning id;

-- name: ObjectUpdateUploadStatus :exec
update storage.objects
set upload_status = sqlc.arg('upload_status')
where id = sqlc.arg('id');

-- name: ObjectUpdateLastAccessedAt :exec
update storage.objects
set last_accessed_at = now()
where id = sqlc.arg('id');

-- name: ObjectUpdate :exec
update storage.objects
set size      = coalesce(sqlc.narg('size'), size),
    mime_type = coalesce(sqlc.narg('mime_type'), mime_type),
    metadata  = coalesce(sqlc.narg('metadata'), metadata)
where id = sqlc.arg('id');

-- name: ObjectDelete :exec
delete
from storage.objects
where id = sqlc.arg('id');

-- name: ObjectGetById :one
select id,
       version,
       bucket_id,
       name,
       mime_type,
       size,
       metadata,
       upload_status,
       last_accessed_at,
       created_at,
       updated_at
from storage.objects
where id = sqlc.arg('id')
limit 1;

-- name: ObjectGetByIdWithBucketName :one
select object.id,
       object.version,
       object.bucket_id,
       bucket.name as bucket_name,
       object.name,
       object.mime_type,
       object.size,
       object.metadata,
       object.upload_status,
       object.last_accessed_at,
       object.created_at,
       object.updated_at
from storage.objects as object
         inner join storage.buckets as bucket on object.bucket_id = bucket.id
where object.id = sqlc.arg('id')
limit 1;

-- name: ObjectGetByName :one
select id,
       version,
       bucket_id,
       name,
       mime_type,
       size,
       metadata,
       upload_status,
       last_accessed_at,
       created_at,
       updated_at
from storage.objects
where name = sqlc.arg('name')
limit 1;

-- name: ObjectGetByBucketIdAndName :one
select id,
       version,
       bucket_id,
       name,
       mime_type,
       size,
       metadata,
       upload_status,
       last_accessed_at,
       created_at,
       updated_at
from storage.objects
where bucket_id = sqlc.arg('bucket_id')
  and name = sqlc.arg('name')
limit 1;

-- name: ObjectsListBucketIdPaged :many
select id,
       bucket_id,
       name,
       mime_type,
       size,
       metadata,
       upload_status,
       last_accessed_at,
       created_at,
       updated_at
from storage.objects
where bucket_id = sqlc.arg('bucket_id')
limit sqlc.arg('limit') offset sqlc.arg('offset');

-- name: ObjectSearchByBucketNameAndObjectPath :many
select object.id,
       object.version,
       object.bucket_id,
       object.name,
       object.mime_type,
       object.size,
       object.metadata,
       object.upload_status,
       object.last_accessed_at,
       object.created_at,
       object.updated_at
from storage.objects as object
where object.bucket_id = (select bucket.id from storage.buckets as bucket where bucket.name = sqlc.arg('bucket_name'))
  and object.name ilike sqlc.arg('object_path')::text || '%'
limit sqlc.arg('limit') offset sqlc.arg('offset');