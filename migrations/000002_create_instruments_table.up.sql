CREATE TABLE IF NOT EXISTS instruments (
  id bigserial PRIMARY KEY,
  created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
  is_deleted boolean NOT NULL DEFAULT FALSE,
  deleted_at timestamp(0),
  name text NOT NULL,
  manufacturer text NOT NULL,
  manufacture_year integer NOT NULL,
  type text NOT NULL,
  estimated_value integer NOT NULL,
  condition text NOT NULL,
  description text,
  famous_owners text[],
  owner_user_id bigint NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
  is_swapped boolean NOT NULL DEFAULT FALSE,
  version integer NOT NULL DEFAULT 1
);
