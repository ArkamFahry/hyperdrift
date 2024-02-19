-- +goose Up
-- +goose StatementBegin

create or replace function storage.on_event_create()
    returns trigger as
$$
begin
    new.id = 'event' || '_' || storage.gen_random_ulid();
    new.version = 0;
    new.created_at = now();

    return new;
end;
$$ language plpgsql;

create table if not exists storage.events
(
    id             text          not null,
    version        int default 0 not null,
    aggregate_type text          not null,
    event_type     text          not null,
    payload        jsonb         null,
    created_at     timestamptz   not null,
    constraint events_id_primary_key primary key (id),
    constraint events_id_version_unique unique (id, version),
    constraint events_id_check check ( trim(id) <> '' ),
    constraint events_version_check check ( version >= 0 ),
    constraint events_aggregate_type_check check ( trim(aggregate_type) <> '' ),
    constraint events_event_type_check check ( trim(event_type) <> '' )
);

create or replace trigger event_on_create
    before insert
    on storage.events
    for each row
execute function storage.on_event_create();

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

drop trigger if exists events_on_create on storage.events;

drop table if exists storage.events;

drop function if exists storage.on_event_create();

-- +goose StatementEnd
