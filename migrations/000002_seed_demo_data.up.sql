INSERT INTO categories (id, name) VALUES
  (1, 'Programming'),
  (2, 'Fiction'),
  (3, 'Science'),
  (4, 'History');

INSERT INTO authors (id, full_name) VALUES
  (1, 'Alan A. A. Donovan'),
  (2, 'Brian W. Kernighan'),
  (3, 'Robert C. Martin'),
  (4, 'Andrew Hunt'),
  (5, 'David Thomas'),
  (6, 'Frank Herbert'),
  (7, 'Isaac Asimov'),
  (8, 'Stephen Hawking'),
  (9, 'Yuval Noah Harari');

INSERT INTO books (id, title, isbn, category_id, publisher, published_year, description) VALUES
  (1, 'The Go Programming Language', '9780134190440', 1, 'Addison-Wesley', 2015, 'A complete guide to Go.'),
  (2, 'Clean Code', '9780132350884', 1, 'Prentice Hall', 2008, 'A handbook of agile software craftsmanship.'),
  (3, 'The Pragmatic Programmer', '9780135957059', 1, 'Addison-Wesley', 2019, 'Your journey to mastery.'),
  (4, 'Dune', '9780441013593', 2, 'Ace', 2005, 'A science fiction classic.'),
  (5, 'Foundation', '9780553293357', 2, 'Bantam Spectra', 1991, 'The first novel of the Foundation series.'),
  (6, 'I, Robot', '9780553382563', 2, 'Bantam Spectra', 2004, 'A collection of robot stories.'),
  (7, 'A Brief History of Time', '9780553380163', 3, 'Bantam', 1998, 'From the Big Bang to black holes.'),
  (8, 'Sapiens', '9780062316110', 4, 'Harper', 2015, 'A brief history of humankind.');

INSERT INTO book_authors (book_id, author_id) VALUES
  (1, 1), (1, 2), (2, 3), (3, 4), (3, 5), (4, 6), (5, 7), (6, 7), (7, 8), (8, 9);
