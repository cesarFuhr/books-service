ALTER TABLE public.bookstable
  ADD COLUMN IF NOT EXISTS archived BOOLEAN DEFAULT false;