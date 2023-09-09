ALTER TABLE public.bookstable
  ADD COLUMN IF NOT EXISTS archived boolean DEFAULT false;