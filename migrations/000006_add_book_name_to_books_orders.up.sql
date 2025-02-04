ALTER TABLE public.books_orders
  ADD COLUMN IF NOT EXISTS book_name text;