-- 1. ISBN becomes optional.
ALTER TABLE books DROP CONSTRAINT IF EXISTS books_isbn_length_check;

-- 2. Empty default for ISBN.
ALTER TABLE books ALTER COLUMN isbn SET DEFAULT '';

-- 3. ISBN is unique only when it is not empty.
CREATE UNIQUE INDEX books_isbn_unique ON books (isbn) WHERE isbn <> '';

-- 4. Compatibility columns.
ALTER TABLE books
    ADD COLUMN legacy_code INTEGER NULL,
    ADD COLUMN publication_city VARCHAR(100) NOT NULL DEFAULT '',
    ADD COLUMN page_count INTEGER NOT NULL DEFAULT 0 CHECK (page_count >= 0),
    ADD COLUMN bbk VARCHAR(30) NOT NULL DEFAULT '',
    ADD COLUMN udk VARCHAR(100) NOT NULL DEFAULT '',
    ADD COLUMN language VARCHAR(3) NOT NULL DEFAULT '' CHECK (language = '' OR language ~ '^[a-z]{3}$');

-- 5. Re-import key.
CREATE UNIQUE INDEX books_legacy_code_key ON books (legacy_code) WHERE legacy_code IS NOT NULL;

-- 6. BBK index.
CREATE INDEX books_bbk_idx ON books (bbk);

-- 7. Department of a copy.
ALTER TABLE copies ADD COLUMN department VARCHAR(20) NOT NULL DEFAULT '';

-- 8. Author order.
ALTER TABLE book_authors ADD COLUMN position SMALLINT NOT NULL DEFAULT 0;
