ALTER TABLE quotes
    RENAME quote TO data;

ALTER TABLE quotes
    ADD COLUMN type VARCHAR(10) NOT NULL DEFAULT 'quote';