-- name: CreateEvent :exec
insert into storage.events
    (id, name, content, status, retries, expires_at, created_at)
values (sqlc.arg('id'),
        sqlc.arg('name'),
        sqlc.arg('content'),
        sqlc.arg('status'),
        sqlc.arg('retries'),
        sqlc.narg('expires_at'),
        sqlc.arg('created_at'));