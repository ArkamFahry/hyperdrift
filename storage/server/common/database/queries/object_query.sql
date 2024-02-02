-- name: CreateObject :exec
insert into storage.objects
    (id, bucket_id, name, content_type, size, public, metadata)
values (sqlc.arg('id'),
        sqlc.arg('bucket_id'),
        sqlc.arg('name'),
        sqlc.arg('content_type'),
        sqlc.arg('size'),
        sqlc.arg('public'),
        sqlc.arg('metadata'));

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
       bucket::text,
       name::text,
       content_type::text,
       size::bigint,
       public::boolean,
       metadata::jsonb,
       upload_status::text,
       last_accessed_at::timestamptz,
       created_at::timestamptz,
       updated_at::timestamptz
from storage.objects_search(sqlc.arg('bucket_name')::text, sqlc.arg('path_prefix')::text, sqlc.narg('levels')::int,
                            sqlc.narg('limit')::int, sqlc.narg('offset')::int);