CREATE TABLE public.quotes (
    id SERIAL PRIMARY KEY ,
    data TEXT,
    type VARCHAR(10) NOT NULL DEFAULT 'quote'
);