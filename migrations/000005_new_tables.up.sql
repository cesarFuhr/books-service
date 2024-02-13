CREATE TYPE access AS ENUM ('user', 'admin');

CREATE TABLE IF NOT EXISTS public.userstable
(
    user_id uuid PRIMARY KEY NOT NULL,
    name text   NOT NULL,
    user_access access DEFAULT 'user'
);

CREATE TYPE order_status AS ENUM ('accepting_items', 'canceled', 'waiting_payment', 'paid');

CREATE TABLE IF NOT EXISTS public.orderstable
(
order_id		uuid PRIMARY KEY NOT NULL,
purchaser_id 	uuid,	
order_status	order_status DEFAULT 'accepting_items',
created_at 		timestamp with time zone DEFAULT now(),
updated_at 	timestamp with time zone DEFAULT now()
);

CREATE TABLE IF NOT EXISTS public.books_orderstable (
    order_id		uuid REFERENCES public.orderstable ON DELETE CASCADE,
    book_id		uuid REFERENCES public.bookstable ON DELETE CASCADE,
    book_units integer,
    book_price_at_order	numeric(6,2),
    created_at 		timestamp with time zone DEFAULT now(),
    updated_at 	timestamp with time zone DEFAULT now(),
    PRIMARY KEY (order_id, book_id)
);

CREATE TABLE IF NOT EXISTS public.paymentstable (
    order_id		uuid REFERENCES public.orderstable ON DELETE CASCADE,
   order_paid          BOOLEAN DEFAULT false,
    created_at 		timestamp with time zone DEFAULT now(),
    updated_at 	timestamp with time zone DEFAULT now(),
    PRIMARY KEY (order_id)
);