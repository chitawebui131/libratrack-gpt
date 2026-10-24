-- 1. DROP INDEX IF EXISTS copies_inventory_no_idx; DROP INDEX IF EXISTS books_bbk_idx; DROP INDEX IF EXISTS books_legacy_code_key;
DROP INDEX IF EXISTS copies_inventory_no_idx;
DROP INDEX IF EXISTS books_bbk_idx;
DROP INDEX IF EXISTS books_legacy_code_key;

-- 2. ALTER TABLE book_authors DROP COLUMN IF EXISTS position;
ALTER TABLE book_authors DROP COLUMN IF EXISTS position;

-- 3. ALTER TABLE copies DROP COLUMN IF EXISTS department;
ALTER TABLE copies DROP COLUMN IF EXISTS department;

-- 4. ALTER TABLE books DROP COLUMN IF EXISTS language, DROP COLUMN IF EXISTS udk, DROP COLUMN IF EXISTS bbk, DROP COLUMN IF EXISTS page_count, DROP COLUMN IF EXISTS publication_city, DROP COLUMN IF EXISTS legacy_code;
ALTER TABLE books DROP COLUMN IF EXISTS language;
ALTER TABLE books DROP COLUMN IF EXISTS udk;
ALTER TABLE books DROP COLUMN IF EXISTS bbk;
ALTER TABLE books DROP COLUMN IF EXISTS page_count;
ALTER TABLE books DROP COLUMN IF EXISTS publication_city;
ALTER TABLE books DROP COLUMN IF EXISTS legacy_code;

-- 5. DROP INDEX IF EXISTS books_isbn_unique;
DROP INDEX IF EXISTS books_isbn_unique;

-- 6. Add a comment line: the unique constraint on isbn is NOT restored because books with an empty isbn may exist; restore it by hand after cleaning the data.
-- the unique constraint on isbn is NOT restored because books with an empty isbn may exist; restore it by hand after cleaning the data.
