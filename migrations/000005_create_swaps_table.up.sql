CREATE TABLE IF NOT EXISTS swaps (
  id bigserial PRIMARY KEY,
  created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
  requester_instrument_id bigint NOT NULL REFERENCES instruments(id) ON DELETE RESTRICT,
  recipient_instrument_id bigint NOT NULL REFERENCES instruments(id) ON DELETE RESTRICT,
  is_accepted boolean NOT NULL DEFAULT FALSE,
  accepted_at timestamp(0) with time zone,
  is_rejected boolean NOT NULL DEFAULT FALSE,
  rejected_at timestamp(0) with time zone,
  is_ended boolean NOT NULL DEFAULT FALSE,
  ended_at timestamp(0) with time zone,
  version integer NOT NULL DEFAULT 1
);
