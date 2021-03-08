-- migrate:up
CREATE TABLE users (
    user_id bigserial PRIMARY KEY,
    role_id bigint DEFAULT (1) NOT NULL,
    first_name varchar(255) NOT NULL,
    last_name varchar(255) NOT NULL,
    email varchar(255) UNIQUE NOT NULL,
    phone varchar(10) NOT NULL,
    password varchar(255) NOT NULL,
    active boolean NOT NULL DEFAULT FALSE,
    created_at timestamptz NOT NULL DEFAULT (now()),
    FOREIGN KEY (role_id) REFERENCES roles (role_id)
);

-- migrate:down
DROP TABLE users;

