-- +goose Up
-- +goose StatementBegin

create table if not exists storage.events
(
    id             text default 'events_' || storage.gen_random_ulid() not null check ( storage.text_non_empty_trimmed_text(id) ),
    aggregate_type text  not null check ( storage.text_non_empty_trimmed_text(aggregate_type) ),
    aggregate_id   text  not null check ( storage.text_non_empty_trimmed_text(aggregate_id) ),
    type           text  not null check ( storage.text_non_empty_trimmed_text(type) ),
    payload        jsonb null,
    constraint events_id_primary_key primary key (id)
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

drop table if exists storage.events;

-- +goose StatementEnd
