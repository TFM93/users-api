CREATE OR REPLACE FUNCTION updated_at_refresh()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TABLE IF NOT EXISTS users(
   id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
   first_name VARCHAR(25) NOT NULL,
   last_name VARCHAR(25)  NOT NULL,
   country_iso_code VARCHAR(2),
   pw VARCHAR (300) NOT NULL,
   nickname VARCHAR (25) UNIQUE NOT NULL,
   email VARCHAR (320) UNIQUE NOT NULL,

   created_at TIMESTAMPTZ NOT NULL DEFAULT current_timestamp,
   updated_at TIMESTAMPTZ NOT NULL DEFAULT current_timestamp
);

CREATE INDEX idx_users_updated_at_id ON users (updated_at DESC, id DESC);

CREATE TRIGGER set_updated_at
BEFORE UPDATE ON users
FOR EACH ROW
EXECUTE FUNCTION updated_at_refresh();