-- name: EventCreate :one
insert into storage.events
    (aggregate_type, event_type, payload)
values (sqlc.arg('aggregate_type'),
        sqlc.arg('event_type'),
        sqlc.narg('payload')) returning id;