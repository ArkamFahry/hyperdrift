-- name: CreateEvent :exec
insert into storage.events
    (id, name, payload, status, producer, timestamp)
values (sqlc.arg('id'),
        sqlc.arg('name'),
        sqlc.arg('payload'),
        sqlc.arg('status'),
        sqlc.arg('producer'),
        sqlc.arg('timestamp'));