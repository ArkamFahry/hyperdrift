-- +goose Up
-- +goose StatementBegin

create or replace function storage.on_object_create()
    returns trigger as
$$
begin
    new.id = 'object' || '_' || storage.gen_random_ulid();
    new.version = 0;
    new.created_at = now();

    return new;
end;
$$ language plpgsql;

create or replace function storage.on_object_update()
    returns trigger as
$$
begin
    new.version = new.version + 1;
    new.updated_at = now();

    return new;
end;
$$ language plpgsql;

create table if not exists storage.objects
(
    id               text                                           not null,
    version          int         default 0                          not null,
    bucket_id        text                                           not null,
    name             text                                           not null,
    mime_type        text        default 'application/octet-stream' not null,
    size             bigint      default 0                          not null,
    metadata         jsonb                                          null,
    upload_status    text        default 'pending'                  not null,
    last_accessed_at timestamptz                                    null,
    locked_at        timestamptz                                    null,
    created_at       timestamptz default now()                      not null,
    updated_at       timestamptz                                    null,
    constraint objects_id_primary_key primary key (id),
    constraint objects_bucket_id_foreign_key foreign key (bucket_id) references storage.buckets (id) on delete no action,
    constraint objects_id_version_unique unique (id, version),
    constraint objects_name_unique unique (bucket_id, name),
    constraint objects_id_check check ( trim(id) <> '' ),
    constraint objects_version_check check ( version >= 0 ),
    constraint objects_bucket_id_check check ( trim(bucket_id) <> '' ),
    constraint objects_name_check check ( trim(name) <> '' ),
    constraint objects_mime_type_check check ( trim(mime_type) <> '' ),
    constraint objects_size_check check ( size > 0 ),
    constraint objects_upload_status_check check ( upload_status in ('pending', 'completed') )
);

create index if not exists objects_bucket_id_name_index on storage.objects using btree (bucket_id, name);

create or replace trigger object_on_create
    before insert
    on storage.objects
    for each row
execute function storage.on_object_create();

create or replace trigger object_on_update
    before update
    on storage.objects
    for each row
execute function storage.on_object_update();



-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

drop trigger if exists object_on_create on storage.objects;

drop trigger if exists object_on_update on storage.objects;

drop index if exists storage.objects_bucket_id_name_idx;

drop table if exists storage.objects;

drop function if exists storage.on_object_update;

drop function if exists storage.on_object_create;

-- +goose StatementEnd
