CREATE TABLE IF NOT EXISTS swaps (
  id bigserial PRIMARY KEY,
  created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
  is_deleted boolean NOT NULL DEFAULT FALSE,
  deleted_at timestamp(0),
  requester_instrument_id bigint NOT NULL REFERENCES instruments(id) ON DELETE RESTRICT,
  recipient_instrument_id bigint NOT NULL REFERENCES instruments(id) ON DELETE RESTRICT,
  requester_user_id bigint NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
  recipient_user_id bigint NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
  is_accepted boolean NOT NULL DEFAULT FALSE
);
