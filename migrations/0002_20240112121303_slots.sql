-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS slots
(
    id         SERIAL CONSTRAINT slots_pk PRIMARY KEY,
    name       VARCHAR NOT NULL,
    created_at TIMESTAMP NOT NULL
);

DO
$$
    BEGIN
        IF (SELECT COUNT(*) FROM slots) = 0 THEN
            -- Insert data only if the table is empty
            INSERT INTO slots (name, created_at)
            VALUES ('Slot 1', NOW()),
                   ('Slot 2', NOW()),
                   ('Slot 3', NOW());
        END IF;
    END
$$;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS slots;
-- +goose StatementEnd
