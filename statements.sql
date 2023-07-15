CREATE TABLE public.bookstable
(
    id uuid PRIMARY KEY NOT NULL,
    name text   UNIQUE NOT NULL,
    price numeric(6,2),
    inventory integer
)