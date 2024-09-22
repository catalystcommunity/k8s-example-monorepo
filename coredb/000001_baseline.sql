-- +goose Up
-- Used for ULID generation
CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- The ULID generator
CREATE OR REPLACE FUNCTION generate_ulid() RETURNS uuid
    AS $$
        SELECT (lpad(to_hex(floor(extract(epoch FROM clock_timestamp()) * 1000)::bigint), 12, '0') || encode(gen_random_bytes(10), 'hex'))::uuid;
    $$ LANGUAGE SQL;


-- Our roles are constrained to a set enforced by the database
CREATE TYPE user_role AS ENUM ('user', 'support', 'admin');

-- Users are of course, users of the system. We could of course limit access to this table
-- to specific roles or something since this is a shared db, but we won't for now
create table users
(
    id         uuid      default generate_ulid() not null primary key,
    created_at timestamp default now() not null,
    updated_at timestamp default now() not null,
    username   text not null,
    email      text not null,
    password   bytea not null,
    salt       bytea not null,
    roles      user_role[] default ARRAY['user'::user_role] not null
);

-- These are what "thing" is being borrowed, so we could add "car" or something
-- We can also limit what kind of users can add types so "support" could but not users
create table thing_types
(
    -- Technically using an int primary key would be faster, and this will be a limited table
    -- but we want consistency over performance here
    id               uuid      default generate_ulid() not null primary key,
    created_at       timestamp default now(),
    updated_at       timestamp default now(),
    name             text not null,
    description      text -- We don't need a description, that's a nice to have option
);

create table owned_things
(
    id               uuid      default generate_ulid() not null primary key,
    created_at       timestamp default now(),
    updated_at       timestamp default now(),
    name             text,
    thing_type       uuid not null references thing_types on delete cascade,
    owner            uuid not null references users
);

-- We could put this as data on the owned_things table, but then we can't retain history
create table borrows 
(
    id               uuid      default generate_ulid() not null primary key,
    created_at       timestamp default now(),
    updated_at       timestamp default now(),
    borrower         uuid not null references users,
    borrowed_at      timestamp default now(),
    borrowed_until   timestamp,
    returned_at      timestamp,
    reposessed       boolean default false
);
