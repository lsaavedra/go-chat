CREATE EXTENSION IF NOT EXISTS "uuid-ossp" SCHEMA public;

-- create role
DO $$
    BEGIN
        CREATE ROLE read_only;
    EXCEPTION WHEN duplicate_object THEN RAISE NOTICE '%, skipping', SQLERRM USING ERRCODE = SQLSTATE;
    END
$$;

-- create DataBase schema
CREATE SCHEMA IF NOT EXISTS "pioneer";

-- Apply grants on all created resources in this applications schema
GRANT USAGE ON SCHEMA pioneer to read_only;
GRANT SELECT ON ALL TABLES IN SCHEMA pioneer TO read_only;
GRANT SELECT ON ALL SEQUENCES IN SCHEMA pioneer TO read_only;
ALTER DEFAULT PRIVILEGES IN SCHEMA pioneer GRANT SELECT ON TABLES TO read_only;
ALTER DEFAULT PRIVILEGES IN SCHEMA pioneer GRANT SELECT ON SEQUENCES TO read_only;

-- trigger function
CREATE OR REPLACE FUNCTION pioneer.updated_at_trigger()
    RETURNS TRIGGER
    LANGUAGE plpgsql
AS
$$
BEGIN
    NEW.updated_at = now();
RETURN NEW;
END;
$$;

-- pioneer company table
CREATE TABLE IF NOT EXISTS "pioneer"."company"
(
    "id"                uuid    default uuid_generate_v4(),
    "code"              varchar(32),
    "name"              varchar(256),
    PRIMARY KEY ("id")
    );

-- indexes of company table
CREATE INDEX IF NOT EXISTS idx_company_id on "pioneer"."company" (id);

-- pioneer borrower table
CREATE TABLE IF NOT EXISTS "pioneer"."borrower"
(
    "id"                uuid    default uuid_generate_v4(),
    "first_name"        varchar(100),
    "middle_name"       varchar(100),
    "last_name"         varchar(100),
    "email"             varchar(100),
    "phone_number"      varchar(32),
    "address_line1"     varchar(100),
    "address_line2"     varchar(100),
    "city"              varchar(100),
    "state"             varchar(32),
    "zip_code"          varchar(32),
    "ssn"               varchar(16),
    "date_of_birth"     date,
    "income"            numeric,
    "scp_consent_date"  timestamp with time zone,
    "created_at"        timestamp with time zone default now(),
    "updated_at"        timestamp with time zone default now(),
    PRIMARY KEY ("id")
    );

CREATE INDEX IF NOT EXISTS idx_borrower_id on "pioneer"."borrower" (id);

DROP TRIGGER IF EXISTS borrower_updated_at_trigger ON "pioneer"."borrower";
CREATE TRIGGER borrower_updated_at_trigger
    BEFORE UPDATE
    ON "pioneer"."borrower"
    FOR EACH ROW
    EXECUTE PROCEDURE "pioneer".updated_at_trigger();

-- pioneer applications table
CREATE TABLE IF NOT EXISTS "pioneer"."application"
(
    "id"                uuid    default uuid_generate_v4(),
    "borrower_id"       uuid,
    "borrow_amount_usd" numeric,
    "installer_id"      uuid,
    "created_at"        timestamp with time zone default now(),
    "updated_at"        timestamp with time zone default now(),
    "deleted_at"        timestamp with time zone,
    "deleted"           boolean not null default false,
    PRIMARY KEY ("id"),
    CONSTRAINT fk_application_borrower
    FOREIGN KEY("borrower_id")
    REFERENCES "pioneer"."borrower"("id")
    ON DELETE CASCADE,
    CONSTRAINT fk_application_installer
    FOREIGN KEY("installer_id")
    REFERENCES "pioneer"."company"("id")
    ON DELETE CASCADE
    );

-- indexes of application table
CREATE INDEX IF NOT EXISTS idx_application_id on "pioneer"."application" (id);

CREATE INDEX IF NOT EXISTS idx_application_borrower_id on "pioneer"."application" (borrower_id);

-- trigger of application table
DROP TRIGGER IF EXISTS application_updated_at_trigger ON "pioneer"."application";
CREATE TRIGGER application_updated_at_trigger
    BEFORE UPDATE
    ON "pioneer"."application"
    FOR EACH ROW
    EXECUTE PROCEDURE "pioneer".updated_at_trigger();


-- pioneer application_note table
CREATE TABLE IF NOT EXISTS "pioneer"."application_note"
(
    "id"                    uuid    default uuid_generate_v4(),
    "application_id"        uuid,
    "note"                  text,
    PRIMARY KEY ("id"),
    CONSTRAINT fk_application_note
    FOREIGN KEY("application_id")
    REFERENCES "pioneer"."application"("id")
    ON DELETE CASCADE
    );

-- indexes of application_note table
CREATE INDEX IF NOT EXISTS idx_application_note_id on "pioneer"."application_note" (id);
CREATE INDEX IF NOT EXISTS idx_application_note_application_id on "pioneer"."application_note" (application_id);

-- pioneer application_state table
CREATE TABLE IF NOT EXISTS "pioneer"."application_state"
(
    "id"                    uuid    default uuid_generate_v4(),
    "application_id"        uuid,
    "application_state"     varchar(32),
    "date"                  timestamp with time zone,
    PRIMARY KEY ("id"),
    CONSTRAINT fk_application_state
    FOREIGN KEY("application_id")
    REFERENCES "pioneer"."application"("id")
    ON DELETE CASCADE
    );

-- indexes of application_state table
CREATE INDEX IF NOT EXISTS idx_application_state_id on "pioneer"."application_state" (id);
CREATE INDEX IF NOT EXISTS idx_application_state_application_id on "pioneer"."application_state" (application_id);

-- fn_application_state_trigger function to log application_state changes

CREATE OR REPLACE FUNCTION "pioneer".fn_application_state_trigger() RETURNS trigger LANGUAGE plpgsql AS
$$
BEGIN
    INSERT INTO pioneer.application_state_history
        (
        application_id,
        application_state,
        date)
        VALUES (
        NEW.application_id,
        NEW.application_state,
        NEW.date);
RETURN NEW;
END;
$$;

-- trigger to populate application_state_history table
DROP TRIGGER IF EXISTS application_state_trigger ON "pioneer"."application_state";
CREATE TRIGGER application_state_trigger
    AFTER INSERT or UPDATE
    ON "pioneer"."application_state"
    FOR EACH ROW
    EXECUTE PROCEDURE "pioneer".fn_application_state_trigger();

-- pioneer application_state table
CREATE TABLE IF NOT EXISTS "pioneer"."application_state_history"
(
    "id"                    uuid    default uuid_generate_v4(),
    "application_id"        uuid,
    "application_state"     varchar(32),
    "date"                  timestamp with time zone,
    PRIMARY KEY ("id"),
    CONSTRAINT fk_application_state
    FOREIGN KEY("application_id")
    REFERENCES "pioneer"."application"("id")
    ON DELETE CASCADE
    );

-- indexes of application_state table
CREATE INDEX IF NOT EXISTS idx_application_state_history_id on "pioneer"."application_state_history" (id);
CREATE INDEX IF NOT EXISTS idx_application_state_history_application_id on "pioneer"."application_state_history" (application_id);

-- pioneer decision table
CREATE TABLE IF NOT EXISTS "pioneer"."decision"
(
    "id"                    uuid    default uuid_generate_v4(),
    "application_id"        uuid,
    "decision_status"       varchar(100),
    "rate_version_id"       integer,
    "decline_details"       jsonb,
    PRIMARY KEY ("id"),
    CONSTRAINT fk_decision
    FOREIGN KEY("application_id")
    REFERENCES "pioneer"."application"("id")
    ON DELETE CASCADE
    );

-- indexes of decision table
CREATE INDEX IF NOT EXISTS idx_decision_id on "pioneer"."decision" (id);
CREATE INDEX IF NOT EXISTS idx_decision_application_id on "pioneer"."decision" (application_id);

-- pioneer offer table
CREATE TABLE IF NOT EXISTS "pioneer"."offer"
(
    "id"                    uuid    default uuid_generate_v4(),
    "decision_id"           uuid,
    "term_months"           integer,
    "rate_type"             varchar(100),
    "apr"                   numeric,
    "dealer_fee"            numeric,
    "selected"              boolean not null default false,
    PRIMARY KEY ("id"),
    CONSTRAINT fk_offer
    FOREIGN KEY("decision_id")
    REFERENCES "pioneer"."decision"("id")
    ON DELETE CASCADE
    );

-- indexes of offer table
CREATE INDEX IF NOT EXISTS idx_offer_id on "pioneer"."offer" (id);
CREATE INDEX IF NOT EXISTS idx_offer_decision_id on "pioneer"."offer" (decision_id);

-- pioneer document table
CREATE TABLE IF NOT EXISTS "pioneer"."document"
(
    "id"                   uuid    default uuid_generate_v4(),
    "application_id"       uuid,
    "document_type"        varchar(256),
    "deleted"              boolean not null default false,
    PRIMARY KEY ("id"),
    CONSTRAINT fk_document
    FOREIGN KEY("application_id")
    REFERENCES "pioneer"."application"("id")
    ON DELETE CASCADE
    );

-- indexes of document table
CREATE INDEX IF NOT EXISTS idx_document_id on "pioneer"."document" (id);
CREATE INDEX IF NOT EXISTS idx_document_application_id on "pioneer"."document" (application_id);