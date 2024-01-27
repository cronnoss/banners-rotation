-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS impressions
(
    id           SERIAL CONSTRAINT impressions_pk PRIMARY KEY,
    slot_id      INT       NOT NULL CONSTRAINT impressions_slots_id_fk REFERENCES slots ON UPDATE CASCADE ON DELETE CASCADE,
    banner_id    INT       NOT NULL CONSTRAINT impressions_banners_id_fk REFERENCES banners ON UPDATE CASCADE ON DELETE CASCADE,
    usergroup_id INT       NOT NULL CONSTRAINT impressions_usergroups_id_fk REFERENCES usergroups ON UPDATE CASCADE ON DELETE CASCADE,
    created_at   TIMESTAMP NOT NULL
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS impressions;
-- +goose StatementEnd
