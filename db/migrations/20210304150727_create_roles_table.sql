-- migrate:up
CREATE TABLE roles (
    role_id bigserial PRIMARY KEY,
    name varchar(55) UNIQUE NOT NULL,
    created_at timestamptz NOT NULL DEFAULT (now())
);

INSERT INTO roles (role_id, name)
    VALUES (1, 'user'), (2, 'host'), (3, 'moderator'), (4, 'admin');

-- migrate:down
DROP TABLE roles;

