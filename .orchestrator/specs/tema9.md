## ТЕМА 9. Security First: JWT, bcrypt, RBAC

Мета блоку: перетворити відкритий CRUD з теми 8 на систему з користувачами, ролями та обмеженим доступом: реєстрація/логін, bcrypt-хешування паролів, JWT-автентифікація, RBAC-middleware, сутність Loan (видача/повернення книг зі штрафом за прострочення). **Структура каталогів, маршрути `/api/v1/...` і формат помилок з теми 8 не змінюються** — додаємо нові файли й закриваємо існуючі маршрути middleware-ами.

**Загальні правила для всіх кроків:**
- Output ONLY the exact file contents in the required Aider format.
- No explanations, no extra text, no markdown outside the code blocks.
- Do not invent extra files or change the directory structure, except the files explicitly listed in the step.
- Use only UTF-8. Never insert non-ASCII characters into .go files (error messages and comments must be in English).
- Keep code simple and correct for Go 1.22+.
- Keep the layers: cmd/api, internal/model, internal/handler, internal/service, internal/repository, internal/middleware. Do not create new internal/ packages. Handlers only bind input, call a service or repository and map errors to HTTP statuses. Business rules live in internal/service. Repositories only access data.
- Security rules: never log, print or return passwords, password hashes, the JWT secret or tokens. Never hardcode secrets in Go code. Do not use dependencies other than those from step 9.1.
- Every error response (including 401 and 403) uses the existing envelope {"error": {"code": "...", "message": "...", "details": "..."}}.
- Do NOT use AutoMigrate anywhere; the schema is managed only by golang-migrate.

**Перед стартом (робить студент, не агент):**
- Проєкт з теми 8 збирається, тести проходять (`go build ./... && go test ./...`), є тег `theme-8`.
- `docker compose up -d` — БД у статусі healthy.
- Згенеруй секрет для JWT: `openssl rand -base64 48`. Вставиш його у свій `.env` (НЕ в `.env.example` і не в Git).
- У Aider додавай у чат (`/add`) файли, які крок змінює (`main.go`, `book_handler.go`, `errors.go` тощо).
- **Прийми бізнес-рішення (див. таблицю нижче) до кроку 9.7.**

**Рішення студента (межа ШІ):** розрахунок штрафу й правила видачі — це продуктові рішення, які ШІ приймати не повинен. Якщо в промпті немає чіткого правила, агент вигадає своє. Нижче — значення за замовчуванням, використані в промптах; обговори їх (з ментором або з «бібліотекарем») і за потреби зміни в `.env.example` (9.1), промпті 9.7 і тестах 9.13.

| Питання | Значення в промптах |
|---|---|
| Строк видачі | 14 днів |
| Штраф за кожен прострочений день | 2 грн (цілі гривні, без float) |
| Максимальний штраф за одну видачу | 100 грн |
| Як рахувати прострочені дні | повні календарні дні за UTC після дати повернення за графіком |
| Макс. активних видач на читача | 5 |

**Матриця доступу (перевір перед кроком 9.11):**

| Ендпоінт | Хто має доступ |
|---|---|
| `GET /health`, `GET /swagger/*`, `POST /auth/register`, `POST /auth/login` | усі (без токена) |
| `GET /me`, `GET /me/loans`, усі `GET` каталогу (books, authors, categories, copies, availability) | будь-який автентифікований (reader, librarian, admin) |
| `POST /books`, `PUT /books/:id`, `POST /books/:id/copies`, `POST /authors`, `POST /categories` | librarian, admin |
| `POST /loans`, `POST /loans/:id/return`, `GET /loans` | librarian, admin |
| `DELETE /books/:id`, `GET /users`, `PUT /users/:id/role` | admin |

---

### 9.1 — Конфігурація та залежності

```
Add configuration and dependencies for authentication and loans to the "libratrack" project.

1. Update .env.example (keep all existing variables and add these):
   JWT_SECRET=change-me-to-a-random-string-of-at-least-32-characters
   JWT_TTL_MINUTES=60
   ADMIN_EMAIL=admin@libratrack.local
   ADMIN_PASSWORD=change-me-admin-password
   SEED_DEMO_USERS=false
   DEMO_USERS_PASSWORD=change-me-demo-password
   LOAN_PERIOD_DAYS=14
   FINE_PER_DAY=2
   FINE_MAX=100
   MAX_ACTIVE_LOANS=5

2. Add dependencies to go.mod (do not change any Go source files yet):
   - github.com/golang-jwt/jwt/v5
   - golang.org/x/crypto (used for bcrypt)

Do not modify any other files (.gitignore already ignores .env).
```

**Перевір сам:**
`.env.example` містить лише заглушки, жодного реального секрету. Свій `.env`: `cp .env.example .env`, потім заміни `JWT_SECRET` і паролі на власні. `.env` у `.gitignore` (`git status` його не показує).

---

### 9.2 — Модель: ролі, запити автентифікації, Loan

```
Update the domain model in internal/model/ for authentication and loans. Create/modify these files:

1. internal/model/user.go (modify the existing file)
   - Keep all existing fields and tags of User. Replace the old comment saying that the table exists only for the schema in theme 8 with a comment that User is now used for authentication and RBAC.
   - PasswordHash must keep the JSON tag "-" so it is never serialized.
   - Add role constants: RoleAdmin = "admin", RoleLibrarian = "librarian", RoleReader = "reader", and func IsValidRole(role string) bool.
   - Add RegisterRequest with JSON tags email, password, full_name and Gin binding tags: email (required,email,max=255), password (required,min=8,max=72), full_name (required,max=255). Add a comment that 72 is the bcrypt input limit in bytes.
   - Add LoginRequest: email (required,email), password (required).
   - Add AuthResponse: AccessToken (json "access_token"), TokenType (json "token_type"), ExpiresIn (int, seconds, json "expires_in"), User (User, json "user").
   - Add UpdateUserRoleRequest: Role (string, json "role", binding "required,oneof=admin librarian reader").

2. internal/model/copy.go (modify the existing file)
   - Keep everything as is and add one field to Copy: Book (*Book, json "book,omitempty", gorm:"foreignKey:BookID"). Add a comment that Book and Book.Copies must never be preloaded together (JSON cycle).

3. internal/model/loan.go (new)
   - Loan struct: ID (uint, primary key), CopyID (uint), UserID (uint), BorrowedAt (time.Time), DueAt (time.Time), ReturnedAt (*time.Time, nullable, json "returned_at"), FineAmount (int, whole UAH, json "fine_amount"; add a comment that money is stored as an integer and never as float), Copy (*Copy, json "copy,omitempty", gorm:"foreignKey:CopyID"), User (*User, json "user,omitempty", gorm:"foreignKey:UserID").
   - Table name must be "loans" (columns: id, copy_id, user_id, borrowed_at, due_at, returned_at, fine_amount).
   - Add CreateLoanRequest: CopyID (uint, json "copy_id", binding "required,gt=0"), UserID (uint, json "user_id", binding "required,gt=0").
   - Add type LoanStatus filter constants as plain strings: LoanStatusActive = "active", LoanStatusReturned = "returned", LoanStatusOverdue = "overdue", and func IsValidLoanStatus(s string) bool.

Use GORM struct tags (gorm:"...") and JSON tags (snake_case) on all models.
Do not touch any other files.
```

**Перевір сам:**
У `User.PasswordHash` тег `json:"-"`. У `RegisterRequest` немає поля `role` (роль при реєстрації клієнт задавати не може — це свідоме рішення безпеки). `go build ./...` проходить.

---

### 9.3 — Міграція: таблиця loans

```
Create the next SQL migration for golang-migrate. The runner internal/repository/migrate.go must not be changed.

1. migrations/000003_auth_and_loans.up.sql:
   - CREATE UNIQUE INDEX users_email_lower_key ON users (lower(email)); (case-insensitive unique email)
   - loans: id BIGSERIAL PRIMARY KEY, copy_id BIGINT NOT NULL REFERENCES copies(id) ON DELETE RESTRICT, user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE RESTRICT, borrowed_at TIMESTAMPTZ NOT NULL DEFAULT now(), due_at TIMESTAMPTZ NOT NULL, returned_at TIMESTAMPTZ NULL, fine_amount INT NOT NULL DEFAULT 0 CHECK (fine_amount >= 0), CHECK (due_at > borrowed_at), CHECK (returned_at IS NULL OR returned_at >= borrowed_at)
   - indexes: loans(user_id), loans(copy_id), a partial UNIQUE index loans_one_active_per_copy ON loans(copy_id) WHERE returned_at IS NULL (a copy can have only one active loan), a partial index on loans(due_at) WHERE returned_at IS NULL

2. migrations/000003_auth_and_loans.down.sql: drop the indexes and the loans table, then drop users_email_lower_key (use IF EXISTS everywhere).

Do not modify any other files.
```

**Перевір сам:**
Файли `.up.sql` / `.down.sql` мають однаковий номер версії `000003`. Після підключення в `main.go` (9.11): `docker exec -it libratrack-db psql -U libratrack -c "\d loans"` — видно FK, CHECK і частковий унікальний індекс.

---

### 9.4 — Хешування паролів і JWT (чисті компоненти без HTTP і БД)

```
Add the security primitives to internal/service/ (package service). They must not depend on Gin, GORM or the database.

1. internal/service/password.go
   - package-level var bcryptCost = bcrypt.DefaultCost (a variable so tests can lower it), using golang.org/x/crypto/bcrypt
   - var ErrPasswordInvalid = errors.New("password must be between 1 and 72 bytes")
   - func HashPassword(plain string) (string, error): returns ErrPasswordInvalid if the password is empty or longer than 72 bytes; otherwise bcrypt.GenerateFromPassword
   - func CheckPassword(hash, plain string) bool: returns false if hash is empty; otherwise bcrypt.CompareHashAndPassword == nil
   - func CheckPasswordDummy(plain string): compares plain against a lazily computed (sync.Once) bcrypt hash of a constant string and ignores the result. Add a comment: it equalizes response time when the email is unknown, so login timing does not reveal which emails exist.

2. internal/service/token.go
   - var ErrInvalidToken = errors.New("invalid token")
   - type Claims struct { Role string `json:"role"`; jwt.RegisteredClaims } using github.com/golang-jwt/jwt/v5
   - type TokenManager struct with unexported fields: secret []byte, ttl time.Duration, now func() time.Time (default time.Now; tests may replace it)
   - const tokenIssuer = "libratrack"
   - func NewTokenManager(secret string, ttl time.Duration) (*TokenManager, error): error if len(secret) < 32 or ttl <= 0
   - func (m *TokenManager) TTL() time.Duration
   - func (m *TokenManager) Issue(userID uint, role string) (string, time.Time, error): HS256; claims: Subject = decimal user id, Role, Issuer = tokenIssuer, IssuedAt = now, ExpiresAt = now + ttl; returns the signed token and the expiry time
   - func (m *TokenManager) Verify(token string) (uint, string, error): parse with jwt.ParseWithClaims and the options jwt.WithValidMethods([]string{"HS256"}), jwt.WithIssuer(tokenIssuer), jwt.WithExpirationRequired(), jwt.WithTimeFunc(m.now). Then check that the subject parses to a uint greater than 0 and that model.IsValidRole(claims.Role) is true. Any failure returns ErrInvalidToken (wrap the cause with %w but never include the token in the message).

Do not modify any other files.
```

**Перевір сам:**
У `Verify` зафіксований алгоритм HS256 (захист від підміни алгоритму, зокрема `alg: none`). Секрет передається ззовні, у коді його немає. Помилки не містять токена.

---

### 9.5 — Middleware: автентифікація та RBAC

```
Add authentication and role-based access control middleware to internal/middleware/.

1. internal/middleware/auth.go (package middleware)
   - type TokenVerifier interface { Verify(token string) (userID uint, role string, err error) }
   - unexported context keys for user id and role
   - func AuthMiddleware(v TokenVerifier) gin.HandlerFunc: read the Authorization header; it must have the form "Bearer <token>" (scheme compared case-insensitively). If the header is missing, malformed or the token fails verification: abort with 401, set header "WWW-Authenticate: Bearer", and respond with the JSON envelope {"error": {"code": "unauthorized", "message": "authentication required", "details": ""}}. Use the same generic message for all failure reasons. On success store user id and role in the Gin context and call c.Next().
   - func RequireRoles(roles ...string) gin.HandlerFunc: if there is no user in the context respond 401 (same envelope as above); if the role is not in the allowed list abort with 403 and {"error": {"code": "forbidden", "message": "insufficient permissions", "details": ""}}. Otherwise c.Next().
   - func CurrentUser(c *gin.Context) (userID uint, role string, ok bool)
   - Write the error envelope with a small unexported helper inside the middleware package (do not import the handler package).

2. Update the file that defines CORSMiddleware, ONLY if needed: Access-Control-Allow-Headers must include both Content-Type and Authorization. Do not change the other behavior.

3. Check the existing logging middleware: it must not log request headers or bodies. If it does, remove that; if it does not, leave the file unchanged.

Do not modify main.go yet.
```

**Перевір сам:**
Усі негативні відповіді авторизації мають однакове загальне повідомлення (не розкриваємо, чому саме токен недійсний). 401 = «не автентифікований», 403 = «автентифікований, але прав немає» — це різні речі. Без `Authorization` у CORS-заголовках браузерний фронтенд (тема 10) не зможе надсилати токен.

---

### 9.6 — Користувачі, AuthService і ендпоінти автентифікації

```
Add user storage, the authentication service and the auth endpoints.

1. internal/repository/user_repository.go (package repository)
   - UserRepository interface: Create(user *model.User) error, FindByID(id uint) (*model.User, error), FindByEmail(email string) (*model.User, error), List(page, limit int) ([]model.User, error), UpdateRole(id uint, role string) (*model.User, error)
   - gormUserRepository implementation, constructor NewUserRepository(db *gorm.DB) UserRepository
   - FindByEmail compares with lower(email) = lower(?). List: defaults page=1, limit=20, max limit=100, ordered by id. UpdateRole returns ErrNotFound if no rows were affected and returns the reloaded user.
   - All errors pass through translateError (a duplicate email gives ErrConflict).

2. internal/service/auth_service.go (package service)
   - var ErrInvalidCredentials = errors.New("invalid credentials"), var ErrEmailTaken = errors.New("email already registered")
   - type AuthService struct and constructor NewAuthService(users repository.UserRepository, tokens *TokenManager) *AuthService
   - Register(req model.RegisterRequest) (*model.User, error): normalize the email (trim spaces, lower case) and full name (trim), hash the password with HashPassword, ALWAYS set Role = model.RoleReader (never take the role from the request), call users.Create; ErrConflict becomes ErrEmailTaken.
   - Login(email, password string) (*model.AuthResponse, error): normalize the email; FindByEmail; if the user does not exist call CheckPasswordDummy(password) and return ErrInvalidCredentials; if PasswordHash is empty or CheckPassword fails return ErrInvalidCredentials (the same error in all cases); otherwise issue a token and return AuthResponse{AccessToken, TokenType "Bearer", ExpiresIn = int(tokens.TTL().Seconds()), User}.
   - GetByID(id uint) (*model.User, error): passthrough to the repository.
   - EnsureUser(email, password, fullName, role string) error: idempotent bootstrap helper. Validate the role with model.IsValidRole and require a password of at least 8 characters. If a user with this email already exists return nil WITHOUT changing anything (never overwrite an existing password or role). Otherwise create the user with a bcrypt hash.

3. internal/handler/auth_handler.go (package handler)
   - Define a small interface AuthService in this file with the methods Register, Login, GetByID (same signatures as above) and constructor NewAuthHandler(svc AuthService).
   - POST /auth/register: ShouldBindJSON(model.RegisterRequest); 201 with the created user (the password hash is never serialized); 409 if ErrEmailTaken; 422 for validation errors.
   - POST /auth/login: ShouldBindJSON(model.LoginRequest); 200 with AuthResponse; 401 with message "invalid email or password" for ErrInvalidCredentials; 422 for validation errors.
   - GET /me: take the user id with middleware.CurrentUser, load the user via GetByID, return 200; if not found respond 401.
   - Same conventions and the same error-response helper as the other handlers.

Do not modify main.go yet.
```

**Перевір сам:**
Роль при реєстрації завжди `reader`. Невідомий email і неправильний пароль дають **однакову** відповідь 401. `EnsureUser` нічого не перезаписує, якщо користувач уже є.

---

### 9.7 — Правила видачі та розрахунок штрафу (бізнес-логіка студента)

> ⚠️ Цей крок — «межа ШІ». Перед запуском переконайся, що правила нижче — **твої** (див. таблицю рішень). Агент реалізує їх буквально і нічого не додає від себе. Потім ти сам перевіряєш їх тестами в 9.13.

```
Add the loan policy with the fine calculation to internal/service/loan_policy.go (package service). It must be pure logic: no database, no Gin, no time.Now() calls inside the functions (all times are passed as parameters).

type LoanPolicy struct { PeriodDays int; FinePerDay int; FineMax int; MaxActiveLoans int }

func NewLoanPolicy(periodDays, finePerDay, fineMax, maxActiveLoans int) (LoanPolicy, error): return an error if any value is less than or equal to 0, or if fineMax < finePerDay.

func (p LoanPolicy) DueDate(borrowedAt time.Time) time.Time:
  returns borrowedAt (converted to UTC) plus PeriodDays * 24 hours.

func (p LoanPolicy) CalculateFine(dueAt, returnedAt time.Time) int:
  Business rules (implement EXACTLY, do not add anything):
  1. Convert both times to UTC and take only their calendar dates.
  2. overdueDays = number of whole calendar days from the due date to the return date. If the return date is the same as or earlier than the due date, overdueDays = 0 (so the fine is 0).
  3. fine = overdueDays * FinePerDay, but never more than FineMax.
  4. The result is an integer amount in whole UAH. No fractions, no float.

Add a short comment above CalculateFine listing these four rules.
Do not modify any other files.
```

**Перевір сам:**
Функції чисті: час приходить параметром (так їх легко тестувати). Жодного `float` для грошей. Перелік правил у коментарі збігається з таблицею рішень, яку ти прийняв.

---

### 9.8 — Репозиторій видач (транзакції та блокування)

```
Add the loan repository with transactional operations.

1. internal/repository/errors.go (modify the existing file): add package-level errors ErrCopyNotAvailable, ErrLoanLimitReached, ErrAlreadyReturned. Do not change the existing errors or translateError.

2. internal/repository/loan_repository.go (package repository):
   - LoanRepository interface:
     * Issue(copyID, userID uint, borrowedAt, dueAt time.Time, maxActiveLoans int) (*model.Loan, error)
     * Return(loanID uint, returnedAt time.Time, calcFine func(dueAt, returnedAt time.Time) int) (*model.Loan, error)
     * List(status string, page, limit int) ([]model.Loan, error)
     * ListByUser(userID uint, page, limit int) ([]model.Loan, error)
   - gormLoanRepository implementation, constructor NewLoanRepository(db *gorm.DB) LoanRepository
   - Issue, all in ONE transaction:
     a) lock the copy row with clause.Locking{Strength: "UPDATE"}; if it does not exist return ErrInvalidReference;
     b) if the copy status is not model.CopyStatusAvailable return ErrCopyNotAvailable;
     c) lock the user row the same way; if it does not exist return ErrInvalidReference;
     d) count the user's loans WHERE returned_at IS NULL; if the count >= maxActiveLoans return ErrLoanLimitReached;
     e) insert the loan (BorrowedAt, DueAt, FineAmount 0);
     f) update the copy status to "borrowed" and updated_at;
     g) reload the loan with Preload("Copy.Book") and return it.
   - Return, all in ONE transaction: lock the loan row FOR UPDATE (ErrNotFound if missing); if ReturnedAt is already set return ErrAlreadyReturned; set returned_at = returnedAt and fine_amount = calcFine(loan.DueAt, returnedAt); set the copy status back to "available" and updated_at; reload with Preload("Copy.Book") and return.
   - List: status is optional ("" = all). "active" = returned_at IS NULL; "returned" = returned_at IS NOT NULL; "overdue" = returned_at IS NULL AND due_at < now(). Use Preload("Copy.Book") and Preload("User"). Order by borrowed_at DESC, id DESC. Pagination defaults: page=1, limit=20, max limit=100.
   - ListByUser: same pagination and ordering, filtered by user_id, Preload("Copy.Book").
   - All errors pass through translateError. A unique violation from the partial index (race) becomes ErrConflict through translateError.

Do not modify any other files.
```

**Перевір сам:**
`Issue` і `Return` — по одній транзакції з `SELECT ... FOR UPDATE` (дивись SQL у консолі: GORM logger на рівні Info). Видача заблоковує і примірник, і читача — інакше дві одночасні видачі можуть обійти ліміт. У `List` немає циклу запитів по кожній видачі (Preload).

---

### 9.9 — LoanService і ендпоінти видач

```
Add the loan service and the loan endpoints.

1. internal/service/loan_service.go (package service)
   - type LoanService struct { loans repository.LoanRepository; policy LoanPolicy; now func() time.Time } and constructor NewLoanService(loans repository.LoanRepository, policy LoanPolicy) *LoanService (now defaults to time.Now; tests may replace it)
   - Issue(copyID, userID uint) (*model.Loan, error): borrowedAt = now().UTC(); dueAt = policy.DueDate(borrowedAt); call loans.Issue(copyID, userID, borrowedAt, dueAt, policy.MaxActiveLoans)
   - Return(loanID uint) (*model.Loan, error): call loans.Return(loanID, now().UTC(), policy.CalculateFine)
   - List(status string, page, limit int) and ListByUser(userID uint, page, limit int): passthrough to the repository
   - No fine or date arithmetic in this file: it only uses the LoanPolicy methods.

2. internal/handler/loan_handler.go (package handler)
   - Define a small interface LoanService in this file with the four methods above and constructor NewLoanHandler(svc LoanService).
   - POST /loans: ShouldBindJSON(model.CreateLoanRequest); 201 with the loan; ErrInvalidReference -> 422; ErrCopyNotAvailable -> 409 with code "copy_not_available"; ErrLoanLimitReached -> 409 with code "loan_limit_reached"; ErrConflict -> 409; validation errors -> 422.
   - POST /loans/:id/return: 200 with the loan (including fine_amount); ErrNotFound -> 404; ErrAlreadyReturned -> 409 with code "already_returned"; invalid id -> same status as the other handlers use for an invalid id.
   - GET /loans: optional query params status (must pass model.IsValidLoanStatus, otherwise 422), page, limit; 200 with a list.
   - GET /me/loans: loans of the current user (take the user id from middleware.CurrentUser, never from query or body); optional page, limit; 200 with a list.
   - Same conventions and the same error-response helper as the other handlers.

Do not modify main.go yet.
```

**Перевір сам:**
У `GET /me/loans` id користувача береться лише з токена: читач не може подивитися чужі видачі, підставивши інший `user_id`. Арифметики штрафу в сервісі немає — лише виклики `LoanPolicy`.

---

### 9.10 — Керування користувачами (admin)

```
Add admin endpoints for managing users in internal/handler/user_handler.go (package handler).

- NewUserHandler(repo repository.UserRepository)
- GET /users: optional page and limit; 200 with a list of users (password hashes are never serialized).
- PUT /users/:id/role: ShouldBindJSON(model.UpdateUserRoleRequest); 200 with the updated user; ErrNotFound -> 404; invalid id or invalid role -> 422. An admin must not change their own role: if the id equals the id from middleware.CurrentUser respond 403 with code "forbidden" and message "you cannot change your own role" (this also guarantees that the last admin cannot be demoted).
- Same conventions and the same error-response helper as the other handlers.

Do not modify any other files.
```

**Перевір сам:**
Адмін не може змінити власну роль — отже останній адмін не зникне. Хендлер не віддає `password_hash` ні в списку, ні у відповіді на `PUT`.

---

### 9.11 — Підключення в main.go: захист маршрутів за RBAC

```
Update cmd/api/main.go.

Configuration (environment variables, read in main.go):
- JWT_SECRET: required, at least 32 characters; otherwise log a clear error and exit.
- JWT_TTL_MINUTES (default 60), LOAN_PERIOD_DAYS (default 14), FINE_PER_DAY (default 2), FINE_MAX (default 100), MAX_ACTIVE_LOANS (default 5): parse as integers; an invalid value or an error from service.NewLoanPolicy / NewTokenManager must log a clear message and exit.
- ADMIN_EMAIL, ADMIN_PASSWORD, SEED_DEMO_USERS, DEMO_USERS_PASSWORD.
- Never log secrets, passwords or tokens (emails are allowed in logs).

Wiring:
- Create the user repository, loan repository, TokenManager, AuthService, LoanService, and the auth, loan and user handlers.
- After the migrations and the DB connection: if ADMIN_EMAIL and ADMIN_PASSWORD are both set, call authService.EnsureUser(ADMIN_EMAIL, ADMIN_PASSWORD, "Administrator", model.RoleAdmin); if SEED_DEMO_USERS is "true", require DEMO_USERS_PASSWORD and call EnsureUser for librarian@libratrack.local (role librarian) and reader@libratrack.local (role reader). Exit with a clear message on any error.
- Keep: gin.New(), the global middleware order (Recovery, Logging, CORS), GET /health with the DB ping, Swagger UI at /swagger/*any, the package-level swag annotations.

Routes under the group /api/v1 (use the existing handler methods; paths of the old routes must stay exactly as before):
- public: POST /auth/register, POST /auth/login
- authenticated group = api.Group("", middleware.AuthMiddleware(tokenManager)):
  - any role: GET /me, GET /me/loans, GET /books, GET /books/:id, GET /books/:id/copies, GET /books/:id/availability, GET /authors, GET /authors/:id, GET /categories, GET /categories/:id
  - staff group = authenticated.Group("", middleware.RequireRoles(model.RoleLibrarian, model.RoleAdmin)): POST /books, PUT /books/:id, POST /books/:id/copies, POST /authors, POST /categories, POST /loans, POST /loans/:id/return, GET /loans
  - admin group = authenticated.Group("", middleware.RequireRoles(model.RoleAdmin)): DELETE /books/:id, GET /users, PUT /users/:id/role
- GET /health and /swagger/*any stay public and outside /api/v1.

Do not add detailed @Summary/@Param annotations yet.
```

**Перевір сам:**
Запусти: `export $(cat .env | xargs) && go run cmd/api/main.go` (у `.env` не повинно бути пробілів у значеннях). Сценарій через curl:
1. Без токена `GET /api/v1/books` → 401.
2. `POST /api/v1/auth/login` адміном (з `.env`) → отримай `access_token`; `GET /api/v1/me` з `Authorization: Bearer ...` → 200.
3. Зареєструй читача; зайди читачем: `GET /books` → 200, `POST /books` → 403, `GET /users` → 403.
4. Адміном: `PUT /users/{id}/role` {"role":"librarian"} → 200; свою роль змінити → 403.
5. Бібліотекарем: `POST /loans` для доступного примірника → 201, `POST /loans` для того ж примірника → 409; `GET /books/{id}/availability` зменшився.
6. Перевір штраф: `docker exec -it libratrack-db psql -U libratrack -c "UPDATE loans SET borrowed_at = now() - interval '30 days', due_at = now() - interval '16 days' WHERE id = 1;"`, потім `POST /loans/1/return` → `fine_amount` = 32 (16 днів × 2 грн). Повторне повернення → 409.
7. `GET /health` без токена → 200.
Увага: примірники зі статусом `borrowed` із seed-міграції (8.9) не мають записів у `loans`; для демо видавай `available` примірники.

---

### 9.12 — Swagger: схема безпеки та анотації

```
Update only the swag annotations (comments), no code changes.

1. In cmd/api/main.go add package-level annotations:
   // @securityDefinitions.apikey BearerAuth
   // @in header
   // @name Authorization
   // @description Type "Bearer" followed by a space and the JWT token.

2. For the new handlers (auth_handler.go, loan_handler.go, user_handler.go) add swag annotations to every method: @Summary, @Tags, @Accept json, @Produce json, @Param (body or path/query), @Success, @Failure (400/401/403/404/409/422 where applicable), @Router. Add "@Security BearerAuth" to every method except register and login.

3. If the existing book, author, category and copy handlers already have swag annotations, add "@Security BearerAuth" and the @Failure 401 / 403 lines that apply to them. If they have no annotations, leave them untouched.

Do not edit anything under docs/ (it is generated). Do not change any Go code, only comments.
```

**Перевір сам:**
Перегенеруй документацію: `swag init -g cmd/api/main.go`. У Swagger UI (`/swagger/index.html`) з'явилась кнопка **Authorize**; після введення `Bearer <token>` захищені ендпоінти працюють із UI.

---

### 9.13 — Тести без БД: сервіси, middleware, хендлери

```
Write unit tests. Use only the standard testing package plus net/http/httptest, encoding/json and the Gin test mode (no testify, no new dependencies). No test may need a database.

1. internal/service/password_test.go: set bcryptCost = bcrypt.MinCost in an init function of the test file. Cover: hash differs from the plain password; CheckPassword true for the right password and false for a wrong one; CheckPassword with an empty hash is false; HashPassword returns ErrPasswordInvalid for an empty password and for a password longer than 72 bytes.

2. internal/service/token_test.go: Cover: NewTokenManager rejects a short secret and ttl <= 0; Issue then Verify returns the same user id and role; an expired token (replace the unexported now field) gives ErrInvalidToken; a token signed with another secret gives ErrInvalidToken; a tampered token gives ErrInvalidToken; a token with the "none" algorithm (jwt.SigningMethodNone with jwt.UnsafeAllowNoneSignatureType, test only) gives ErrInvalidToken; a token with a wrong issuer gives ErrInvalidToken; a token with an unknown role gives ErrInvalidToken; the token string never appears in the error text.

3. internal/service/loan_policy_test.go: table-driven tests with policy (PeriodDays 14, FinePerDay 2, FineMax 100, MaxActiveLoans 5):
   - DueDate: borrowed 2025-01-01 10:00 UTC -> due 2025-01-15 10:00 UTC
   - CalculateFine, due 2025-01-15 12:00 UTC:
     returned 2025-01-10 09:00 -> 0
     returned 2025-01-15 12:00 (exactly at due) -> 0
     returned 2025-01-15 23:30 (same calendar day) -> 0
     returned 2025-01-16 00:30 (next calendar day) -> 2
     returned 2025-01-20 08:00 -> 10
     returned 2025-03-06 12:00 (50 days) -> 100
     returned 2025-06-01 12:00 (far later, cap) -> 100
   - NewLoanPolicy returns an error for zero values and for fineMax < finePerDay.

4. internal/middleware/auth_test.go: with a hand-written fake TokenVerifier and httptest. Cover: no Authorization header -> 401 with the error envelope and the WWW-Authenticate header; wrong scheme -> 401; token rejected by the verifier -> 401; valid token -> 200 and CurrentUser returns the expected id and role; RequireRoles allows a listed role (200), rejects an unlisted role (403 with code "forbidden"), and returns 401 when no user is in the context.

5. internal/handler/auth_handler_test.go: with a hand-written fake AuthService. Cover: register -> 201 and the JSON has no "password" and no "password_hash"; register with a password shorter than 8 characters -> 422; duplicate email (fake returns service.ErrEmailTaken) -> 409; login -> 200 with access_token; login with ErrInvalidCredentials -> 401; login with an invalid body -> 422.

6. internal/handler/loan_handler_test.go: with a hand-written fake LoanService and the real middleware.AuthMiddleware with a fake verifier. Cover: POST /loans -> 201; copy not available -> 409 with code "copy_not_available"; loan limit -> 409 with code "loan_limit_reached"; POST /loans/:id/return -> 200; already returned -> 409; unknown loan -> 404; GET /loans with an invalid status -> 422; GET /me/loans calls the service with the user id from the token (the fake records it).

Make sure `go test ./...` passes without a database.
```

**Перевір сам:**
`go test ./... -v` — усе PASS. **Додай власноруч (не проси агента) ще 2 граничні випадки у `loan_policy_test.go`** — наприклад, перехід через північ за UTC та рівно на межі ліміту. Якщо ти змінював правила з таблиці рішень, онови числа в тестах: тести — це твоя специфікація, а не підігнана під код.

---

### 9.14 — Інтеграційні тести видач (PostgreSQL)

```
Add integration tests for loans in internal/repository/loan_repository_test.go (package repository). Standard library only.

First, update the existing helper setupTestDB in internal/repository/repository_test.go so that it also truncates the "loans" table (add it to the TRUNCATE ... RESTART IDENTITY CASCADE list). Do not change anything else in that file.

Tests in the new file (skip with t.Skip("TEST_DATABASE_URL is not set") when the variable is empty, using setupTestDB):
- Issue: a successful loan sets the copy status to "borrowed" and returns a loan with due_at and Copy.Book loaded.
- Issue the same copy twice -> ErrCopyNotAvailable.
- Issue with a non-existing copy and with a non-existing user -> ErrInvalidReference.
- Limit: with maxActiveLoans = 2 the third loan of one user returns ErrLoanLimitReached.
- Return: sets returned_at, writes fine_amount from the calcFine function (use a stub function returning a fixed number), sets the copy status back to "available".
- Return twice -> ErrAlreadyReturned; Return of a non-existing loan -> ErrNotFound.
- List with status "active", "returned" and "overdue" (create an overdue loan by inserting a row with due_at in the past directly through GORM) and ListByUser with pagination.
- Concurrency: start 5 goroutines issuing the same available copy to 5 different users at the same time; exactly one must succeed and the others must return ErrCopyNotAvailable or ErrConflict.

All tests must pass with TEST_DATABASE_URL set (also with -race) and be skipped when it is not set.
```

**Перевір сам:**
`TEST_DATABASE_URL=postgres://libratrack:libratrack@localhost:5432/libratrack_test?sslmode=disable go test ./... -v -race`. Без змінної репозиторні тести — SKIP. Тест конкурентності має стабільно проходити: саме тут видно користь від `FOR UPDATE` і часткового унікального індексу.

---

### 9.15 — (Опційно) Refresh-токени: міграція, модель, репозиторій

> Робиться, якщо в курсі розглядаєте refresh-токени. Якщо ні — пропусти 9.15–9.16 і обговори їх лише теоретично (навіщо короткий access-токен, де зберігати refresh, що таке rotation).

```
Add storage for refresh tokens. Create/modify these files:

1. migrations/000004_refresh_tokens.up.sql: table refresh_tokens: id BIGSERIAL PRIMARY KEY, user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE, token_hash VARCHAR(64) NOT NULL UNIQUE, expires_at TIMESTAMPTZ NOT NULL, revoked_at TIMESTAMPTZ NULL, created_at TIMESTAMPTZ NOT NULL DEFAULT now(); index on refresh_tokens(user_id).
   migrations/000004_refresh_tokens.down.sql: DROP TABLE IF EXISTS refresh_tokens.

2. internal/model/refresh_token.go: RefreshToken struct (ID, UserID, TokenHash with json "-", ExpiresAt, RevokedAt *time.Time, CreatedAt) mapped to table "refresh_tokens". Add RefreshRequest { RefreshToken string `json:"refresh_token" binding:"required"` }.
   internal/model/user.go: add RefreshToken (string, json "refresh_token,omitempty") to AuthResponse.

3. internal/repository/refresh_token_repository.go:
   - RefreshTokenRepository interface: Create(t *model.RefreshToken) error, FindByHash(hash string) (*model.RefreshToken, error), Revoke(id uint) error, RevokeAllForUser(userID uint) error
   - gormRefreshTokenRepository, constructor NewRefreshTokenRepository(db *gorm.DB) RefreshTokenRepository
   - FindByHash returns ErrNotFound if missing; all errors pass through translateError.

Do not modify any other files. Only the SHA-256 hash of a refresh token is ever stored, never the token itself.
```

**Перевір сам:**
У БД зберігається лише SHA-256 хеш refresh-токена (як і з паролями — сира цінність ніколи не лежить у базі).

---

### 9.16 — (Опційно) Refresh-токени: сервіс, ендпоінти, підключення

```
Add refresh token rotation.

1. .env.example: add REFRESH_TTL_HOURS=168 and change JWT_TTL_MINUTES to 15.

2. internal/service/auth_service.go (modify): NewAuthService additionally receives a repository.RefreshTokenRepository and a refresh TTL (time.Duration).
   - Helper: generate a refresh token as 32 random bytes from crypto/rand encoded with base64.RawURLEncoding; store only its SHA-256 hex hash (crypto/sha256) in refresh_tokens with expires_at = now + refresh TTL.
   - Login: besides the access token also create a refresh token and return it in AuthResponse.RefreshToken.
   - Refresh(refreshToken string) (*model.AuthResponse, error): hash the value and look it up. Not found or expired -> ErrInvalidCredentials. If the token is already revoked this is a reuse attempt: revoke ALL refresh tokens of that user and return ErrInvalidCredentials. Otherwise revoke the used token (rotation), load the user, issue a new access token and a NEW refresh token, return both.
   - Logout(refreshToken string) error: revoke the token if it exists; unknown tokens are ignored (return nil).

3. internal/handler/auth_handler.go (modify): add POST /auth/refresh (RefreshRequest -> 200 AuthResponse, 401 for ErrInvalidCredentials, 422 for validation errors) and POST /auth/logout (RefreshRequest -> 204). Extend the AuthService interface in the handler file accordingly.

4. cmd/api/main.go (modify): read REFRESH_TTL_HOURS (default 168), create the refresh token repository, pass it to NewAuthService, register POST /auth/refresh and POST /auth/logout as public routes under /api/v1.

5. Update the existing tests only where the changed constructor signature or AuthResponse requires it, and add a service-level test with fake repositories: rotation returns a new pair, reuse of an old refresh token revokes all tokens of the user.

Never log refresh tokens.
```

**Перевір сам:**
Login повертає пару `access_token` (15 хв) + `refresh_token` (7 діб). `POST /auth/refresh` видає нову пару, а старий refresh більше не працює; повторне використання старого — 401 і відкликання всіх токенів користувача. Обговори: чому access-токен короткий, де клієнт зберігає refresh (cookie `HttpOnly` vs localStorage) — це тема 10.

---

### 9.17 — Фінальна перевірка та дрібні виправлення

```
Review the whole project and fix any remaining issues so that:

- go mod tidy works
- go vet ./... reports no problems
- go build ./... succeeds
- go test ./... passes (repository tests are skipped when TEST_DATABASE_URL is not set, and pass when it is set)
- docker compose up -d starts PostgreSQL and the app starts against it; migrations 000001-000003 (and 000004 if refresh tokens were added) are applied automatically
- the app refuses to start when JWT_SECRET is missing or shorter than 32 characters
- GET /health works without a token; every /api/v1 route except register, login (and refresh/logout if present) returns 401 without a token
- the role matrix is enforced exactly: reader cannot create or change books, loans or users; librarian cannot delete books or manage users; admin can do everything
- there are no secrets, passwords or tokens in Go code, in logs or in error messages
- there are no AutoMigrate calls and no non-ASCII characters in any .go file

Make only the minimal necessary changes.
```

**Перевір сам:**
`go test ./...` і повний ручний сценарій з 9.11. Окремо переглянь консоль застосунку під час логіну: у логах не має бути паролів, токенів, заголовка `Authorization`.

---

### 9.18 — Перевірка готового коду (фінальний аудит)

```
Perform a final audit of the entire libratrack project after theme 9.

Check and report (do not rewrite working code unless something is broken):

1. Directory structure is unchanged (cmd/api, internal/model, handler, service, repository, middleware) plus migrations/, docker-compose.yml, .env.example. No new packages under internal/.
2. go.mod contains github.com/golang-jwt/jwt/v5 and golang.org/x/crypto in addition to the theme 8 dependencies.
3. Passwords: stored only as bcrypt hashes; User.PasswordHash has json:"-"; registration enforces min 8 and max 72 characters; the password never appears in a response, log or error.
4. Registration always creates role "reader"; the role cannot be set through POST /auth/register.
5. Login returns the same 401 error for an unknown email, a wrong password and an empty password hash, and calls CheckPasswordDummy for unknown emails.
6. JWT: HS256 only (Verify pins the method and the issuer and requires exp); the secret comes from JWT_SECRET (at least 32 characters, the app fails fast otherwise); the token has a short TTL; the token string is never logged.
7. AuthMiddleware returns 401 and RequireRoles returns 403 in the standard error envelope; CORS allows the Authorization header; the logging middleware does not log headers or bodies.
8. The route matrix in main.go matches the table in the course (public / any authenticated / librarian+admin / admin only), and the old route paths are unchanged.
9. GET /me/loans uses the user id from the token only; an admin cannot change their own role.
10. Migrations 000003 (and 000004 if present) have .up.sql and .down.sql; loans has FKs with ON DELETE RESTRICT, CHECK constraints, the partial unique index (one active loan per copy) and the case-insensitive unique email index; AutoMigrate is not used anywhere.
11. Loan Issue and Return each run in a single transaction with row locking (FOR UPDATE); the fine is calculated only by LoanPolicy.CalculateFine; money is an int, never a float; the policy values come from environment variables.
12. EnsureUser never overwrites an existing user; demo users are created only when SEED_DEMO_USERS=true; no credentials are hardcoded in Go code or in migrations.
13. (If refresh tokens were added) only SHA-256 hashes are stored; rotation revokes the used token; reuse of a revoked token revokes all tokens of the user.
14. Tests exist for password, token, loan policy, middleware, auth and loan handlers (no DB needed) and for loan repository (skipped without TEST_DATABASE_URL); go test ./... passes.
15. Swagger has the BearerAuth security definition and the protected endpoints carry @Security BearerAuth.
16. No non-UTF8 or non-ASCII characters in any .go file; .env is in .gitignore.

If everything is correct - reply with a short confirmation.
If something is broken - make the minimal fix and show only the changed files.
```

**Перевір сам:**
Фінальна перевірка всього проєкту. Закоміть і поставте тег `theme-9` — це стартова точка для теми 10 (веб-інтерфейс).

---

### Для обговорення на занятті (компроміси, які варто проговорити)

- **Роль у JWT** береться з токена, тому зміна ролі адміном починає діяти лише після завершення терміну дії access-токена. Альтернатива — читати користувача з БД на кожен запит (свіжі права, але додатковий запит).
- **Чому 401 ≠ 403** і чому однакова відповідь на «невідомий email» та «неправильний пароль».
- **Чому штраф і ліміти — не для ШІ:** правила видно в одному місці (`loan_policy.go`), покриті твоїми тестами, і їх можна змінити, не чіпаючи HTTP чи БД.
- **Що поза межами теми 9** (hardening перед продакшеном): rate limiting для `/auth/login`, блокування акаунту після N невдалих спроб, HTTPS, ротація `JWT_SECRET`.
