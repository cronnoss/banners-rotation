-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS rotations
(
    slot_id    INT       NOT NULL CONSTRAINT rotations_slots_id_fk REFERENCES slots ON UPDATE CASCADE ON DELETE CASCADE,
    banner_id  INT       NOT NULL CONSTRAINT rotations_banners_id_fk REFERENCES banners ON UPDATE CASCADE ON DELETE CASCADE,
    created_at TIMESTAMP NOT NULL,
    CONSTRAINT rotations_pk
    PRIMARY KEY (slot_id, banner_id)
);

DO
$$
    BEGIN
        IF (SELECT COUNT(*) FROM rotations) = 0 THEN
            -- Insert data only if the table is empty
            INSERT INTO rotations (slot_id, banner_id, created_at)
            VALUES (1, 1, NOW()),
                   (2, 2, NOW()),
                   (3, 3, NOW()),
                   (1, 4, NOW()),
                   (2, 5, NOW()),
                   (3, 6, NOW());
        END IF;
    END
$$;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS rotations;
-- +goose StatementEnd
