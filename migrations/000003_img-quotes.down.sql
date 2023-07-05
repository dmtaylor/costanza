ALTER TABLE quotes
    DROP COLUMN type;

ALTER TABLE quotes
    RENAME data TO quote;