-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS clicks
(
    id           SERIAL CONSTRAINT clicks_pk PRIMARY KEY,
    slot_id      INT       NOT NULL CONSTRAINT clicks_slots_id_fk REFERENCES slots ON UPDATE CASCADE ON DELETE CASCADE,
    banner_id    INT       NOT NULL CONSTRAINT clicks_banners_id_fk REFERENCES banners ON UPDATE CASCADE ON DELETE CASCADE,
    usergroup_id INT       NOT NULL CONSTRAINT clicks_usergroups_id_fk REFERENCES usergroups ON UPDATE CASCADE ON DELETE CASCADE,
    created_at   TIMESTAMP NOT NULL
);

DO
$$
    BEGIN
        IF (SELECT COUNT(*) FROM clicks) = 0 THEN
            -- Insert data only if the table is empty
            INSERT INTO clicks (slot_id, banner_id, usergroup_id, created_at)
            VALUES (1, 1, 1, NOW()),
                   (2, 2, 2, NOW()),
                   (3, 3, 3, NOW());
        END IF;
    END
$$;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS clicks;
-- +goose StatementEnd
