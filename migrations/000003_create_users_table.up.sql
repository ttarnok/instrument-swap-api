CREATE TABLE IF NOT EXISTS users (
  id bigserial PRIMARY KEY,
  created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
  is_deleted boolean NOT NULL DEFAULT FALSE,
  deleted_at timestamp(0),
  name text NOT NULL,
  email citext UNIQUE NOT NULL,
  passwords_hash bytea NOT NULL,
  activated bool NOT NULL,
  version integer NOT NULL DEFAULT 1
);
