-- +goose Up
-- +goose StatementBegin

create extension if not exists pgcrypto;

create or replace function storage.gen_random_ulid()
    returns text
as
$$
declare
    -- Crockford's Base32
    encoding  bytea = '0123456789ABCDEFGHJKMNPQRSTVWXYZ';
    timestamp bytea = E'\\000\\000\\000\\000\\000\\000';
    output    text  = '';
    unix_time bigint;
    ulid      bytea;
BEGIN
    unix_time = (extract(epoch from clock_timestamp()) * 1000)::bigint;
    timestamp = set_byte(timestamp, 0, (unix_time >> 40)::bit(8)::integer);
    timestamp = set_byte(timestamp, 1, (unix_time >> 32)::bit(8)::integer);
    timestamp = set_byte(timestamp, 2, (unix_time >> 24)::bit(8)::integer);
    timestamp = set_byte(timestamp, 3, (unix_time >> 16)::bit(8)::integer);
    timestamp = set_byte(timestamp, 4, (unix_time >> 8)::bit(8)::integer);
    timestamp = set_byte(timestamp, 5, unix_time::bit(8)::integer);

    -- 10 entropy bytes
    ulid = timestamp || gen_random_bytes(10);

    -- Encode the timestamp
    output = output || chr(get_byte(encoding, (get_byte(ulid, 0) & 224) >> 5));
    output = output || chr(get_byte(encoding, (get_byte(ulid, 0) & 31)));
    output = output || chr(get_byte(encoding, (get_byte(ulid, 1) & 248) >> 3));
    output = output || chr(get_byte(encoding, ((get_byte(ulid, 1) & 7) << 2) | ((get_byte(ulid, 2) & 192) >> 6)));
    output = output || chr(get_byte(encoding, (get_byte(ulid, 2) & 62) >> 1));
    output = output || chr(get_byte(encoding, ((get_byte(ulid, 2) & 1) << 4) | ((get_byte(ulid, 3) & 240) >> 4)));
    output = output || chr(get_byte(encoding, ((get_byte(ulid, 3) & 15) << 1) | ((get_byte(ulid, 4) & 128) >> 7)));
    output = output || chr(get_byte(encoding, (get_byte(ulid, 4) & 124) >> 2));
    output = output || chr(get_byte(encoding, ((get_byte(ulid, 4) & 3) << 3) | ((get_byte(ulid, 5) & 224) >> 5)));
    output = output || chr(get_byte(encoding, (get_byte(ulid, 5) & 31)));

    -- Encode the entropy
    output = output || chr(get_byte(encoding, (get_byte(ulid, 6) & 248) >> 3));
    output = output || chr(get_byte(encoding, ((get_byte(ulid, 6) & 7) << 2) | ((get_byte(ulid, 7) & 192) >> 6)));
    output = output || chr(get_byte(encoding, (get_byte(ulid, 7) & 62) >> 1));
    output = output || chr(get_byte(encoding, ((get_byte(ulid, 7) & 1) << 4) | ((get_byte(ulid, 8) & 240) >> 4)));
    output = output || chr(get_byte(encoding, ((get_byte(ulid, 8) & 15) << 1) | ((get_byte(ulid, 9) & 128) >> 7)));
    output = output || chr(get_byte(encoding, (get_byte(ulid, 9) & 124) >> 2));
    output = output || chr(get_byte(encoding, ((get_byte(ulid, 9) & 3) << 3) | ((get_byte(ulid, 10) & 224) >> 5)));
    output = output || chr(get_byte(encoding, (get_byte(ulid, 10) & 31)));
    output = output || chr(get_byte(encoding, (get_byte(ulid, 11) & 248) >> 3));
    output = output || chr(get_byte(encoding, ((get_byte(ulid, 11) & 7) << 2) | ((get_byte(ulid, 12) & 192) >> 6)));
    output = output || chr(get_byte(encoding, (get_byte(ulid, 12) & 62) >> 1));
    output = output || chr(get_byte(encoding, ((get_byte(ulid, 12) & 1) << 4) | ((get_byte(ulid, 13) & 240) >> 4)));
    output = output || chr(get_byte(encoding, ((get_byte(ulid, 13) & 15) << 1) | ((get_byte(ulid, 14) & 128) >> 7)));
    output = output || chr(get_byte(encoding, (get_byte(ulid, 14) & 124) >> 2));
    output = output || chr(get_byte(encoding, ((get_byte(ulid, 14) & 3) << 3) | ((get_byte(ulid, 15) & 224) >> 5)));
    output = output || chr(get_byte(encoding, (get_byte(ulid, 15) & 31)));

    RETURN output;
END
$$
    LANGUAGE plpgsql
    VOLATILE;

create or replace function storage.text_non_empty_trimmed_text(val text) returns boolean as
$$
begin
    return trim(val) <> '';
end;
$$ language plpgsql;

create or replace function storage.text_null_or_non_empty_trimmed_text(val text) returns boolean as
$$
begin
    if val is null then
        return true;
    end if;

    return trim(val) <> '';
end;
$$ language plpgsql;

create or replace function storage.array_null_or_contains_empty_trimmed_text(val text[]) returns boolean as
$$
declare
    i int;
begin
    if val is null then
        return true;
    end if;

    for i in array_lower(val, 1) .. coalesce(array_upper(val, 1), 0)
        loop
            if trim(val[i]) = '' then
                return false;
            end if;
        end loop;

    return true;
end;
$$ language plpgsql;

create or replace function storage.array_null_or_text_values_unique(val text[])
    returns boolean as
$$
begin
    if val is null then
        return true;
    end if;

    return array_length(val, 1) = array_length(array(select distinct unnest(val)), 1);
end;
$$ language plpgsql;

create or replace function storage.bigint_null_or_non_zero_bigint(val bigint)
    returns boolean as
$$
begin
    if val is null then
        return true;
    end if;

    return val <> 0;
end
$$ language plpgsql;



create or replace function storage.on_create()
    returns trigger as
$$
begin
    new.id = tg_table_name || '_' || storage.gen_random_ulid();
    new.version = 0;
    new.created_at = now();

    return new;
end;
$$ language plpgsql;

create or replace function storage.on_update()
    returns trigger as
$$
begin
    new.version = new.version + 1;
    new.updated_at = now();

    return new;
end;
$$ language plpgsql;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

drop function if exists storage.on_update();

drop function if exists storage.on_create();

drop function if exists storage.bigint_null_or_non_zero_bigint();

drop function if exists storage.array_null_or_text_values_unique();

drop function if exists storage.array_null_or_contains_empty_trimmed_text();

drop function if exists storage.text_null_or_non_empty_trimmed_text();

drop function if exists storage.text_non_empty_trimmed_text();

drop function if exists storage.gen_random_ulid();

-- +goose StatementEnd
