-- name: CreateObject :one
insert into storage.objects
    (bucket_id, name, content_type, size, public, metadata, upload_status)
values (sqlc.arg('bucket_id'),
        sqlc.arg('name'),
        sqlc.narg('content_type'),
        sqlc.arg('size'),
        sqlc.arg('public'),
        sqlc.arg('metadata'),
        sqlc.arg('upload_status')) returning id;

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
set size         = coalesce(sqlc.arg('size'), size),
    content_type = coalesce(sqlc.arg('content_type'), content_type),
    metadata     = coalesce(sqlc.arg('metadata'), metadata)
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
       bucket_id,
       name,
       path_tokens,
       content_type,
       size,
       public,
       metadata,
       upload_status,
       last_accessed_at,
       created_at,
       updated_at
from storage.objects
where id = sqlc.arg('id')
limit 1;

-- name: GetObjectByIdWithBucketName :one
select o.id,
       o.bucket_id,
       b.name as bucket_name,
       o.name,
       o.path_tokens,
       o.content_type,
       o.size,
       o.public,
       o.metadata,
       o.upload_status,
       o.last_accessed_at,
       o.created_at,
       o.updated_at
from storage.objects as o
inner join storage.buckets as b on o.bucket_id = b.id
where o.id = sqlc.arg('id')
limit 1;

-- name: GetObjectByName :one
select id,
       bucket_id,
       name,
       path_tokens,
       content_type,
       size,
       public,
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
       bucket_id,
       name,
       path_tokens,
       content_type,
       size,
       public,
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
       path_tokens,
       content_type,
       size,
       public,
       metadata,
       upload_status,
       last_accessed_at,
       created_at,
       updated_at
from storage.objects
where bucket_id = sqlc.arg('bucket_id')
limit sqlc.arg('limit') offset sqlc.arg('offset');

-- name: SearchObjectsByPath :many
select id::text,
       version::int,
       name::text,
       bucket_id::text,
       bucket_name::text,
       content_type::text,
       size::bigint,
       public::boolean,
       metadata::jsonb,
       upload_status::text,
       last_accessed_at::timestamptz,
       created_at::timestamptz,
       updated_at::timestamptz
from storage.objects_search(sqlc.arg('bucket_name')::text, sqlc.arg('object_path')::text, sqlc.narg('level')::int,
                            sqlc.narg('limit')::int, sqlc.narg('offset')::int);