-- name: EventCreate :one
insert into storage.events
    (aggregate_type, aggregate_id, event_type, payload)
values (sqlc.arg('aggregate_type'),
        sqlc.arg('aggregate_id'),
        sqlc.arg('event_type'),
        sqlc.narg('payload')) returning id;