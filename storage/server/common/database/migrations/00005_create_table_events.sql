-- +goose Up
-- +goose StatementBegin

create table if not exists storage.events
(
    id             text default 'event_' || gen_random_uuid() not null check ( storage.text_non_empty_trimmed_text(id) ),
    aggregate_type text  not null check ( storage.text_non_empty_trimmed_text(aggregate_type) ),
    aggregate_id   text  not null check ( storage.text_non_empty_trimmed_text(aggregate_id) ),
    type           text  not null check ( storage.text_non_empty_trimmed_text(type) ),
    payload        jsonb null,
    constraint events_id_pk primary key (id)
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

drop trigger if exists events_increment_version on storage.events;

drop trigger if exists events_set_updated_at on storage.events;

drop table if exists storage.events;

-- +goose StatementEnd
