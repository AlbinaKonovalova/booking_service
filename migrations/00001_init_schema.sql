-- +goose Up
-- +goose StatementBegin

CREATE TABLE resources (
    id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    name       TEXT        NOT NULL,
    created_at TIMESTAMPTZ NOT NULL    DEFAULT now(),
    removed_at TIMESTAMPTZ
);

CREATE INDEX idx_resources_removed_at ON resources (removed_at);

CREATE TABLE bookings (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    resource_id UUID        NOT NULL REFERENCES resources (id),
    start_time  TIMESTAMPTZ NOT NULL,
    end_time    TIMESTAMPTZ NOT NULL,
    check_in    TIMESTAMPTZ NOT NULL,
    check_out   TIMESTAMPTZ NOT NULL,
    status      TEXT        NOT NULL DEFAULT 'CREATED',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);


CREATE INDEX idx_bookings_resource_status     ON bookings (resource_id, status);
CREATE INDEX idx_bookings_status_start_time   ON bookings (status, start_time);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS bookings;
DROP TABLE IF EXISTS resources;

-- +goose StatementEnd