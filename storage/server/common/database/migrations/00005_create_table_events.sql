-- +goose Up
-- +goose StatementBegin

create table if not exists storage.events
(
    id             text  not null check ( storage.text_non_empty_trimmed_text(id) ),
    aggregate_type text  not null check ( storage.text_non_empty_trimmed_text(aggregate_type) ),
    aggregate_id   text  not null check ( storage.text_non_empty_trimmed_text(aggregate_id) ),
    type           text  not null check ( storage.text_non_empty_trimmed_text(type) ),
    payload        jsonb null,
    constraint events_id_pk primary key (id)
);

create or replace trigger events_set_updated_at
    before update
    on storage.events
    for each row
execute function storage.set_updated_at();

create or replace trigger events_increment_version
    before update
    on storage.events
    for each row
execute function storage.increment_version();

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

drop trigger if exists events_increment_version on storage.events;

drop trigger if exists events_set_updated_at on storage.events;

drop table if exists storage.events;

-- +goose StatementEnd
