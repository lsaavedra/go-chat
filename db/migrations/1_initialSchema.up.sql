CREATE EXTENSION IF NOT EXISTS "uuid-ossp" SCHEMA public;

-- create role
DO $$
    BEGIN
        CREATE ROLE read_only;
    EXCEPTION WHEN duplicate_object THEN RAISE NOTICE '%, skipping', SQLERRM USING ERRCODE = SQLSTATE;
    END
$$;

-- create DataBase schema
CREATE SCHEMA IF NOT EXISTS "chatrooms";

-- Apply grants on all created resources in this applications schema
GRANT USAGE ON SCHEMA chatrooms to read_only;
GRANT SELECT ON ALL TABLES IN SCHEMA chatrooms TO read_only;
GRANT SELECT ON ALL SEQUENCES IN SCHEMA chatrooms TO read_only;
ALTER DEFAULT PRIVILEGES IN SCHEMA chatrooms GRANT SELECT ON TABLES TO read_only;
ALTER DEFAULT PRIVILEGES IN SCHEMA chatrooms GRANT SELECT ON SEQUENCES TO read_only;

-- trigger function
CREATE OR REPLACE FUNCTION chatrooms.updated_at_trigger()
    RETURNS TRIGGER
    LANGUAGE plpgsql
AS
$$
BEGIN
    NEW.updated_at = now();
RETURN NEW;
END;
$$;

-- users table
CREATE TABLE IF NOT EXISTS "users"
(
    "id"                uuid    default uuid_generate_v4(),
    "first_name"              varchar(32) not null,
    "last_name"              varchar(256) not null,
    "nickname" varchar(256) not null,
    "password" varchar(256) not null,
    "email" varchar(256) not null,
    "created_at" timestamp with time zone default now(),
    "updated_at" timestamp with time zone default now(),
    "deleted_at" timestamp with time zone,
    PRIMARY KEY ("id")
 );

