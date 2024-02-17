-- name: CreateObject :one
insert into storage.objects
    (bucket_id, name, mime_type, size, metadata, upload_status)
values (sqlc.arg('bucket_id'),
        sqlc.arg('name'),
        sqlc.narg('content_type'),
        sqlc.arg('size'),
        sqlc.arg('metadata'),
        sqlc.arg('upload_status'))
returning id;

-- name: UpdateObjectUploadStatus :exec
update storage.objects
set upload_status = sqlc.arg('upload_status')
where id = sqlc.arg('id');

-- name: UpdateObjectLastAccessedAt :exec
update storage.objects
set last_accessed_at = now()
where id = sqlc.arg('id');

-- name: UpdateObject :exec
update storage.objects
set size      = coalesce(sqlc.arg('size'), size),
    mime_type = coalesce(sqlc.arg('mime_type'), mime_type),
    metadata  = coalesce(sqlc.arg('metadata'), metadata)
where id = sqlc.arg('id');

-- name: MakeObjectPublic :exec
update storage.objects
set public = true
where id = sqlc.arg('id');

-- name: MakeObjectPrivate :exec
update storage.objects
set public = false
where id = sqlc.arg('id');

-- name: MergeObjectMetadata :exec
update storage.objects
set metadata = metadata || sqlc.arg('metadata')
where id = sqlc.arg('id');

-- name: DeleteObject :exec
delete
from storage.objects
where id = sqlc.arg('id');

-- name: GetObjectById :one
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

-- name: GetObjectByIdWithBucketName :one
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

-- name: GetObjectByName :one
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

-- name: GetObjectByBucketIdAndName :one
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

-- name: ListObjectsByBucketIdPaged :many
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

-- name: SearchObjectsByBucketNameAndPath :many
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