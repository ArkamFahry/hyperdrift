-- name: CreateEvent :one
insert into storage.events
    (aggregate_type, aggregate_id, type, payload)
values (sqlc.arg('aggregate_type'),
        sqlc.arg('aggregate_id'),
        sqlc.arg('type'),
        sqlc.narg('payload')) returning id;