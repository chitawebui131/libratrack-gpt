## ТЕМА 8 (версія для малої моделі). Database Integration: PostgreSQL + GORM

Мета блоку та сама, що в `tema8.md`: PostgreSQL + GORM замість in-memory сховища, зв'язки Author/Category/Copy, міграції golang-migrate, ендпоінт доступності. Різниця лише в нарізці: **один крок = один файл** (максимум два дрібні), бо `qwen2.5-coder:7b` у режимі `whole` ламається на кроках, де треба створити чи змінити 5 файлів одразу (див. невдалий крок 8.2).

Нумерація кроків суцільна (8.1 ... 8.43) без літер, щоб не ламати розбір заголовків у скрипті-запускачі. Формат кожного кроку (заголовок, один блок коду з промптом, блок «Перевір сам») збережено.

**Що ми зрозуміли з невдалого запуску і як це враховано:**
- Модель сама додає зайві імпорти (`strings`, `gin/binding`, `validator`) і ламає збірку. У кожному промпті тепер явний список дозволених імпортів і заборона на все інше.
- У режимі `whole` модель відповіла порожніми `diff`-блоками. У кожному промпті є рядок «COMPLETE content, whole file, never a diff».
- Повтори «виправлення» погіршували стан. Правило: **не більше 1 повтору**; якщо крок не вдався, натисни `s`, зроби `git reset --hard HEAD~1` і запусти крок ще раз (або винеси складну частину в окремий крок).
- Складні місця (транзакції GORM, `Association`, `pgconn`, `clause.Associations`) розписані у промптах майже покроково, бо там 7B вигадує неіснуючі API.
- `Summarization failed ... cannot schedule new futures after shutdown` ігноруємо: це збій стиснення історії чату, а не помилка коду.

**Рекомендація: допиши в `CONVENTIONS.md` (його вже читає Aider) такі правила:**
- Output the complete content of every file (whole file); never output a diff.
- Do not add imports that are not used; do not import gin, binding or validator in internal/model.
- Do not create files that were not requested; do not rename existing functions or types.
- Use only ASCII characters in .go files.
- If a requested file already exists, keep everything not mentioned in the task unchanged.

**Про збірку (`go build ./...` після кожного кроку):**
- Кроки 8.1-8.14 залишають проєкт компільованим.
- Кроки **8.15-8.22 — «вікно», де збірка падає** (інтерфейс BookRepository оголошено раніше, ніж реалізовано всі його методи, а старий main.go ще посилається на видалений репозиторій). Це очікувано: у вікні натискай Enter. Після кроку **8.22** збірка знову має бути зеленою.
- Після кроків 8.24-8.31 збірка також має проходити, окрім випадків, зазначених у «Перевір сам».

**Перед стартом (робить студент, не агент):**
- Проєкт з теми 7 збирається, тести проходять, закомічено стан теми 7 (`git tag theme-7`).
- Docker Desktop (або docker + docker compose) встановлений.
- Залежності ставимо вручну, не через Aider (він не вміє `go get`): `go get gorm.io/gorm gorm.io/driver/postgres github.com/jackc/pgx/v5 github.com/golang-migrate/migrate/v4 github.com/golang-migrate/migrate/v4/database/postgres github.com/golang-migrate/migrate/v4/source/file`. Не запускай `go mod tidy` до кінця теми (він видалить пакети, які ще ніде не імпортуються).

**Рішення студента перед запуском (за бажанням, значення за замовчуванням уже в промптах):** назва БД і користувача (`libratrack`), порт 5432, дані для seed (кроки 8.32-8.35), розмір сторінки за замовчуванням (20, максимум 100).

---

### 8.1 — docker-compose.yml для PostgreSQL

```
RULES: Output the COMPLETE content of the file below (whole file, never a diff, no explanations).

Create docker-compose.yml in the project root with ONE service named "db":
- image: postgres:16-alpine
- container_name: libratrack-db
- environment: POSTGRES_USER=libratrack, POSTGRES_PASSWORD=libratrack, POSTGRES_DB=libratrack
- ports: "5432:5432"
- volumes: libratrack_pgdata:/var/lib/postgresql/data
- healthcheck: test ["CMD-SHELL", "pg_isready -U libratrack -d libratrack"], interval 5s, timeout 3s, retries 10
Add a top-level "volumes:" section declaring libratrack_pgdata.
Do not create any other files.
```

**Перевір сам:**
`docker compose up -d`, потім `docker compose ps`: сервіс `db` у стані healthy.

---

### 8.2 — .env.example

```
RULES: Output the COMPLETE content of the file below (whole file, never a diff, no explanations).

Create .env.example in the project root with exactly these four lines and nothing else:
PORT=8080
ALLOWED_ORIGIN=*
DATABASE_URL=postgres://libratrack:libratrack@localhost:5432/libratrack?sslmode=disable
MIGRATIONS_PATH=migrations
```

**Перевір сам:**
Скопіюй у `.env`: `cp .env.example .env`. Змінні експортуй перед запуском застосунку: `export $(grep -v '^#' .env | xargs)` (на Windows задай їх у PowerShell через `$env:ІМ'Я="..."`).

---

### 8.3 — .gitignore

```
RULES: Output the COMPLETE content of the file below (whole file, never a diff, no explanations).

Create .gitignore in the project root with exactly one entry per line:
.env
*.exe
*.out
bin/
.idea/
.vscode/
.aider*
check-output/
all-results/
```

**Перевір сам:**
Файл `.env` не відстежується git (`git status`). Перед наступним кроком переконайся, що виконав `go get ...` з розділу «Перед стартом».

---

### 8.4 — Модель Author

```
RULES: Output the COMPLETE content of the file below (whole file, never a diff, no explanations). The ONLY allowed import is "time". Do NOT import gin, binding or validator: validation is done only with struct tags.

Create internal/model/author.go (package model) containing exactly two types and nothing else.

1. type Author struct, one field per line:
   ID        uint      `gorm:"primaryKey" json:"id"`
   FullName  string    `gorm:"size:255;not null" json:"full_name"`
   Bio       string    `gorm:"type:text;not null;default:''" json:"bio"`
   CreatedAt time.Time `json:"created_at"`
   UpdatedAt time.Time `json:"updated_at"`

2. type CreateAuthorRequest struct:
   FullName string `json:"full_name" binding:"required,max=255"`
   Bio      string `json:"bio" binding:"max=2000"`

No functions, no extra comments.
```

**Перевір сам:**
`go build ./internal/model/...` проходить. У файлі лише імпорт `"time"`.

---

### 8.5 — Модель Category

```
RULES: Output the COMPLETE content of the file below (whole file, never a diff, no explanations). The ONLY allowed import is "time". Do NOT import gin, binding or validator.

Create internal/model/category.go (package model) containing exactly two types and nothing else.

1. type Category struct, one field per line:
   ID        uint      `gorm:"primaryKey" json:"id"`
   Name      string    `gorm:"size:100;not null;uniqueIndex" json:"name"`
   CreatedAt time.Time `json:"created_at"`
   UpdatedAt time.Time `json:"updated_at"`

2. type CreateCategoryRequest struct:
   Name string `json:"name" binding:"required,max=100"`

No functions, no extra comments.
```

**Перевір сам:**
`go build ./internal/model/...` проходить.

---

### 8.6 — Модель Copy і BookAvailability

```
RULES: Output the COMPLETE content of the file below (whole file, never a diff, no explanations). The ONLY allowed import is "time". Do NOT import gin, binding or validator.

Create internal/model/copy.go (package model) with exactly these declarations and nothing else.

1. type CopyStatus string
2. A const block:
   CopyStatusAvailable   CopyStatus = "available"
   CopyStatusBorrowed    CopyStatus = "borrowed"
   CopyStatusLost        CopyStatus = "lost"
   CopyStatusUnderRepair CopyStatus = "under_repair"
3. type Copy struct, one field per line:
   ID          uint       `gorm:"primaryKey" json:"id"`
   BookID      uint       `gorm:"not null;index" json:"book_id"`
   InventoryNo string     `gorm:"size:50;not null;uniqueIndex" json:"inventory_no"`
   Status      CopyStatus `gorm:"size:20;not null;default:available" json:"status"`
   CreatedAt   time.Time  `json:"created_at"`
   UpdatedAt   time.Time  `json:"updated_at"`
4. type CreateCopyRequest struct:
   InventoryNo string     `json:"inventory_no" binding:"required,max=50"`
   Status      CopyStatus `json:"status" binding:"omitempty,oneof=available borrowed lost under_repair"`
5. type BookAvailability struct:
   BookID          uint `json:"book_id"`
   TotalCopies     int  `json:"total_copies"`
   AvailableCopies int  `json:"available_copies"`
```

**Перевір сам:**
`go build ./internal/model/...` проходить; у файлі немає імпортів, крім `"time"`.

---

### 8.7 — Модель User (лише для схеми)

```
RULES: Output the COMPLETE content of the file below (whole file, never a diff, no explanations). The ONLY allowed import is "time".

Create internal/model/user.go (package model) with exactly one type.
Above the type put this one-line comment: // User exists only for the schema and foreign keys in theme 8; authentication is implemented in theme 9.

type User struct, one field per line:
   ID           uint      `gorm:"primaryKey" json:"id"`
   Email        string    `gorm:"size:255;not null;uniqueIndex" json:"email"`
   PasswordHash string    `gorm:"size:255;not null;default:''" json:"-"`
   FullName     string    `gorm:"size:255;not null;default:''" json:"full_name"`
   Role         string    `gorm:"size:20;not null;default:reader" json:"role"`
   CreatedAt    time.Time `json:"created_at"`

No functions, no other types.
```

**Перевір сам:**
`go build ./internal/model/...` проходить. Усі чотири нові моделі (author, category, copy, user) готові; `book.go` змінимо пізніше (крок 8.14).

---

### 8.8 — Міграція 000001, частина 1: categories, authors, books

```
RULES: Output the COMPLETE content of the file below (whole file, never a diff, no explanations). This is a plain SQL file (PostgreSQL), no Go code.

Create migrations/000001_init_schema.up.sql with exactly these three tables, in this order, one column per line:

CREATE TABLE categories (
  id BIGSERIAL PRIMARY KEY,
  name VARCHAR(100) NOT NULL UNIQUE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE authors (
  id BIGSERIAL PRIMARY KEY,
  full_name VARCHAR(255) NOT NULL,
  bio TEXT NOT NULL DEFAULT '',
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE books (
  id BIGSERIAL PRIMARY KEY,
  title VARCHAR(255) NOT NULL,
  isbn VARCHAR(20) NOT NULL UNIQUE,
  category_id BIGINT NOT NULL REFERENCES categories(id) ON DELETE RESTRICT,
  publisher VARCHAR(255) NOT NULL DEFAULT '',
  published_year INT NOT NULL DEFAULT 0,
  description TEXT NOT NULL DEFAULT '',
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

Do not add anything else to the file.
```

**Перевір сам:**
Файл містить рівно три `CREATE TABLE`. Далі (крок 8.9) цей же файл розширюємо.

---

### 8.9 — Міграція 000001, частина 2: book_authors, copies, users, індекси

```
RULES: Output the COMPLETE content of the file migrations/000001_init_schema.up.sql (whole file, never a diff, no explanations). Keep the three existing tables (categories, authors, books) EXACTLY as they are, and append the following at the end of the file, in this order:

CREATE TABLE book_authors (
  book_id BIGINT NOT NULL REFERENCES books(id) ON DELETE CASCADE,
  author_id BIGINT NOT NULL REFERENCES authors(id) ON DELETE RESTRICT,
  PRIMARY KEY (book_id, author_id)
);

CREATE TABLE copies (
  id BIGSERIAL PRIMARY KEY,
  book_id BIGINT NOT NULL REFERENCES books(id) ON DELETE CASCADE,
  inventory_no VARCHAR(50) NOT NULL UNIQUE,
  status VARCHAR(20) NOT NULL DEFAULT 'available' CHECK (status IN ('available','borrowed','lost','under_repair')),
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE users (
  id BIGSERIAL PRIMARY KEY,
  email VARCHAR(255) NOT NULL UNIQUE,
  password_hash VARCHAR(255) NOT NULL DEFAULT '',
  full_name VARCHAR(255) NOT NULL DEFAULT '',
  role VARCHAR(20) NOT NULL DEFAULT 'reader' CHECK (role IN ('admin','librarian','reader')),
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_books_category_id ON books(category_id);
CREATE INDEX idx_book_authors_author_id ON book_authors(author_id);
CREATE INDEX idx_copies_book_id ON copies(book_id);
CREATE INDEX idx_copies_status ON copies(status);
```

**Перевір сам:**
Файл має 6 таблиць і 4 індекси, перші три таблиці не змінились.

---

### 8.10 — Міграція 000001: down

```
RULES: Output the COMPLETE content of the file below (whole file, never a diff, no explanations). Plain SQL.

Create migrations/000001_init_schema.down.sql with exactly these lines, in this order:
DROP TABLE IF EXISTS copies;
DROP TABLE IF EXISTS book_authors;
DROP TABLE IF EXISTS books;
DROP TABLE IF EXISTS authors;
DROP TABLE IF EXISTS categories;
DROP TABLE IF EXISTS users;
```

**Перевір сам:**
Порядок видалення враховує зовнішні ключі (спочатку залежні таблиці).

---

### 8.11 — Запуск міграцій (migrate.go)

```
RULES: Output the COMPLETE content of the file below (whole file, never a diff, no explanations). Allowed imports ONLY: "errors", "github.com/golang-migrate/migrate/v4", and two blank imports: _ "github.com/golang-migrate/migrate/v4/database/postgres" and _ "github.com/golang-migrate/migrate/v4/source/file". Nothing else.

Create internal/repository/migrate.go (package repository) with one function:

func RunMigrations(databaseURL, migrationsPath string) error

Body, exactly this logic:
1. m, err := migrate.New("file://"+migrationsPath, databaseURL); if err != nil return err
2. defer m.Close()
3. if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) { return err }
4. return nil

No other functions, no logging.
```

**Перевір сам:**
`go build ./internal/repository/...` проходить (старий in-memory репозиторій ще на місці). Якщо збірка скаржиться на відсутній пакет, запусти `go get` з розділу «Перед стартом».

---

### 8.12 — Підключення до PostgreSQL (db.go)

```
RULES: Output the COMPLETE content of the file below (whole file, never a diff, no explanations). Allowed imports ONLY: "time", "gorm.io/driver/postgres", "gorm.io/gorm", "gorm.io/gorm/logger". Nothing else.

Create internal/repository/db.go (package repository) with one function:

func NewPostgresDB(databaseURL string) (*gorm.DB, error)

Logic, exactly:
1. db, err := gorm.Open(postgres.Open(databaseURL), &gorm.Config{Logger: logger.Default.LogMode(logger.Info)}); on error return nil, err
2. sqlDB, err := db.DB(); on error return nil, err
3. sqlDB.SetMaxOpenConns(10); sqlDB.SetMaxIdleConns(5); sqlDB.SetConnMaxLifetime(30 * time.Minute)
4. return db, nil

Do NOT call AutoMigrate anywhere: the schema is managed only by golang-migrate.
```

**Перевір сам:**
`go build ./internal/repository/...` проходить. Нема жодного `AutoMigrate`.
**Перед наступним кроком виконай вручну (не Aider):** видали старий код теми 7, який більше не потрібен: `git rm internal/repository/book_repository.go internal/repository/book_repository_test.go internal/handler/book_handler.go` і всі `*_test.go` у `internal/handler/`, якщо вони є (`git rm internal/handler/*_test.go`). Старий `main.go` залишається. Після цього збірка падатиме, поки не завершиться крок 8.22. Це очікувано.

---

### 8.13 — Спільні помилки репозиторію (errors.go)

```
RULES: Output the COMPLETE content of the file below (whole file, never a diff, no explanations). Allowed imports ONLY: "errors", "github.com/jackc/pgx/v5/pgconn", "gorm.io/gorm". Nothing else.

Create internal/repository/errors.go (package repository).

1. Package-level variables:
   var ErrNotFound = errors.New("not found")
   var ErrConflict = errors.New("conflict")
   var ErrInvalidReference = errors.New("invalid reference")

2. func translateError(err error) error with this exact logic:
   - if err == nil return nil
   - if errors.Is(err, gorm.ErrRecordNotFound) return ErrNotFound
   - declare: var pgErr *pgconn.PgError
   - if errors.As(err, &pgErr) then: if pgErr.Code == "23505" return ErrConflict; if pgErr.Code == "23503" return ErrInvalidReference
   - otherwise return err unchanged
```

**Перевір сам:**
`ErrNotFound` оголошений лише тут (старий файл репозиторію видалено на попередньому кроці).

---

### 8.14 — Нова модель Book

```
RULES: Output the COMPLETE content of the file below (whole file, never a diff, no explanations). The ONLY allowed import is "time". Do NOT import gin, binding or validator. This file REPLACES the old internal/model/book.go completely: remove the old plain-text Author and Category string fields and the old comment about them.

Create internal/model/book.go (package model) with exactly these three types.

1. type Book struct, one field per line:
   ID            uint      `gorm:"primaryKey" json:"id"`
   Title         string    `gorm:"size:255;not null" json:"title"`
   ISBN          string    `gorm:"size:20;not null;uniqueIndex" json:"isbn"`
   CategoryID    uint      `gorm:"not null;index" json:"category_id"`
   Category      *Category `gorm:"foreignKey:CategoryID" json:"category,omitempty"`
   Publisher     string    `gorm:"size:255;not null;default:''" json:"publisher"`
   PublishedYear int       `gorm:"not null;default:0" json:"published_year"`
   Description   string    `gorm:"type:text;not null;default:''" json:"description"`
   Authors       []Author  `gorm:"many2many:book_authors" json:"authors,omitempty"`
   Copies        []Copy    `gorm:"foreignKey:BookID" json:"copies,omitempty"`
   CreatedAt     time.Time `json:"created_at"`
   UpdatedAt     time.Time `json:"updated_at"`

2. type CreateBookRequest struct, one field per line:
   Title         string `json:"title" binding:"required,max=255"`
   ISBN          string `json:"isbn" binding:"required,min=10,max=20"`
   CategoryID    uint   `json:"category_id" binding:"required,gt=0"`
   AuthorIDs     []uint `json:"author_ids" binding:"required,min=1,dive,gt=0"`
   Publisher     string `json:"publisher" binding:"max=255"`
   PublishedYear int    `json:"published_year" binding:"gte=0,lte=2100"`
   Description   string `json:"description" binding:"max=5000"`

3. type UpdateBookRequest struct with exactly the same fields and tags as CreateBookRequest.
```

**Перевір сам:**
У `Book` немає полів `Author string` і `Category string`. `go build ./internal/model/...` проходить; решта проєкту поки не збирається.

---

### 8.15 — Репозиторій книг: інтерфейс, Create, FindByID

```
RULES: Output the COMPLETE content of the file below (whole file, never a diff, no explanations). Allowed imports ONLY: "gorm.io/gorm", "gorm.io/gorm/clause", "libratrack/internal/model". Nothing else. Methods List, Update and Delete are implemented in OTHER files, do NOT implement them here.

Create internal/repository/book_repository.go (package repository).

1. The interface:
   type BookRepository interface {
       Create(book *model.Book, authorIDs []uint) error
       FindByID(id uint) (*model.Book, error)
       List(categoryID uint, page, limit int) ([]model.Book, error)
       Update(book *model.Book, authorIDs []uint) error
       Delete(id uint) error
   }

2. type gormBookRepository struct { db *gorm.DB }
   func NewBookRepository(db *gorm.DB) BookRepository { return &gormBookRepository{db: db} }

3. Helper func uniqueIDs(ids []uint) []uint: returns the IDs without duplicates, keeping the original order.

4. Helper func loadAuthors(tx *gorm.DB, ids []uint) ([]model.Author, error):
   - ids = uniqueIDs(ids)
   - var authors []model.Author
   - if err := tx.Where("id IN ?", ids).Find(&authors).Error; err != nil return nil, err
   - if len(authors) != len(ids) return nil, ErrInvalidReference (some author does not exist)
   - return authors, nil

5. Method (r *gormBookRepository) Create(book *model.Book, authorIDs []uint) error:
   err := r.db.Transaction(func(tx *gorm.DB) error {
       authors, err := loadAuthors(tx, authorIDs)
       if err != nil { return err }
       book.Authors = nil
       book.Category = nil
       book.Copies = nil
       if err := tx.Omit(clause.Associations).Create(book).Error; err != nil { return err }
       return tx.Model(book).Association("Authors").Replace(authors)
   })
   if err != nil { return translateError(err) }
   then reload: var fresh model.Book; if err := r.db.Preload("Authors").Preload("Category").First(&fresh, book.ID).Error; err != nil { return translateError(err) }; *book = fresh; return nil

6. Method (r *gormBookRepository) FindByID(id uint) (*model.Book, error):
   var b model.Book
   if err := r.db.Preload("Authors").Preload("Category").Preload("Copies").First(&b, id).Error; err != nil { return nil, translateError(err) }
   return &b, nil
```

**Перевір сам:**
Збірка тут падає (методи `List`, `Update`, `Delete` ще не реалізовані): це очікувано. `Omit(clause.Associations)` є в Create: без нього GORM може тихо створити «порожнього» автора.

---

### 8.16 — Репозиторій книг: List (спочатку наївний, з N+1)

```
RULES: Output the COMPLETE content of the file below (whole file, never a diff, no explanations). The ONLY allowed import is "libratrack/internal/model". Nothing else. This is intentionally a NAIVE implementation for a teaching demo: do not optimize it.

Create internal/repository/book_repository_list.go (package repository) with one method:

func (r *gormBookRepository) List(categoryID uint, page, limit int) ([]model.Book, error)

Logic, exactly:
1. if page < 1 { page = 1 }; if limit < 1 { limit = 20 }; if limit > 100 { limit = 100 }
2. q := r.db.Order("id ASC").Offset((page - 1) * limit).Limit(limit)
3. if categoryID > 0 { q = q.Where("category_id = ?", categoryID) }
4. var books []model.Book; if err := q.Find(&books).Error; err != nil { return nil, translateError(err) }
5. Put this comment line directly above the loop: // NAIVE: N+1 queries, fixed in step 8.23
   Then loop with an index: for i := range books { ... }
   Inside the loop:
   a) if err := r.db.Model(&books[i]).Association("Authors").Find(&books[i].Authors); err != nil { return nil, translateError(err) }
   b) var cat model.Category; if err := r.db.First(&cat, books[i].CategoryID).Error; err == nil { books[i].Category = &cat }
6. if books == nil { books = []model.Book{} }; return books, nil
```

**Перевір сам:**
Над циклом є коментар `NAIVE`. Збірка ще падає (немає `Update` і `Delete`).

---

### 8.17 — Репозиторій книг: Update і Delete

```
RULES: Output the COMPLETE content of the file below (whole file, never a diff, no explanations). Allowed imports ONLY: "gorm.io/gorm", "libratrack/internal/model". Nothing else.

Create internal/repository/book_repository_update.go (package repository) with two methods.

1. func (r *gormBookRepository) Update(book *model.Book, authorIDs []uint) error
   err := r.db.Transaction(func(tx *gorm.DB) error {
       authors, err := loadAuthors(tx, authorIDs)
       if err != nil { return err }
       res := tx.Model(&model.Book{}).Where("id = ?", book.ID).Updates(map[string]any{
           "title": book.Title, "isbn": book.ISBN, "category_id": book.CategoryID,
           "publisher": book.Publisher, "published_year": book.PublishedYear, "description": book.Description,
       })
       if res.Error != nil { return res.Error }
       if res.RowsAffected == 0 { return ErrNotFound }
       var current model.Book
       if err := tx.First(&current, book.ID).Error; err != nil { return err }
       return tx.Model(&current).Association("Authors").Replace(authors)
   })
   if err != nil { return translateError(err) }
   then reload exactly like in Create: var fresh model.Book; r.db.Preload("Authors").Preload("Category").First(&fresh, book.ID); on error return translateError(err); *book = fresh; return nil

2. func (r *gormBookRepository) Delete(id uint) error
   res := r.db.Delete(&model.Book{}, id)
   if res.Error != nil { return translateError(res.Error) }
   if res.RowsAffected == 0 { return ErrNotFound }
   return nil
```

**Перевір сам:**
Тепер `go build ./internal/repository/...` має пройти (всі 5 методів інтерфейсу реалізовані). Збірка всього проєкту ще падає через хендлер і `main.go`.

---

### 8.18 — Спільні помилки й хелпери хендлерів (handler/errors.go)

```
RULES: Output the COMPLETE content of the file below (whole file, never a diff, no explanations). Allowed imports ONLY: "errors", "net/http", "strconv", "github.com/gin-gonic/gin", "libratrack/internal/repository". Nothing else.

Create internal/handler/errors.go (package handler).

1. The error envelope is {"error": {"code": "...", "message": "...", "details": "..."}}.
   func respondError(c *gin.Context, status int, code, message, details string)
   Body: c.JSON(status, gin.H{"error": gin.H{"code": code, "message": message, "details": details}})

2. func respondValidation(c *gin.Context, err error)
   Calls respondError(c, http.StatusUnprocessableEntity, "validation_error", "request validation failed", err.Error())

3. func respondRepoError(c *gin.Context, err error): maps the repository errors with errors.Is:
   - repository.ErrNotFound -> 404, code "not_found", message "resource not found"
   - repository.ErrConflict -> 409, code "conflict", message "resource already exists"
   - repository.ErrInvalidReference -> 422, code "invalid_reference", message "referenced resource does not exist"
   - anything else -> 500, code "internal_error", message "internal server error" (details must be an empty string for the 500 case)

4. func parseID(c *gin.Context) (uint, bool):
   v, err := strconv.ParseUint(c.Param("id"), 10, 32)
   if err != nil || v == 0 { respondError(c, http.StatusBadRequest, "invalid_id", "invalid id", ""); return 0, false }
   return uint(v), true
```

**Перевір сам:**
Формат помилки такий самий, як у темі 7. Ці хелпери повторно використовуються в усіх хендлерах теми 8, а в темах 9-10 вони згадані як «наявні».

---

### 8.19 — BookHandler: структура, Create, GetByID

```
RULES: Output the COMPLETE content of the file below (whole file, never a diff, no explanations). Allowed imports ONLY: "net/http", "github.com/gin-gonic/gin", "libratrack/internal/model", "libratrack/internal/repository". Nothing else. Methods List, Update and Delete are implemented in OTHER files, do NOT write them here.

Create internal/handler/book_handler.go (package handler).

1. type BookHandler struct { repo repository.BookRepository }
   func NewBookHandler(repo repository.BookRepository) *BookHandler { return &BookHandler{repo: repo} }

2. func (h *BookHandler) Create(c *gin.Context):
   - var req model.CreateBookRequest; if err := c.ShouldBindJSON(&req); err != nil { respondValidation(c, err); return }
   - book := model.Book{Title: req.Title, ISBN: req.ISBN, CategoryID: req.CategoryID, Publisher: req.Publisher, PublishedYear: req.PublishedYear, Description: req.Description}
   - if err := h.repo.Create(&book, req.AuthorIDs); err != nil { respondRepoError(c, err); return }
   - c.JSON(http.StatusCreated, book)

3. func (h *BookHandler) GetByID(c *gin.Context):
   - id, ok := parseID(c); if !ok { return }
   - book, err := h.repo.FindByID(id); if err != nil { respondRepoError(c, err); return }
   - c.JSON(http.StatusOK, book)
```

**Перевір сам:**
Методи названо саме `Create` і `GetByID`. Збірка ще падає (немає `List`, `Update`, `Delete` хендлера, старий `main.go`).

---

### 8.20 — BookHandler: List

```
RULES: Output the COMPLETE content of the file below (whole file, never a diff, no explanations). Allowed imports ONLY: "net/http", "strconv", "github.com/gin-gonic/gin". Nothing else.

Create internal/handler/book_handler_list.go (package handler).

1. Helper func queryInt(c *gin.Context, name string, def int) (int, bool):
   - s := c.Query(name); if s == "" return def, true
   - v, err := strconv.Atoi(s); if err != nil || v < 0 { respondError(c, http.StatusUnprocessableEntity, "validation_error", "invalid query parameter", name); return 0, false }
   - return v, true

2. func (h *BookHandler) List(c *gin.Context):
   - categoryID, ok := queryInt(c, "category_id", 0); if !ok { return }
   - page, ok := queryInt(c, "page", 1); if !ok { return }
   - limit, ok := queryInt(c, "limit", 20); if !ok { return }
   - books, err := h.repo.List(uint(categoryID), page, limit); if err != nil { respondRepoError(c, err); return }
   - c.JSON(http.StatusOK, books)
```

**Перевір сам:**
Параметр називається `category_id` (раніше був текстовий `category`).

---

### 8.21 — BookHandler: Update і Delete

```
RULES: Output the COMPLETE content of the file below (whole file, never a diff, no explanations). Allowed imports ONLY: "net/http", "github.com/gin-gonic/gin", "libratrack/internal/model". Nothing else.

Create internal/handler/book_handler_update.go (package handler).

1. func (h *BookHandler) Update(c *gin.Context):
   - id, ok := parseID(c); if !ok { return }
   - var req model.UpdateBookRequest; if err := c.ShouldBindJSON(&req); err != nil { respondValidation(c, err); return }
   - book := model.Book{ID: id, Title: req.Title, ISBN: req.ISBN, CategoryID: req.CategoryID, Publisher: req.Publisher, PublishedYear: req.PublishedYear, Description: req.Description}
   - if err := h.repo.Update(&book, req.AuthorIDs); err != nil { respondRepoError(c, err); return }
   - c.JSON(http.StatusOK, book)

2. func (h *BookHandler) Delete(c *gin.Context):
   - id, ok := parseID(c); if !ok { return }
   - if err := h.repo.Delete(id); err != nil { respondRepoError(c, err); return }
   - c.Status(http.StatusNoContent)
```

**Перевір сам:**
Пакет `handler` тепер має `Create`, `GetByID`, `List`, `Update`, `Delete`. Збірка пакета `./internal/handler/...` має пройти; ще падає `cmd/api` (крок 8.22).

---

### 8.22 — main.go: БД, міграції, маршрути книг

```
RULES: Output the COMPLETE content of the file below (whole file, never a diff, no explanations). Keep ALL existing imports that are still used, the swag package-level annotations (@title, @version, @description, @BasePath), the Swagger route, the middleware registration order (Recovery, Logging, CORS) and the GET /health endpoint EXACTLY as they are in the current file. Additional imports allowed: "log", "os", "libratrack/internal/handler", "libratrack/internal/repository".

Rewrite cmd/api/main.go so that it:
1. Defines a helper func getEnv(key, fallback string) string returning the fallback when the variable is empty.
2. Reads PORT (default "8080"), MIGRATIONS_PATH (default "migrations") and DATABASE_URL. If DATABASE_URL is empty: log.Fatal("DATABASE_URL is required").
3. Calls repository.RunMigrations(databaseURL, migrationsPath) and then repository.NewPostgresDB(databaseURL); on any error calls log.Fatalf with a clear message.
4. Creates bookRepo := repository.NewBookRepository(db) and bookHandler := handler.NewBookHandler(bookRepo). Remove the old in-memory repository creation.
5. Registers the book routes under the group /api/v1 exactly like before, using these handler methods:
   POST /books -> bookHandler.Create
   GET /books -> bookHandler.List
   GET /books/:id -> bookHandler.GetByID
   PUT /books/:id -> bookHandler.Update
   DELETE /books/:id -> bookHandler.Delete
6. Starts the server on ":" + port.
Do not add any other routes in this step.
```

**Перевір сам:**
Тут збірка знову має бути **зеленою**: `go build ./...`. Запуск: `docker compose up -d`, експортуй `.env`, `go run ./cmd/api`. Створи категорію й автора в БД вручну для швидкої перевірки: `docker exec -it libratrack-db psql -U libratrack -c "INSERT INTO categories(name) VALUES ('Programming'); INSERT INTO authors(full_name) VALUES ('Test Author');"`, потім `curl -X POST localhost:8080/api/v1/books -H "Content-Type: application/json" -d '{"title":"Go Book","isbn":"9781234567890","category_id":1,"author_ids":[1]}'` (очікуєш 201 з автором і категорією у відповіді).

---

### 8.23 — Виправлення N+1 (List з Preload)

```
RULES: Output the COMPLETE content of the file internal/repository/book_repository_list.go (whole file, never a diff, no explanations). The ONLY allowed import is "libratrack/internal/model". Nothing else.

Rewrite the method List so that it no longer has any per-book queries. Keep the signature:
func (r *gormBookRepository) List(categoryID uint, page, limit int) ([]model.Book, error)

Logic, exactly:
1. if page < 1 { page = 1 }; if limit < 1 { limit = 20 }; if limit > 100 { limit = 100 }
2. Put this comment above the query: // Preload replaces 1+2N queries with 3 queries (books, authors, categories) regardless of the number of books.
3. q := r.db.Preload("Authors").Preload("Category").Order("id ASC").Offset((page - 1) * limit).Limit(limit)
4. if categoryID > 0 { q = q.Where("category_id = ?", categoryID) }
5. var books []model.Book; if err := q.Find(&books).Error; err != nil { return nil, translateError(err) }
6. if books == nil { books = []model.Book{} }; return books, nil

Remove the loop and the NAIVE comment completely.
```

**Перевір сам:**
Створи 5+ книг, виклич `GET /api/v1/books` і порахуй `SELECT` у консолі GORM: було 1+2N, стало 3. Зафіксуй обидва числа (до та після) для демонстрації.

---

### 8.24 — Репозиторій авторів

```
RULES: Output the COMPLETE content of the file below (whole file, never a diff, no explanations). Allowed imports ONLY: "gorm.io/gorm", "libratrack/internal/model". Nothing else.

Create internal/repository/author_repository.go (package repository).

1. type AuthorRepository interface {
       Create(author *model.Author) error
       FindByID(id uint) (*model.Author, error)
       List() ([]model.Author, error)
   }
2. type gormAuthorRepository struct { db *gorm.DB }
   func NewAuthorRepository(db *gorm.DB) AuthorRepository { return &gormAuthorRepository{db: db} }
3. Create: if err := r.db.Create(author).Error; err != nil { return translateError(err) }; return nil
4. FindByID: var a model.Author; if err := r.db.First(&a, id).Error; err != nil { return nil, translateError(err) }; return &a, nil
5. List: var authors []model.Author; if err := r.db.Order("full_name ASC").Find(&authors).Error; err != nil { return nil, translateError(err) }; if authors == nil { authors = []model.Author{} }; return authors, nil
```

**Перевір сам:**
`go build ./internal/repository/...` проходить.

---

### 8.25 — Репозиторій категорій

```
RULES: Output the COMPLETE content of the file below (whole file, never a diff, no explanations). Allowed imports ONLY: "gorm.io/gorm", "libratrack/internal/model". Nothing else.

Create internal/repository/category_repository.go (package repository).

1. type CategoryRepository interface {
       Create(category *model.Category) error
       FindByID(id uint) (*model.Category, error)
       List() ([]model.Category, error)
   }
2. type gormCategoryRepository struct { db *gorm.DB }
   func NewCategoryRepository(db *gorm.DB) CategoryRepository { return &gormCategoryRepository{db: db} }
3. Create: if err := r.db.Create(category).Error; err != nil { return translateError(err) }; return nil
4. FindByID: var c model.Category; if err := r.db.First(&c, id).Error; err != nil { return nil, translateError(err) }; return &c, nil
5. List: var categories []model.Category; if err := r.db.Order("name ASC").Find(&categories).Error; err != nil { return nil, translateError(err) }; if categories == nil { categories = []model.Category{} }; return categories, nil
```

**Перевір сам:**
Дублікат назви категорії дасть `ErrConflict` (через `translateError`, колонка `name` унікальна).

---

### 8.26 — AuthorHandler

```
RULES: Output the COMPLETE content of the file below (whole file, never a diff, no explanations). Allowed imports ONLY: "net/http", "github.com/gin-gonic/gin", "libratrack/internal/model", "libratrack/internal/repository". Nothing else.

Create internal/handler/author_handler.go (package handler).

1. type AuthorHandler struct { repo repository.AuthorRepository }
   func NewAuthorHandler(repo repository.AuthorRepository) *AuthorHandler
2. func (h *AuthorHandler) Create(c *gin.Context):
   bind model.CreateAuthorRequest with ShouldBindJSON (error -> respondValidation, return); author := model.Author{FullName: req.FullName, Bio: req.Bio}; h.repo.Create(&author) (error -> respondRepoError, return); c.JSON(http.StatusCreated, author)
3. func (h *AuthorHandler) GetByID(c *gin.Context):
   id, ok := parseID(c); if !ok return; h.repo.FindByID(id) (error -> respondRepoError, return); c.JSON(http.StatusOK, author)
4. func (h *AuthorHandler) List(c *gin.Context):
   h.repo.List() (error -> respondRepoError, return); c.JSON(http.StatusOK, authors)
```

**Перевір сам:**
`go build ./internal/handler/...` проходить.

---

### 8.27 — CategoryHandler

```
RULES: Output the COMPLETE content of the file below (whole file, never a diff, no explanations). Allowed imports ONLY: "net/http", "github.com/gin-gonic/gin", "libratrack/internal/model", "libratrack/internal/repository". Nothing else.

Create internal/handler/category_handler.go (package handler).

1. type CategoryHandler struct { repo repository.CategoryRepository }
   func NewCategoryHandler(repo repository.CategoryRepository) *CategoryHandler
2. func (h *CategoryHandler) Create(c *gin.Context):
   bind model.CreateCategoryRequest with ShouldBindJSON (error -> respondValidation, return); category := model.Category{Name: req.Name}; h.repo.Create(&category) (error -> respondRepoError, return; a duplicate name becomes 409 through ErrConflict); c.JSON(http.StatusCreated, category)
3. func (h *CategoryHandler) GetByID(c *gin.Context):
   id, ok := parseID(c); if !ok return; h.repo.FindByID(id) (error -> respondRepoError, return); c.JSON(http.StatusOK, category)
4. func (h *CategoryHandler) List(c *gin.Context):
   h.repo.List() (error -> respondRepoError, return); c.JSON(http.StatusOK, categories)
```

**Перевір сам:**
Структура ідентична `AuthorHandler`, лише з категоріями.

---

### 8.28 — Репозиторій примірників

```
RULES: Output the COMPLETE content of the file below (whole file, never a diff, no explanations). Allowed imports ONLY: "gorm.io/gorm", "libratrack/internal/model". Nothing else.

Create internal/repository/copy_repository.go (package repository).

1. type CopyRepository interface {
       Create(copy *model.Copy) error
       ListByBook(bookID uint) ([]model.Copy, error)
       Availability(bookID uint) (*model.BookAvailability, error)
   }
2. type gormCopyRepository struct { db *gorm.DB }
   func NewCopyRepository(db *gorm.DB) CopyRepository { return &gormCopyRepository{db: db} }
3. Private helper func (r *gormCopyRepository) ensureBookExists(bookID uint) error:
   var n int64; if err := r.db.Model(&model.Book{}).Where("id = ?", bookID).Count(&n).Error; err != nil { return translateError(err) }; if n == 0 { return ErrNotFound }; return nil
4. Create: if err := r.ensureBookExists(copy.BookID); err != nil { return err }; if copy.Status == "" { copy.Status = model.CopyStatusAvailable }; if err := r.db.Create(copy).Error; err != nil { return translateError(err) }; return nil
5. ListByBook: ensureBookExists first (return its error); then var copies []model.Copy; r.db.Where("book_id = ?", bookID).Order("id ASC").Find(&copies); on error translateError; if copies == nil use an empty slice; return copies, nil
6. Availability: ensureBookExists first. Then compute with ONE SQL query (do not load rows into memory):
   result := model.BookAvailability{BookID: bookID}
   err := r.db.Raw("SELECT COUNT(*) AS total_copies, COUNT(*) FILTER (WHERE status = 'available') AS available_copies FROM copies WHERE book_id = ?", bookID).Scan(&result).Error
   on error return nil, translateError(err); result.BookID = bookID; return &result, nil
```

**Перевір сам:**
У `Availability` один SQL-запит з `COUNT ... FILTER`, а не цикл. `go build ./internal/repository/...` проходить.

---

### 8.29 — CopyHandler: Create і ListByBook

```
RULES: Output the COMPLETE content of the file below (whole file, never a diff, no explanations). Allowed imports ONLY: "net/http", "github.com/gin-gonic/gin", "libratrack/internal/model", "libratrack/internal/repository". Nothing else. The Availability method is implemented in ANOTHER file, do NOT write it here.

Create internal/handler/copy_handler.go (package handler).

1. type CopyHandler struct { repo repository.CopyRepository }
   func NewCopyHandler(repo repository.CopyRepository) *CopyHandler
2. func (h *CopyHandler) Create(c *gin.Context) (route POST /books/:id/copies):
   bookID, ok := parseID(c); if !ok return
   var req model.CreateCopyRequest; if err := c.ShouldBindJSON(&req); err != nil { respondValidation(c, err); return }
   copy := model.Copy{BookID: bookID, InventoryNo: req.InventoryNo, Status: req.Status}
   if err := h.repo.Create(&copy); err != nil { respondRepoError(c, err); return }   (404 for unknown book, 409 for duplicate inventory number)
   c.JSON(http.StatusCreated, copy)
3. func (h *CopyHandler) ListByBook(c *gin.Context) (route GET /books/:id/copies):
   bookID, ok := parseID(c); if !ok return
   copies, err := h.repo.ListByBook(bookID); if err != nil { respondRepoError(c, err); return }
   c.JSON(http.StatusOK, copies)
```

**Перевір сам:**
Збірка пакета handler ще проходить без `Availability`; метод додаємо на наступному кроці.

---

### 8.30 — CopyHandler: доступність книги

```
RULES: Output the COMPLETE content of the file below (whole file, never a diff, no explanations). Allowed imports ONLY: "net/http", "github.com/gin-gonic/gin". Nothing else.

Create internal/handler/copy_availability_handler.go (package handler) with one method (route GET /books/:id/availability):

func (h *CopyHandler) Availability(c *gin.Context):
   bookID, ok := parseID(c); if !ok return
   result, err := h.repo.Availability(bookID); if err != nil { respondRepoError(c, err); return }
   c.JSON(http.StatusOK, result)
The JSON must look like {"book_id": 1, "total_copies": 5, "available_copies": 3}.
```

**Перевір сам:**
`go build ./internal/handler/...` проходить.

---

### 8.31 — main.go: усі маршрути і /health з перевіркою БД

```
RULES: Output the COMPLETE content of the file below (whole file, never a diff, no explanations). Keep ALL existing imports that are still used, the swag annotations, the Swagger route, the middleware registration order (Recovery, Logging, CORS), the config reading, the migrations call, the database connection and the existing book routes EXACTLY as they are in the current file. Additional allowed imports: "context", "net/http", "time".

Update cmd/api/main.go:
1. After creating bookRepo and bookHandler, also create authorRepo := repository.NewAuthorRepository(db), categoryRepo := repository.NewCategoryRepository(db), copyRepo := repository.NewCopyRepository(db), and handlers authorHandler := handler.NewAuthorHandler(authorRepo), categoryHandler := handler.NewCategoryHandler(categoryRepo), copyHandler := handler.NewCopyHandler(copyRepo).
2. Register in the /api/v1 group, in addition to the book routes:
   POST /books/:id/copies -> copyHandler.Create
   GET /books/:id/copies -> copyHandler.ListByBook
   GET /books/:id/availability -> copyHandler.Availability
   POST /authors -> authorHandler.Create; GET /authors -> authorHandler.List; GET /authors/:id -> authorHandler.GetByID
   POST /categories -> categoryHandler.Create; GET /categories -> categoryHandler.List; GET /categories/:id -> categoryHandler.GetByID
3. Change GET /health: get sqlDB, err := db.DB(); create ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second); defer cancel(); if err != nil or sqlDB.PingContext(ctx) fails, respond c.JSON(http.StatusServiceUnavailable, gin.H{"status": "error"}); otherwise c.JSON(http.StatusOK, gin.H{"status": "ok"}).
Do not add detailed @Summary/@Param annotations.
```

**Перевір сам:**
`go build ./...` зелена. Повний сценарій curl: створи категорію `POST /api/v1/categories`, автора `POST /api/v1/authors`, книгу, примірник `POST /api/v1/books/1/copies {"inventory_no":"INV-X1"}` і `GET /api/v1/books/1/availability`. Зупини контейнер і переконайся, що `/health` повертає 503.

---

### 8.32 — Seed-міграція (частина 1): категорії та автори

```
RULES: Output the COMPLETE content of the file below (whole file, never a diff, no explanations). Plain SQL (PostgreSQL), ASCII only (no accented or non-Latin characters).

Create migrations/000002_seed_demo_data.up.sql containing exactly these two statements:

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
```

**Перевір сам:**
Файл містить лише ASCII. Далі (кроки 8.33-8.34) його розширюємо.

---

### 8.33 — Seed-міграція (частина 2): книги та зв'язки з авторами

```
RULES: Output the COMPLETE content of the file migrations/000002_seed_demo_data.up.sql (whole file, never a diff, no explanations). Keep the two existing INSERT statements (categories and authors) EXACTLY as they are and append the following at the end. Plain SQL, ASCII only.

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
```

**Перевір сам:**
Одна книга має двох авторів (Go, Pragmatic), один автор має дві книги (Asimov). ISBN можна замінити на власні.

---

### 8.34 — Seed-міграція (частина 3): примірники та послідовності

```
RULES: Output the COMPLETE content of the file migrations/000002_seed_demo_data.up.sql (whole file, never a diff, no explanations). Keep ALL existing statements EXACTLY as they are and append the following at the end. Plain SQL, ASCII only.

INSERT INTO copies (book_id, inventory_no, status) VALUES
  (1, 'INV-0001', 'available'),
  (1, 'INV-0002', 'available'),
  (1, 'INV-0003', 'borrowed'),
  (1, 'INV-0004', 'available'),
  (2, 'INV-0005', 'available'),
  (2, 'INV-0006', 'available'),
  (2, 'INV-0007', 'borrowed'),
  (3, 'INV-0008', 'available'),
  (3, 'INV-0009', 'available'),
  (3, 'INV-0010', 'available'),
  (4, 'INV-0011', 'available'),
  (4, 'INV-0012', 'lost'),
  (4, 'INV-0013', 'available'),
  (5, 'INV-0014', 'available'),
  (5, 'INV-0015', 'available'),
  (5, 'INV-0016', 'available'),
  (6, 'INV-0017', 'available'),
  (6, 'INV-0018', 'under_repair'),
  (6, 'INV-0019', 'available'),
  (7, 'INV-0020', 'available'),
  (7, 'INV-0021', 'available'),
  (7, 'INV-0022', 'available'),
  (8, 'INV-0023', 'available'),
  (8, 'INV-0024', 'available'),
  (8, 'INV-0025', 'available');

SELECT setval(pg_get_serial_sequence('categories', 'id'), (SELECT MAX(id) FROM categories));
SELECT setval(pg_get_serial_sequence('authors', 'id'), (SELECT MAX(id) FROM authors));
SELECT setval(pg_get_serial_sequence('books', 'id'), (SELECT MAX(id) FROM books));
SELECT setval(pg_get_serial_sequence('copies', 'id'), (SELECT MAX(id) FROM copies));
```

**Перевір сам:**
Вставлено 25 примірників, статуси змішані. Без `setval` наступний `INSERT` через API впаде на конфлікті первинного ключа.

---

### 8.35 — Seed-міграція: down

```
RULES: Output the COMPLETE content of the file below (whole file, never a diff, no explanations). Plain SQL.

Create migrations/000002_seed_demo_data.down.sql with exactly these lines, in this order:
DELETE FROM copies WHERE book_id BETWEEN 1 AND 8;
DELETE FROM book_authors WHERE book_id BETWEEN 1 AND 8;
DELETE FROM books WHERE id BETWEEN 1 AND 8;
DELETE FROM authors WHERE id BETWEEN 1 AND 9;
DELETE FROM categories WHERE id BETWEEN 1 AND 4;
```

**Перевір сам:**
Перезапусти застосунок: міграція 2 застосується автоматично (`docker exec -it libratrack-db psql -U libratrack -c "SELECT count(*) FROM books;"` повертає 8). Якщо ти вже створював тестові записи вручну з id 1..4 на кроках 8.22/8.31, очисти БД: `docker compose down -v && docker compose up -d`.

---

### 8.36 — Тести: допоміжний код для БД

```
RULES: Output the COMPLETE content of the file below (whole file, never a diff, no explanations). Allowed imports ONLY: "os", "testing", "gorm.io/gorm", "libratrack/internal/model". Nothing else.

Create internal/repository/testhelper_test.go (package repository).

1. func setupTestDB(t *testing.T) *gorm.DB:
   - t.Helper()
   - url := os.Getenv("TEST_DATABASE_URL"); if url == "" { t.Skip("TEST_DATABASE_URL is not set") }
   - if err := RunMigrations(url, "../../migrations"); err != nil { t.Fatalf("migrations: %v", err) }
   - db, err := NewPostgresDB(url); if err != nil { t.Fatalf("db: %v", err) }
   - if err := db.Exec("TRUNCATE TABLE book_authors, copies, books, authors, categories, users RESTART IDENTITY CASCADE").Error; err != nil { t.Fatalf("truncate: %v", err) }
   - return db
2. func mustCategory(t *testing.T, db *gorm.DB, name string) model.Category: creates a category with the name, fails the test on error, returns it.
3. func mustAuthor(t *testing.T, db *gorm.DB, name string) model.Author: creates an author with FullName = name, fails the test on error, returns it.
4. func mustBook(t *testing.T, db *gorm.DB, title, isbn string, categoryID uint, authorIDs []uint) model.Book:
   book := model.Book{Title: title, ISBN: isbn, CategoryID: categoryID}; err := NewBookRepository(db).Create(&book, authorIDs); fails the test on error; returns book.
```

**Перевір сам:**
Без `TEST_DATABASE_URL` тести будуть пропущені (SKIP), а не впадуть. Створи окрему тестову БД: `docker exec -it libratrack-db psql -U libratrack -c "CREATE DATABASE libratrack_test;"`. Тести очищають таблиці, тому не вказуй робочу БД.

---

### 8.37 — Тести книг: Create і FindByID

```
RULES: Output the COMPLETE content of the file below (whole file, never a diff, no explanations). Allowed imports ONLY: "errors", "testing", "libratrack/internal/model". Nothing else. Use the helpers setupTestDB, mustCategory, mustAuthor, mustBook from testhelper_test.go.

Create internal/repository/book_repository_create_test.go (package repository) with these test functions:

1. TestBookCreateAndFind: setupTestDB; create category and two authors; repo := NewBookRepository(db); create a book with both author IDs through the repo; FindByID(book.ID) must return the book with len(Authors) == 2, a non-nil Category with the correct name.
2. TestBookCreateDuplicateISBN: creating a second book with the same ISBN returns an error for which errors.Is(err, ErrConflict) is true.
3. TestBookCreateInvalidCategory: category_id 9999 returns an error for which errors.Is(err, ErrInvalidReference) is true.
4. TestBookCreateInvalidAuthor: authorIDs containing 9999 returns ErrInvalidReference, and afterwards no author with id 9999 exists in the table authors (count with db.Model(&model.Author{}).Where("id = ?", 9999).Count(&n)).
5. TestBookFindNotFound: FindByID(9999) returns ErrNotFound.
```

**Перевір сам:**
`go test ./internal/repository/... -run TestBook -v` без `TEST_DATABASE_URL` дає SKIP, а з ним PASS. Четвертий тест перевіряє, що «порожній» автор не створюється мовчки.

---

### 8.38 — Тести книг: Update і Delete

```
RULES: Output the COMPLETE content of the file below (whole file, never a diff, no explanations). Allowed imports ONLY: "errors", "testing", "libratrack/internal/model". Nothing else. Use the helpers from testhelper_test.go.

Create internal/repository/book_repository_update_test.go (package repository) with these test functions:

1. TestBookUpdate: setupTestDB; category, two authors, a book with the first author; call repo.Update with a model.Book{ID: book.ID, Title: "New Title", ISBN: book.ISBN, CategoryID: category.ID} and authorIDs of the SECOND author only; then FindByID must show Title "New Title" and exactly one author, the second one.
2. TestBookUpdateNotFound: Update of a book with ID 9999 returns ErrNotFound.
3. TestBookDelete: delete the book; then FindByID returns ErrNotFound; also insert a copy for the book BEFORE deleting (use db.Create(&model.Copy{BookID: book.ID, InventoryNo: "INV-T1"})) and check after deletion that the number of copies with this book_id is 0 (cascade).
4. TestBookDeleteNotFound: Delete(9999) returns ErrNotFound.
```

**Перевір сам:**
Тест видалення перевіряє каскад на примірниках (`ON DELETE CASCADE` у міграції).

---

### 8.39 — Тести книг: List

```
RULES: Output the COMPLETE content of the file below (whole file, never a diff, no explanations). Allowed imports ONLY: "testing". Nothing else. Use the helpers from testhelper_test.go.

Create internal/repository/book_repository_list_test.go (package repository) with these test functions:

1. TestBookListByCategory: setupTestDB; two categories "A" and "B"; one author; three books in category A and two in category B (unique ISBNs); List(categoryA.ID, 1, 20) returns 3 books; List(categoryB.ID, 1, 20) returns 2; List(0, 1, 20) returns 5.
2. TestBookListPagination: five books in one category; List(0, 1, 2) returns 2 books; List(0, 3, 2) returns 1 book; List(0, 4, 2) returns 0 books and a non-nil empty slice.
3. TestBookListLoadsRelations: every returned book has a non-nil Category and exactly the authors that were assigned.
```

**Перевір сам:**
Тест №3 проходить і з наївною, і з Preload-реалізацією; число запитів перевіряється вручну по консолі GORM (крок 8.23).

---

### 8.40 — Тести примірників і доступності

```
RULES: Output the COMPLETE content of the file below (whole file, never a diff, no explanations). Allowed imports ONLY: "errors", "testing", "libratrack/internal/model". Nothing else. Use the helpers from testhelper_test.go.

Create internal/repository/copy_repository_test.go (package repository) with these test functions:

1. TestCopyCreateAndList: setupTestDB; category, author, book; create two copies with NewCopyRepository(db).Create (statuses empty and "lost"); the first one must get the status "available" automatically; ListByBook returns 2 copies.
2. TestCopyCreateUnknownBook: Create for BookID 9999 returns ErrNotFound.
3. TestCopyDuplicateInventoryNo: the same InventoryNo twice returns ErrConflict.
4. TestAvailability: book with 4 copies: 2 "available", 1 "borrowed", 1 "lost"; Availability returns TotalCopies 4 and AvailableCopies 2.
5. TestAvailabilityNoCopies: a book without copies returns 0 and 0 (not an error).
6. TestAvailabilityUnknownBook: Availability(9999) returns ErrNotFound.
```

**Перевір сам:**
Для копій використовуй `model.CopyStatusLost` тощо; не пиши рядки статусів вручну, окрім порожнього.

---

### 8.41 — Тести хендлера книг (без БД, з fake-репозиторієм)

```
RULES: Output the COMPLETE content of the file below (whole file, never a diff, no explanations). Allowed imports ONLY: "bytes", "encoding/json", "net/http", "net/http/httptest", "testing", "github.com/gin-gonic/gin", "libratrack/internal/model", "libratrack/internal/repository". Nothing else.

Create internal/handler/book_handler_test.go (package handler).

1. A fake repository type fakeBookRepo implementing repository.BookRepository with function fields:
   createFn func(book *model.Book, authorIDs []uint) error
   findFn func(id uint) (*model.Book, error)
   deleteFn func(id uint) error
   The methods Create, FindByID and Delete call these functions; List returns an empty slice and nil; Update returns nil.
2. A helper func newTestRouter(repo repository.BookRepository) *gin.Engine: gin.SetMode(gin.TestMode); r := gin.New(); h := NewBookHandler(repo); registers POST /books -> h.Create, GET /books/:id -> h.GetByID, DELETE /books/:id -> h.Delete.
3. Tests (use httptest.NewRecorder and http.NewRequest, JSON bodies built with bytes.NewBufferString):
   - TestCreateBook201: fake createFn sets book.ID = 1 and returns nil; valid JSON {"title":"T","isbn":"1234567890","category_id":1,"author_ids":[1]}; expect 201 and a JSON body with "id" equal to 1.
   - TestCreateBookValidation422: body {"isbn":"1234567890"} (no title); expect 422 and the JSON path error.code == "validation_error".
   - TestCreateBookConflict409: createFn returns repository.ErrConflict; expect 409 and error.code == "conflict".
   - TestGetBookNotFound404: findFn returns nil, repository.ErrNotFound; expect 404 and error.code == "not_found".
   - TestGetBookInvalidID400: GET /books/abc; expect 400 and error.code == "invalid_id".
   - TestDeleteBook204: deleteFn returns nil; expect 204.
```

**Перевір сам:**
`go test ./internal/handler/... -v`: усі тести проходять без БД.

---

### 8.42 — Тести хендлера примірників і доступності

```
RULES: Output the COMPLETE content of the file below (whole file, never a diff, no explanations). Allowed imports ONLY: "bytes", "encoding/json", "net/http", "net/http/httptest", "testing", "github.com/gin-gonic/gin", "libratrack/internal/model", "libratrack/internal/repository". Nothing else.

Create internal/handler/copy_handler_test.go (package handler).

1. A fake type fakeCopyRepo implementing repository.CopyRepository with function fields createFn, listFn and availabilityFn (each with the same signature as the interface method); the interface methods call them.
2. A helper func newCopyRouter(repo repository.CopyRepository) *gin.Engine: gin.SetMode(gin.TestMode); gin.New(); NewCopyHandler(repo); routes POST /books/:id/copies -> Create, GET /books/:id/copies -> ListByBook, GET /books/:id/availability -> Availability.
3. Tests:
   - TestAvailabilityOK: availabilityFn returns &model.BookAvailability{BookID: 1, TotalCopies: 5, AvailableCopies: 3}; GET /books/1/availability; expect 200 and a JSON body with total_copies 5 and available_copies 3.
   - TestAvailabilityBookNotFound: availabilityFn returns nil, repository.ErrNotFound; expect 404.
   - TestCreateCopy201: valid body {"inventory_no":"INV-1"}; createFn returns nil; expect 201.
   - TestCreateCopyValidation422: empty body {}; expect 422.
   - TestCreateCopyDuplicate409: createFn returns repository.ErrConflict; expect 409.
```

**Перевір сам:**
`go test ./internal/handler/... -v` проходить без БД.

---

### 8.43 — Фінальна ручна перевірка (без Aider, для малої моделі)

```
RULES: Output the COMPLETE content of the file below (whole file, never a diff, no explanations).

Create docs/theme8-checklist.md in English with a short checklist (markdown checkboxes) for theme 8 with exactly these items, one per line:
- go mod tidy works
- go vet ./... reports no problems
- go build ./... succeeds
- go test ./... passes (integration tests are skipped without TEST_DATABASE_URL)
- go test ./... passes with TEST_DATABASE_URL set to a separate test database
- docker compose up -d starts PostgreSQL and the app applies migrations 000001 and 000002 automatically
- GET /health returns {"status":"ok"} and returns 503 when the database is stopped
- all CRUD endpoints work under /api/v1 (books, authors, categories, copies, availability)
- there are no AutoMigrate calls in the project
- ErrNotFound is declared in exactly one place
- there are no non-ASCII characters in any .go file
- GET /api/v1/books uses 3 SELECT queries regardless of the number of books
```

**Перевір сам:**
Цей крок лише фіксує чекліст у репозиторії; фінальні перевірки робимо вручну командами нижче (аудит усього проєкту за один промпт для 7B-моделі нереальний):
- `go mod tidy && go vet ./... && go build ./... && go test ./...`
- `grep -rn AutoMigrate --include=*.go .` (нічого не знайдено)
- `grep -rn "ErrNotFound = " --include=*.go .` (рівно один рядок, у `internal/repository/errors.go`)
- `grep -rnP '[^\x00-\x7F]' --include=*.go .` (порожній вивід; на Windows PowerShell: `Select-String -Path (Get-ChildItem -Recurse -Filter *.go) -Pattern '[^\x00-\x7F]'`)
- повний сценарій curl: категорія -> автор -> книга -> примірники -> availability, а також `GET /api/v1/books` з підрахунком `SELECT` у консолі.
Закоміть і поставте тег `theme-8`: це стартова точка для теми 9. Для моделей 14B+ можна використати оригінальний аудит із `tema8.md`.
