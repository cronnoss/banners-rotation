-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS usergroups
(
    id         SERIAL CONSTRAINT usergroups_pk PRIMARY KEY,
    name       VARCHAR   NOT NULL,
    created_at TIMESTAMP NOT NULL
);

DO
$$
    BEGIN
        IF (SELECT COUNT(*) FROM usergroups) = 0 THEN
            -- Insert data only if the table is empty
            INSERT INTO usergroups (name, created_at)
            VALUES ('Groups 1', NOW()),
                   ('Groups 2', NOW()),
                   ('Groups 3', NOW()),
                   ('Groups 4', NOW()),
                   ('Groups 5', NOW());
        END IF;
    END
$$;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS usergroups;
-- +goose StatementEnd
