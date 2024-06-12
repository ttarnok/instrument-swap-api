CREATE INDEX IF NOT EXISTS instruments_is_deleted_idx ON instruments USING HASH (is_deleted);
CREATE INDEX IF NOT EXISTS instruments_name_idx ON instruments USING GIN (to_tsvector('simple', name));
CREATE INDEX IF NOT EXISTS instruments_manufacturer_idx ON instruments (manufacturer);
CREATE INDEX IF NOT EXISTS instruments_type_idx ON instruments (type);
CREATE INDEX IF NOT EXISTS instruments_famous_owners_idx ON instruments USING GIN (famous_owners);
