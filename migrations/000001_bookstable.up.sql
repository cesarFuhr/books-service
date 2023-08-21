/*
CREATE OR REPLACE FUNCTION trigger_set_timestamp()
RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at = NOW();
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;
*/

CREATE TABLE IF NOT EXISTS public.bookstable
(
    id uuid PRIMARY KEY NOT NULL,
    name text   UNIQUE NOT NULL,
    price numeric(6,2),
    inventory integer,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
)

/*
CREATE TRIGGER set_timestamp
BEFORE UPDATE ON public.bookstable
FOR EACH ROW
EXECUTE PROCEDURE trigger_set_timestamp();
*/