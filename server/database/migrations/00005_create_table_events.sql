-- +goose Up
-- +goose StatementBegin

create table if not exists storage.events
(
    id             text          not null check ( storage.text_non_empty_trimmed_text(id) ),
    version        int default 0 not null check ( version >= 0 ),
    aggregate_type text          not null check ( storage.text_non_empty_trimmed_text(aggregate_type) ),
    aggregate_id   text          not null check ( storage.text_non_empty_trimmed_text(aggregate_id) ),
    event_type     text          not null check ( storage.text_non_empty_trimmed_text(event_type) ),
    payload        jsonb         null,
    created_at     timestamptz   not null,
    constraint events_id_primary_key primary key (id)
);

create or replace trigger events_on_create
    before insert
    on storage.events
    for each row
execute function storage.on_create('event');

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

drop trigger if exists events_on_create on storage.events;

drop table if exists storage.events;

-- +goose StatementEnd
