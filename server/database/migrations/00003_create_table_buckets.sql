-- +goose Up
-- +goose StatementBegin

create or replace function storage.on_bucket_create()
    returns trigger as
$$
begin
    new.id = 'bucket_' || storage.gen_random_ulid();
    new.version = 0;
    new.created_at = now();

    if new.allowed_mime_types is not null then
        new.allowed_mime_types = array(select distinct unnest(new.allowed_mime_types));
    end if;

    return new;
end;
$$ language plpgsql;

create or replace function storage.on_bucket_update()
    returns trigger as
$$
begin
    new.version = new.version + 1;
    new.updated_at = now();

    return new;
end;
$$ language plpgsql;

create table if not exists storage.buckets
(
    id                      text                      not null,
    version                 int         default 0     not null,
    name                    text                      not null,
    allowed_mime_types      text[]                    null,
    max_allowed_object_size bigint                    null,
    public                  boolean     default false not null,
    disabled                boolean     default false not null,
    locked                  boolean     default false not null,
    lock_reason             text                      null,
    locked_at               timestamptz               null,
    created_at              timestamptz default now() not null,
    updated_at              timestamptz               null,
    constraint buckets_id_primary_key primary key (id),
    constraint buckets_id_version_unique unique (id, version),
    constraint buckets_name_unique unique (name),
    constraint buckets_id_check check ( trim(id) <> '' ),
    constraint buckets_version_check check ( version >= 0 ),
    constraint buckets_name_check check ( trim(name) <> '' ),
    constraint buckets_allowed_mime_types_check check ( allowed_mime_types is null or
                                                        (array_length(allowed_mime_types, 1) > 0 and
                                                         not '' = any (allowed_mime_types)) ),
    constraint buckets_max_allowed_object_size_check check ( max_allowed_object_size is null or max_allowed_object_size > 0 ),
    constraint buckets_lock_reason_check check ( lock_reason is null or trim(lock_reason) <> '' )
);

create index if not exists buckets_name_index on storage.buckets using btree (name);

create or replace trigger bucket_on_create
    before insert
    on storage.buckets
    for each row
execute function storage.on_bucket_create();

create or replace trigger bucket_on_update
    before update
    on storage.buckets
    for each row
execute function storage.on_bucket_update();


-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

drop trigger if exists buckets_on_update on storage.buckets;

drop trigger if exists buckets_on_create on storage.buckets;

drop index if exists storage.buckets_name_idx;

drop table if exists storage.buckets;

drop function if exists storage.on_bucket_create();

drop function if exists storage.on_bucket_update();

-- +goose StatementEnd