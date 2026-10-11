## ТЕМА 10. Full-Stack Go: Web Application with Template Engine

Мета блоку: дати LibraTrack «обличчя» — серверний веб-інтерфейс на `html/template` без SPA-фреймворків: каталог із пошуком і фільтрами для читачів, сторінка книги з доступністю, «Мої видачі», панель бібліотекаря (книги, автори, категорії, примірники, видачі/повернення, звіти), керування користувачами для адміна. Автентифікація в браузері — JWT з теми 9 в `HttpOnly`-cookie + CSRF-захист. Наприкінці — Docker, README, ER-діаграма й CHANGELOG, щоб проєкт був придатний як портфоліо. **Структура `internal/`, REST API `/api/v1` і правила з тем 8–9 не змінюються** — веб-шар додається поруч і викликає ті самі репозиторії та сервіси (не ходить у власний API по HTTP).

**Загальні правила для всіх кроків:**
- Output ONLY the exact file contents in the required Aider format.
- No explanations, no extra text, no markdown outside the code blocks.
- Do not invent extra files or change the directory structure, except the files explicitly listed in the step.
- Use only UTF-8. Never insert non-ASCII characters into .go files (error messages, comments and identifiers must be in English). User-facing Ukrainian text lives ONLY in web/templates/*.html and web/messages.json.
- Keep code simple and correct for Go 1.22+. Use only the standard library plus Gin and the dependencies already in go.mod. No new dependencies.
- Keep the layers: cmd/api, internal/model, internal/handler, internal/service, internal/repository, internal/middleware. Do not create new internal/ packages. The only new top-level Go package is web/ (embedded templates and static files).
- Web handlers only parse forms, call existing services/repositories and render templates. Business rules stay in internal/service.
- Templates use html/template only. Never use template.HTML, template.JS, template.URL or similar on user data. No inline <script>, inline event handlers (onclick=...) or style="..." attributes (the Content-Security-Policy forbids them). No JavaScript at all, no CDN links, no external fonts.
- Every state-changing form is a POST with a hidden csrf_token field and ends with a redirect (303 See Other, Post/Redirect/Get).
- Never log or render passwords, hashes, tokens or cookies.

**Перед стартом (робить студент, не агент):**
- Проєкт з теми 9 збирається, тести проходять (`go build ./... && go test ./...`), є тег `theme-9`; `docker compose up -d` — БД healthy.
- У `.env` постав `JWT_TTL_MINUTES=60` або більше: сесія в браузері живе стільки ж, скільки JWT (refresh-токени з 9.15–9.16 веб-шар не використовує).
- Відкрий `SEED_DEMO_USERS=true` у `.env` — матимеш готових `admin`, `librarian@libratrack.local` і `reader@libratrack.local` для ручних перевірок.
- Додавай в Aider (`/add`) файли, які крок змінює.
- **Прийми UX-рішення (таблиця нижче) до кроку 10.6.**

**UX-рішення студента (межа ШІ):** що бачить читач, а що бібліотекар, і де стоїть кнопка повернення — це продуктові рішення. Нижче — значення, закладені в промпти; зміни їх свідомо, до запуску відповідних кроків.

| Питання | Рішення в промптах |
|---|---|
| Каталог для анонімів | закритий: потрібен вхід (так само, як API у темі 9) |
| Що робить читач на сторінці книги | бачить кількість доступних примірників і підказку «звернись до бібліотекаря» (самообслуговування й резервування — стретч-фіча) |
| Хто бачить таблицю примірників | лише бібліотекар і адмін |
| Як видати книгу | форма на сторінці книги: вибір доступного примірника + email читача |
| Де кнопка «Прийняти повернення» | у списку активних видач у панелі бібліотекаря |
| Прострочення | червоний бейдж + поточний штраф за правилами `LoanPolicy` |
| Розмір сторінки каталогу | 12 книг |
| Мова інтерфейсу | українська (тексти в шаблонах і `web/messages.json`) |

**Карта веб-маршрутів (перевір перед кроком 10.12):**

| Група | Маршрути | Доступ |
|---|---|---|
| Публічні | `GET /` (редирект), `GET/POST /login`, `GET/POST /register`, `/static/*` | усі |
| Автентифіковані | `GET /catalog`, `GET /books/:id`, `GET /my/loans`, `POST /logout` | reader, librarian, admin |
| Бібліотекар | `GET /staff/books`, `GET /staff/new-book`, `POST /staff/books`, `GET+POST /staff/books/:id/edit`, `POST /staff/books/:id/copies`, `GET+POST /staff/authors`, `GET+POST /staff/categories`, `GET /staff/loans`, `POST /staff/loans`, `POST /staff/loans/:id/return`, `GET /staff/reports` | librarian, admin |
| Адмін | `POST /staff/books/:id/delete`, `GET /admin/users`, `POST /admin/users/:id/role` | admin |

`/health`, `/swagger/*` і `/api/v1/*` залишаються як у темі 9. Cookie **не** автентифікує API-запити: API приймає лише `Authorization: Bearer`.

---

### 10.1 — Каркас: embed, повідомлення, CSS, базовий layout

```
Add the scaffolding for the server-rendered web UI. Create/modify these files:

1. .env.example (modify): keep all existing variables and add
   COOKIE_SECURE=false
   with a comment line above it: set to true when the app is served over HTTPS. Also add a comment above JWT_TTL_MINUTES: the browser session lives as long as the JWT, use 60 or more for the web UI.

2. web/embed.go (package web):
   - //go:embed templates/*.html static messages.json  into  var FS embed.FS
   - func Templates() (fs.FS, error) returning fs.Sub(FS, "templates")
   - func Static() (fs.FS, error) returning fs.Sub(FS, "static")
   - func Messages() ([]byte, error) returning the content of messages.json

3. web/messages.json: one flat JSON object (UTF-8, short Ukrainian texts without HTML) with exactly these keys:
   flash.logged_out, flash.registered, flash.book_created, flash.book_updated, flash.book_deleted, flash.author_created, flash.category_created, flash.copy_created, flash.loan_issued, flash.loan_returned, flash.role_updated,
   error.invalid_credentials, error.email_taken, error.validation, error.not_found, error.forbidden, error.server, error.conflict, error.csrf, error.invalid_reference, error.copy_not_available, error.loan_limit_reached, error.already_returned, error.reader_not_found, error.self_role_change, error.delete_not_confirmed,
   field.email, field.password, field.full_name, field.title, field.isbn, field.published_year, field.publisher, field.description, field.category_id, field.author_ids, field.name, field.bio, field.inventory_no, field.status, field.copy_id, field.user_id, field.role, field.unknown
   Keys starting with "error." are shown as errors, keys starting with "flash." as success messages. Field texts are short labels (for example "Email", "Password").

4. web/static/css/style.css: a minimal responsive stylesheet: system font stack, container with max-width 960px, top navigation bar, tables (wrapped by a .table-wrap with overflow-x: auto), forms (labels, inputs, selects, textarea, buttons), .flash and .flash-error, .badge, .badge-ok, .badge-warn, .badge-overdue, .pagination, .card, .muted. No external fonts, no CDN, no JavaScript, no frameworks.

5. web/templates/layout.html: define the template "layout" (a complete HTML5 page, lang="uk", charset utf-8, viewport meta, <title>{{template "title" .}} - LibraTrack</title>, stylesheet link /static/css/style.css). The page data contract is:
   .User (nil for anonymous, otherwise .User.ID, .User.Role, .User.IsStaff, .User.IsAdmin), .Flash (a message key or empty string), .FlashIsError (bool), .CSRFToken (string), .Data (page specific).
   Navigation: link to /catalog and /my/loans for any logged-in user; for staff the links /staff/books, /staff/authors, /staff/categories, /staff/loans, /staff/reports; for admin also /admin/users; a logout form (POST /logout with the hidden csrf_token field) for logged-in users; links to /login and /register for anonymous users. Show the user's role by printing the string {{.User.Role}} directly.
   If .Flash is not empty render <div class="flash {{if .FlashIsError}}flash-error{{end}}">{{msg .Flash}}</div>. Then {{template "content" .}} inside <main class="container">.
   Every page template must define the blocks "title" and "content". The template function msg(key string) string is provided by the renderer (step 10.3).

Do not modify any other files. No JavaScript anywhere.
```

**Перевір сам:**
`web/messages.json` — валідний JSON з усіма ключами (`python -m json.tool web/messages.json`). У CSS і шаблонах немає зовнішніх URL. `go build ./...` проходить (пакет `web` лише вбудовує файли).

---

### 10.2 — Веб-middleware: cookie-автентифікація, CSRF, заголовки безпеки

```
Add web middleware to internal/middleware/web.go (package middleware). Reuse the existing context keys and the TokenVerifier interface from auth.go so that middleware.CurrentUser works for web requests too.

- const AuthCookieName = "libratrack_token", const csrfCookieName = "csrf_token", const csrfFormField = "csrf_token".
- func SetAuthCookie(c *gin.Context, token string, maxAgeSeconds int, secure bool) and func ClearAuthCookie(c *gin.Context, secure bool): use http.SetCookie with Path "/", HttpOnly true, SameSite Lax, Secure = secure.
- func WebAuthMiddleware(v TokenVerifier) gin.HandlerFunc: read the cookie AuthCookieName; if it is missing or the token fails verification, clear the cookie, respond with c.Redirect(http.StatusSeeOther, "/login") and abort. On success store user id and role in the context (same keys as AuthMiddleware) and call c.Next(). It must NOT read the Authorization header.
- func WebOptionalAuth(v TokenVerifier) gin.HandlerFunc: same, but never redirects or aborts: if the cookie is valid the user is stored in the context, otherwise the request continues anonymously.
- func WebRequireRoles(onForbidden gin.HandlerFunc, roles ...string) gin.HandlerFunc: if the role from the context is not in the list, call onForbidden(c) and abort (the caller renders the 403 page).
- func CSRFMiddleware(secure bool, onFail gin.HandlerFunc) gin.HandlerFunc (double-submit cookie): read the cookie csrf_token; if it is missing or not 43 characters long, generate 32 random bytes with crypto/rand, encode with base64.RawURLEncoding and set the cookie (Path "/", HttpOnly, SameSite Lax, Secure = secure). Store the token in the Gin context. For the methods POST, PUT, PATCH and DELETE compare c.PostForm(csrfFormField) with the cookie value using subtle.ConstantTimeCompare; if they differ (or the field is empty) call onFail(c) and abort.
- func CSRFToken(c *gin.Context) string returns the token stored in the context.
- func SecurityHeaders() gin.HandlerFunc: set Content-Security-Policy "default-src 'self'; img-src 'self' data:; base-uri 'none'; form-action 'self'; frame-ancestors 'none'", X-Content-Type-Options "nosniff", X-Frame-Options "DENY", Referrer-Policy "same-origin".

Do not modify any other files.
```

**Перевір сам:**
Cookie автентифікації й CSRF мають `HttpOnly` і `SameSite=Lax`. `WebAuthMiddleware` не читає заголовок `Authorization`, а `AuthMiddleware` з теми 9 не читає cookie — це дві різні моделі автентифікації, і змішувати їх не можна (інакше API стане вразливим до CSRF).

---

### 10.3 — Рендерер шаблонів, flash-повідомлення, сторінка помилки

```
Add the template renderer.

1. internal/handler/web_render.go (package handler):
   - type PageUser struct { ID uint; Role string; IsStaff bool; IsAdmin bool }
   - type PageData struct { User *PageUser; Flash string; FlashIsError bool; CSRFToken string; Data any }
   - type FormErrors struct { ErrorKey string; FieldKeys []string } (a message key such as "error.conflict" and a list of field names such as "title", to be embedded into page data structs)
   - type ErrorPage struct { Status int; MessageKey string }
   - type Renderer struct { pages map[string]*template.Template; messages map[string]string; secure bool }
   - func NewRenderer(tplFS fs.FS, messagesJSON []byte, secure bool) (*Renderer, error): parse the JSON into the messages map; for every *.html file in tplFS except layout.html parse template.New("layout").Funcs(funcs).ParseFS(tplFS, "layout.html", <file>) and store it under the file name (for example "catalog.html"). Return an error on any parse problem (the app must fail at startup, not at the first request). Template functions: msg(key string) string (returns the key itself if unknown), date(t time.Time) string with layout "02.01.2006", datetime(t time.Time) string with layout "02.01.2006 15:04", dateptr(t *time.Time) string ("" for nil), add(a, b int) int, sub(a, b int) int, seq(from, to int) []int, hasID(id uint, ids []uint) bool, authorNames(authors []model.Author) string (comma-separated full names).
   - func (r *Renderer) Render(c *gin.Context, status int, page string, data any): build PageData (User from middleware.CurrentUser, CSRFToken from middleware.CSRFToken(c), the flash described below), execute the template "layout" into a bytes.Buffer, and only on success write it with c.Data(status, "text/html; charset=utf-8", buf.Bytes()). On an execution error log the error text (never the data) and respond with a plain-text 500 "internal server error". An unknown page name is the same 500.
   - Flash: func (r *Renderer) SetFlash(c *gin.Context, key string) sets a cookie "flash" (value = key, only if the key exists in messages, Path "/", HttpOnly, SameSite Lax, MaxAge 60, Secure = r.secure). Render reads the cookie, uses it only if the key exists in messages, sets FlashIsError when the key starts with "error.", and deletes the cookie (MaxAge -1).
   - func (r *Renderer) Error(c *gin.Context, status int, messageKey string): Render the page "error.html" with ErrorPage data.
   - func FieldKeys(err error) []string: if err is a validator.ValidationErrors (github.com/go-playground/validator/v10, already a Gin dependency) map each FieldError.Field() through this explicit map and return the unique results: Email=email, Password=password, FullName=full_name, Title=title, ISBN=isbn, PublishedYear=published_year, Publisher=publisher, Description=description, CategoryID=category_id, AuthorIDs=author_ids, Name=name, Bio=bio, InventoryNo=inventory_no, Status=status, CopyID=copy_id, UserID=user_id, Role=role; any other field becomes "unknown".

2. web/templates/error.html: define "title" (use the HTTP status number only, for example {{.Data.Status}}) and "content": a short block with the status code, the text {{msg .Data.MessageKey}} and a link to /catalog.

Do not modify any other files.
```

**Перевір сам:**
Шаблони парсяться при старті (`NewRenderer`), а не під час першого запиту. Відповідь пишеться в буфер і лише потім у `ResponseWriter` — користувач не отримає «пів-сторінки» при помилці шаблону. `msg` повертає ключ, якщо переклад відсутній (помилка видна, але не падає).

---

### 10.4 — Сторінки входу, реєстрації та виходу

```
Add the web authentication pages.

1. web/templates/login.html and web/templates/register.html. Each defines "title" and "content". Page data: .Data.Email (login) or .Data.Email and .Data.FullName (register), plus the embedded FormErrors fields .Data.ErrorKey and .Data.FieldKeys. Show the error as {{if .Data.ErrorKey}}<p class="flash flash-error">{{msg .Data.ErrorKey}}</p>{{end}} and each failed field as {{range .Data.FieldKeys}}...{{msg (printf "field.%s" .)}}...{{end}}. Forms: method="post", hidden input csrf_token with {{.CSRFToken}}, labeled inputs with the proper type and autocomplete attributes (email, current-password / new-password), required, and for the registration password minlength="8" maxlength="72". Never put the password into a value attribute.

2. internal/handler/web_auth_handler.go (package handler):
   - NewWebAuthHandler(r *Renderer, auth AuthService, secure bool), using the AuthService interface that already exists in auth_handler.go.
   - GET /login: if middleware.CurrentUser reports a logged-in user redirect (303) to /catalog; otherwise render login.html.
   - POST /login: read email and password with c.PostForm; call auth.Login; on service.ErrInvalidCredentials re-render login.html with status 401, ErrorKey "error.invalid_credentials" and the email kept; on any other error render r.Error with 500 and "error.server"; on success call middleware.SetAuthCookie(c, resp.AccessToken, resp.ExpiresIn, secure) and redirect 303 to /catalog. There is no "next" parameter (no open redirects).
   - GET /register and POST /register: build a model.RegisterRequest from c.PostForm (email, password, full_name), validate it with binding.Validator.ValidateStruct (github.com/gin-gonic/gin/binding); on validation errors re-render with status 422, ErrorKey "error.validation", FieldKeys from FieldKeys(err) and the email and full name kept (never the password); on service.ErrEmailTaken re-render with status 409 and "error.email_taken"; on success SetFlash "flash.registered" and redirect 303 to /login.
   - POST /logout: middleware.ClearAuthCookie, SetFlash "flash.logged_out", redirect 303 to /login.

Do not modify main.go yet.
```

**Перевір сам:**
Після невдалого входу сторінка не показує, чи існує такий email (та сама відповідь 401, що й у API). Поле пароля ніколи не повертається в розмітку. Немає параметра `next` — отже, немає open redirect.

---

### 10.5 — Репозиторій каталогу: пошук, фільтри, популярні книги

```
Add catalog search and a report query. Create these files:

1. internal/model/catalog.go (package model):
   - CatalogFilter: Query (string), CategoryID (uint), AuthorID (uint), OnlyAvailable (bool), Page (int), Limit (int)
   - CatalogItem: Book (Book), TotalCopies (int), AvailableCopies (int)
   - CatalogPage: Items ([]CatalogItem), Total (int64), Page (int), Limit (int), TotalPages (int)
   - PopularBook: BookID (uint, gorm column book_id), Title (string), LoanCount (int, gorm column loan_count)

2. internal/repository/catalog_repository.go (package repository):
   - CatalogRepository interface: Search(f model.CatalogFilter) (*model.CatalogPage, error) and PopularBooks(limit int) ([]model.PopularBook, error)
   - gormCatalogRepository, constructor NewCatalogRepository(db *gorm.DB) CatalogRepository; all errors pass through translateError.
   - Search: defaults page=1, limit=12, max limit=50; trim Query and cut it to 100 characters. Escape the characters \, % and _ in Query with a backslash before building the pattern "%" + escaped + "%". Conditions (all combined with AND, ALL values passed as bound parameters, never concatenated into the SQL string):
     * Query: books.title ILIKE ? OR books.isbn ILIKE ? OR books.id IN (SELECT ba.book_id FROM book_authors ba JOIN authors a ON a.id = ba.author_id WHERE a.full_name ILIKE ?)
     * CategoryID > 0: books.category_id = ?
     * AuthorID > 0: books.id IN (SELECT book_id FROM book_authors WHERE author_id = ?)
     * OnlyAvailable: EXISTS (SELECT 1 FROM copies c WHERE c.book_id = books.id AND c.status = 'available')
     Count the total with the same conditions, then load the page ordered by title, id with Preload("Authors") and Preload("Category"). Then fill TotalCopies and AvailableCopies for all books of the page with ONE aggregate query: SELECT book_id, COUNT(*) AS total, COUNT(*) FILTER (WHERE status = 'available') AS available FROM copies WHERE book_id IN ? GROUP BY book_id (books without copies get 0). No per-book queries.
   - PopularBooks: one query: SELECT b.id AS book_id, b.title, COUNT(l.id) AS loan_count FROM loans l JOIN copies c ON c.id = l.copy_id JOIN books b ON b.id = c.book_id GROUP BY b.id, b.title ORDER BY loan_count DESC, b.title LIMIT ?; default limit 10, max 50.

Do not modify any other files.
```

**Перевір сам:**
У `Search` немає конкатенації користувацького тексту в SQL (захист від SQL injection) і немає запитів у циклі (N+1 уже розбирали в темі 8): книги + `Preload` + один агрегат. Символи `%` і `_` у пошуковому рядку екрануються, тож пошук `100%` шукає буквально «100%».

---

### 10.6 — Каталог і сторінка книги

```
Add the catalog pages.

1. web/templates/catalog.html. Defines "title" and "content". Data: .Data.Query, .Data.CategoryID, .Data.AuthorID, .Data.OnlyAvailable, .Data.Result (*model.CatalogPage), .Data.Categories ([]model.Category), .Data.Authors ([]model.Author), .Data.PrevURL and .Data.NextURL (strings, empty if there is no such page). A GET form (action="/catalog") with a text input q, a select category_id, a select author_id, a checkbox available (value 1) and a submit button, all pre-filled. Results as a table or a list of cards: title (link to /books/{id}), authors (authorNames), category name, publisher/year, and a badge: "available X of Y" (badge-ok when X > 0, otherwise badge-warn). Empty state text when there are no results. Pagination links using PrevURL and NextURL plus "page N of M".

2. web/templates/book.html. Defines "title" and "content". Data: .Data.Book (*model.Book), .Data.Total, .Data.Available (ints), .Data.AvailableCopies ([]model.Copy). Show title, authors, category, ISBN, publisher, year and the description (print it as plain text inside an element that has the CSS class "prewrap"; add white-space: pre-line for .prewrap to style.css). For readers (not .User.IsStaff): the availability badge and a short hint that the book is issued by the librarian. For staff additionally: the table of all copies (inventory number, status), an "Edit" link to /staff/books/{id}/edit, and, only if .Data.AvailableCopies is not empty, an issue form: method="post" action="/staff/loans", hidden csrf_token, hidden book_id, a select copy_id listing the available copies by inventory number, an email input reader_email (required) and a submit button.

3. internal/handler/web_catalog_handler.go (package handler):
   - NewWebCatalogHandler(r *Renderer, catalog repository.CatalogRepository, books repository.BookRepository, authors repository.AuthorRepository, categories repository.CategoryRepository)
   - GET /catalog: read the query parameters q, category_id, author_id, available (only the value "1" means true) and page; a non-numeric or negative number is treated as unset, a missing page is 1. Call catalog.Search with Limit 12, load the categories and authors for the filter selects with their existing List methods (if List is paginated, request a large page size), build PrevURL and NextURL with net/url url.Values (keep the filters, set page) and render catalog.html.
   - GET /books/:id: an invalid id gives r.Error(404, "error.not_found"); load the book with books.FindByID (which preloads authors, category and copies); repository.ErrNotFound gives the same 404 page; other errors give 500 "error.server". Compute Total, Available and AvailableCopies from the already loaded Copies (no extra queries) and render book.html.

Do not modify main.go yet.
```

**Перевір сам:**
Опис книги виводиться як звичайний текст (екранується автоматично), а не як HTML. Форму видачі бачить лише бібліотекар/адмін, і тільки якщо є доступні примірники. Перенос рядків забезпечує CSS `white-space: pre-line`, а не `template.HTML`.

---

### 10.7 — «Мої видачі» (та поточний штраф)

```
Add the "My loans" page.

1. internal/service/loan_service.go (modify): add the method ProjectedFine(l *model.Loan) int. If l.ReturnedAt is not nil return l.FineAmount; otherwise return s.policy.CalculateFine(l.DueAt, s.now().UTC()). Do not duplicate any fine arithmetic.

2. web/templates/my_loans.html. Defines "title" and "content". Data: .Data.Rows (a slice of LoanRow with fields Loan model.Loan, Overdue bool, Fine int), .Data.PrevURL, .Data.NextURL. A table: book title (link to /books/{Loan.Copy.Book.ID}), borrowed date, due date, return date (dateptr), status badge (returned = badge-ok, overdue = badge-overdue, otherwise badge-warn "active"), fine in UAH (show the Fine column only as a number). Empty state when there are no loans.

3. internal/handler/web_loans_handler.go (package handler):
   - Define a small interface webLoanService in this file with the methods ListByUser(userID uint, page, limit int) ([]model.Loan, error), List(status string, page, limit int) ([]model.Loan, error), Issue(copyID, userID uint) (*model.Loan, error), Return(loanID uint) (*model.Loan, error) and ProjectedFine(l *model.Loan) int.
   - type LoanRow struct { Loan model.Loan; Overdue bool; Fine int }
   - NewWebLoansHandler(r *Renderer, loans webLoanService) (more dependencies are added in later steps).
   - GET /my/loans: the user id comes ONLY from middleware.CurrentUser (never from query or form); page param (default 1), limit 20; build the rows in Go: Overdue = ReturnedAt == nil and time.Now().UTC() is after DueAt, Fine = loans.ProjectedFine; render my_loans.html. The template contains no business logic.

Do not modify main.go yet.
```

**Перевір сам:**
Штраф у веб-шарі рахується тим самим `LoanPolicy.CalculateFine` через `ProjectedFine` — формула існує в одному місці. Читач бачить лише власні видачі: id береться з токена.

---

### 10.8 — Панель бібліотекаря: книги та примірники

```
Add the staff pages for books and copies.

1. web/templates/staff_books.html: "title" and "content". Data: .Data.Books ([]model.Book), .Data.PrevURL, .Data.NextURL. A "New book" link to /staff/new-book; a table with title, ISBN, category, authors, an "Edit" link, and (only if .User.IsAdmin) a delete form: method="post" action="/staff/books/{id}/delete", hidden csrf_token, a required checkbox confirm with value "yes" and the label "Yes, delete", and a submit button.

2. web/templates/staff_book_form.html: "title" and "content" for both creating and editing. Data fields: Mode ("create" or "edit"), BookID uint, Title, ISBN, Publisher, Description, PublishedYear (strings, so the submitted text can be shown again), CategoryID uint, SelectedAuthorIDs []uint, Categories []model.Category, Authors []model.Author, Copies []model.Copy, plus the embedded FormErrors (.Data.ErrorKey, .Data.FieldKeys). Form: method="post", action "/staff/books" for create or "/staff/books/{id}/edit" for edit, hidden csrf_token; inputs title, isbn, publisher, published_year (type number), textarea description, select category_id, a multiple select author_ids (use hasID to mark selected options). In edit mode also show the copies table and a second form: method="post" action="/staff/books/{id}/copies" with hidden csrf_token, input inventory_no, select status (available, borrowed, lost, under_repair) and a submit button.

3. internal/handler/web_staff_handler.go (package handler): NewWebStaffHandler(r *Renderer, books repository.BookRepository, authors repository.AuthorRepository, categories repository.CategoryRepository, copies repository.CopyRepository). Check the exact method signatures in internal/repository/book_repository.go and copy_repository.go and use them as they are.
   - GET /staff/books: page param, limit 20, books.List without a category filter, render staff_books.html.
   - GET /staff/new-book: empty form (Mode "create").
   - POST /staff/books: build a model.CreateBookRequest from c.PostForm (title, isbn, publisher, description, published_year as int where an empty string means 0 and a non-number is a validation error, category_id as uint, author_ids with c.PostFormArray as []uint); validate with binding.Validator.ValidateStruct. Validation error: re-render the form with status 422, ErrorKey "error.validation", FieldKeys from FieldKeys(err) and all submitted values kept. repository.ErrConflict (duplicate ISBN): status 409 and "error.conflict". repository.ErrInvalidReference: 422 and "error.invalid_reference". Success: SetFlash "flash.book_created" and redirect 303 to /staff/books.
   - GET /staff/books/:id/edit and POST /staff/books/:id/edit: same logic with model.UpdateBookRequest; invalid id or ErrNotFound gives the 404 page; success SetFlash "flash.book_updated" and redirect to /staff/books.
   - POST /staff/books/:id/delete: if c.PostForm("confirm") != "yes" SetFlash "error.delete_not_confirmed" and redirect to /staff/books WITHOUT calling the repository; otherwise books.Delete (ErrNotFound gives the 404 page), SetFlash "flash.book_deleted", redirect 303 to /staff/books.
   - POST /staff/books/:id/copies: build a model.CreateCopyRequest (inventory_no, status; an empty status means "available"), validate; book id from the path; ErrNotFound gives the 404 page; ErrConflict (duplicate inventory number) re-renders the edit form with status 409 and "error.conflict"; success SetFlash "flash.copy_created" and redirect 303 to /staff/books/{id}/edit.

Do not modify main.go yet.
```

**Перевір сам:**
Видалення — лише POST із підтвердженням (чекбокс; JavaScript заборонений CSP), а маршрут у кроці 10.12 буде тільки для адміна. Після кожного POST — редирект 303 (оновлення сторінки не повторює дію).

---

### 10.9 — Панель бібліотекаря: автори та категорії

```
Add the staff pages for authors and categories.

1. web/templates/staff_authors.html and web/templates/staff_categories.html: "title" and "content". Each shows a table of existing items (authors: full name and bio; categories: name) and below it a creation form: method="post" action="/staff/authors" (inputs full_name and a textarea bio) or action="/staff/categories" (input name), hidden csrf_token, and the error block from FormErrors (.Data.ErrorKey, .Data.FieldKeys rendered with msg and printf "field.%s"). Data: .Data.Items plus the submitted values (FullName, Bio or Name) and the embedded FormErrors.

2. internal/handler/web_staff_handler.go (modify): add the methods
   - GET /staff/authors and POST /staff/authors: build model.CreateAuthorRequest from c.PostForm (full_name, bio), validate with binding.Validator.ValidateStruct; validation error: re-render with 422, "error.validation" and FieldKeys; success: authors.Create, SetFlash "flash.author_created", redirect 303 to /staff/authors.
   - GET /staff/categories and POST /staff/categories: model.CreateCategoryRequest (name); repository.ErrConflict (duplicate name): re-render with 409 and "error.conflict"; success: SetFlash "flash.category_created", redirect 303 to /staff/categories.
   - Use the existing List methods of the repositories to show the items (if they paginate, request a large page size).

Do not modify any other files.
```

**Перевір сам:**
Дублікат назви категорії повертає ту саму форму з повідомленням, а не 500. Усі поля, що користувач ввів, зберігаються в формі після помилки.

---

### 10.10 — Панель бібліотекаря: видачі, повернення, звіти

```
Add the staff pages for loans and reports.

1. web/templates/staff_loans.html: "title" and "content". Data: .Data.Status ("active", "overdue" or "returned"), .Data.Rows (slice of LoanRow), .Data.PrevURL, .Data.NextURL. Three links/tabs to /staff/loans?status=active, ?status=overdue, ?status=returned. A table: reader (Loan.User.FullName and email), book title (Loan.Copy.Book.Title), copy inventory number, borrowed, due, returned (dateptr), badge (overdue = badge-overdue), fine. For rows where Loan.ReturnedAt is nil add a return form: method="post" action="/staff/loans/{id}/return", hidden csrf_token, submit button "Accept return".

2. web/templates/staff_reports.html: "title" and "content". Data: .Data.Popular ([]model.PopularBook) shown as a table (title linked to /books/{BookID}, loan count) and .Data.Overdue (slice of OverdueReader with fields FullName, Email, Count int, TotalFine int) shown as a table. Empty state texts for both.

3. internal/handler/web_loans_handler.go (modify): NewWebLoansHandler additionally receives users repository.UserRepository and catalog repository.CatalogRepository. Add:
   - type OverdueReader struct { FullName string; Email string; Count int; TotalFine int }
   - GET /staff/loans: query status (one of active, overdue, returned; empty means active; anything else gives 422 via r.Error with "error.validation"), page, limit 20; loans.List(status, page, limit); build rows (Overdue and Fine the same way as in /my/loans) and render staff_loans.html.
   - POST /staff/loans (issue): read copy_id, book_id and reader_email from c.PostForm. Normalize the email (trim, lower case) and look the reader up with users.FindByEmail; if not found SetFlash "error.reader_not_found" and redirect 303 back to /books/{book_id}. Build model.CreateLoanRequest{CopyID, UserID}, validate it with binding.Validator.ValidateStruct, call loans.Issue. Error mapping with SetFlash and a redirect 303 back to /books/{book_id}: repository.ErrCopyNotAvailable -> "error.copy_not_available", repository.ErrLoanLimitReached -> "error.loan_limit_reached", repository.ErrInvalidReference -> "error.invalid_reference", validation -> "error.validation"; any other error -> r.Error 500. Build the redirect URL ONLY from a book_id parsed with strconv.ParseUint (never redirect to a user supplied string); if book_id is invalid redirect to /catalog. Success: SetFlash "flash.loan_issued" and redirect 303 to /staff/loans.
   - POST /staff/loans/:id/return: loans.Return(id); ErrNotFound -> 404 page; ErrAlreadyReturned -> SetFlash "error.already_returned" and redirect 303 to /staff/loans; success: SetFlash "flash.loan_returned" and redirect 303 to /staff/loans?status=returned.
   - GET /staff/reports: catalog.PopularBooks(10); loans.List("overdue", 1, 100); group the overdue loans by reader in Go (map by user id), summing Count and TotalFine (ProjectedFine), sort by Count descending, then by FullName; render staff_reports.html.

Do not modify main.go yet.
```

**Перевір сам:**
Редирект після помилки видачі будується лише з `book_id`, розібраного як число (немає open redirect). Звіт «читачі з простроченням» групує в Go; сума штрафу береться через `ProjectedFine`, формули у веб-шарі немає.

---

### 10.11 — Адмін: керування користувачами

```
Add the admin users page.

1. web/templates/admin_users.html: "title" and "content". Data: .Data.Users ([]model.User), .Data.CurrentUserID (uint), .Data.PrevURL, .Data.NextURL. A table: full name, email, created date (date), role. For every row except the current user: a form method="post" action="/admin/users/{id}/role" with hidden csrf_token, a select role (admin, librarian, reader; the current role selected) and a submit button. For the current user's row show only the role text (no form).

2. internal/handler/web_users_handler.go (package handler): NewWebUsersHandler(r *Renderer, users repository.UserRepository).
   - GET /admin/users: page param (default 1), limit 20, users.List, render admin_users.html (CurrentUserID from middleware.CurrentUser).
   - POST /admin/users/:id/role: invalid id gives the 404 page; if the id equals the current user id SetFlash "error.self_role_change" and redirect 303 to /admin/users without calling the repository; build model.UpdateUserRoleRequest from c.PostForm("role") and validate it with binding.Validator.ValidateStruct (error: SetFlash "error.validation" and redirect); users.UpdateRole (ErrNotFound gives the 404 page); success SetFlash "flash.role_updated" and redirect 303 to /admin/users.

Do not modify any other files.
```

**Перевір сам:**
Як і в API, адмін не може змінити власну роль. Хендлер не віддає в шаблон `password_hash` (його немає в JSON, а шаблон його й не використовує).

---

### 10.12 — Підключення в main.go: веб-маршрути та групи доступу

```
Update cmd/api/main.go.

Configuration and setup:
- Read COOKIE_SECURE (parse with strconv.ParseBool, default false; an invalid value logs a clear message and exits).
- Load the embedded templates and messages: web.Templates(), web.Messages(), then handler.NewRenderer(tplFS, messages, cookieSecure). Exit with a clear log message on any error (the app must not start with broken templates).
- Create the CatalogRepository and all web handlers (WebAuthHandler, WebCatalogHandler, WebLoansHandler, WebStaffHandler, WebUsersHandler), reusing the repositories and services that already exist. Do NOT call the REST API from the web handlers.

Keep unchanged: gin.New(), the global middleware order (Recovery, Logging, CORS), GET /health with the DB ping, Swagger UI at /swagger/*any, ALL /api/v1 routes with their Bearer middleware and role groups, the package-level swag annotations, the startup admin/demo user bootstrap.

Web routing (the security headers and CSRF middleware are applied ONLY to the web groups below, never globally, because the Content-Security-Policy would break Swagger UI):
- forbidden := func(c *gin.Context) { renderer.Error(c, http.StatusForbidden, "error.forbidden") }, csrfFail := func(c *gin.Context) { renderer.Error(c, http.StatusForbidden, "error.csrf") }
- web := r.Group("", middleware.SecurityHeaders(), middleware.CSRFMiddleware(cookieSecure, csrfFail))
- public := web.Group("", middleware.WebOptionalAuth(tokenManager)): GET / (303 redirect to /catalog), GET and POST /login, GET and POST /register
- authed := web.Group("", middleware.WebAuthMiddleware(tokenManager)): GET /catalog, GET /books/:id, GET /my/loans, POST /logout
- staff := authed.Group("", middleware.WebRequireRoles(forbidden, model.RoleLibrarian, model.RoleAdmin)): GET /staff/books, GET /staff/new-book, POST /staff/books, GET and POST /staff/books/:id/edit, POST /staff/books/:id/copies, GET and POST /staff/authors, GET and POST /staff/categories, GET /staff/loans, POST /staff/loans, POST /staff/loans/:id/return, GET /staff/reports
- admin := authed.Group("", middleware.WebRequireRoles(forbidden, model.RoleAdmin)): POST /staff/books/:id/delete, GET /admin/users, POST /admin/users/:id/role
- Static files: serve web.Static() under /static through a group that uses only middleware.SecurityHeaders(), for example r.Group("/static", middleware.SecurityHeaders()).StaticFS("/", http.FS(staticFS)).
- r.NoRoute: if the path starts with /api/ respond with the JSON envelope {"error": {"code": "not_found", "message": "route not found", "details": ""}} and status 404; otherwise render the HTML 404 page with renderer.Error(c, 404, "error.not_found") (run it through the SecurityHeaders middleware).
```

**Перевір сам:**
`export $(cat .env | xargs) && go run cmd/api/main.go`, відкрий `http://localhost:8080/`. Перевір: анонім → `/catalog` перенаправляє на `/login`; вхід `reader@libratrack.local` → каталог; читач на `/staff/books` бачить 403-сторінку; `/swagger/index.html` працює (CSP його не зачіпає); `curl -i localhost:8080/api/v1/books` без токена → 401 JSON; **той самий запит із cookie, але без `Authorization` → теж 401** (cookie не діє на API). `curl -i localhost:8080/nonexistent` → HTML 404, `/api/v1/nonexistent` → JSON 404.

---

### 10.13 — (Опційно) Доопрацювання вигляду

> UX-рішення — твої. Цей крок змінює лише стилі й розмітку, не логіку. Пропусти, якщо вигляд уже влаштовує.

```
Improve the look of the web UI. Modify only web/static/css/style.css and, where needed, the class attributes in web/templates/*.html. Do not change any Go code, form fields, form actions, template blocks or message keys.

Requirements:
- CSS custom properties for colors and spacing on :root; automatic dark theme through @media (prefers-color-scheme: dark).
- Readable typography (system font stack, comfortable line-height), visible keyboard focus outlines on links, inputs and buttons.
- Sticky top navigation that wraps on narrow screens; tables scroll horizontally inside .table-wrap on phones; the catalog uses a card grid on wide screens and a single column on narrow ones.
- Distinct colors for .badge-ok, .badge-warn and .badge-overdue with enough contrast; destructive buttons (delete) visually distinct from normal ones.
- No JavaScript, no external fonts or CDN resources, no inline styles.
```

**Перевір сам:**
Перевір у DevTools на ширині 375px і в темній темі системи. `grep -rn "style=" web/templates` — порожньо (інакше CSP заблокує).

---

### 10.14 — Тести без БД: шаблони, middleware, хендлери

```
Write tests that need no database. Use only the standard testing package plus net/http/httptest, net/url and the Gin test mode (no testify, no new dependencies). Define NEW fakes with the prefix webFake in a new file internal/handler/web_fakes_test.go so they do not clash with fakes that already exist in other test files of the package; reuse nothing by name from them.

1. internal/handler/web_render_test.go (package handler): build the renderer from the real embedded files (web.Templates(), web.Messages()).
   - Every template file parses (NewRenderer returns no error).
   - Every key listed in step 10.1 exists in messages.json and has a non-empty text.
   - No template contains an inline <script>, an "onclick=" or "onsubmit=" attribute, or a style=" attribute (scan the embedded files with strings.Contains).
   - Render every page ("login.html", "register.html", "error.html", "catalog.html", "book.html", "my_loans.html", "staff_books.html", "staff_book_form.html", "staff_authors.html", "staff_categories.html", "staff_loans.html", "staff_reports.html", "admin_users.html") with representative data and a staff user in the context: status 200 and no 500.
   - XSS check: render book.html with the title "<script>alert(1)</script>" and a description containing "<img src=x onerror=alert(1)>": the body must contain "&lt;script&gt;" and must not contain "<script>alert(1)" or "<img src=x".
   - Forms contain a hidden csrf_token input (check login.html, staff_authors.html, staff_loans.html).

2. internal/middleware/web_test.go (package middleware) with a fake TokenVerifier:
   - WebAuthMiddleware: no cookie gives 303 to /login; an invalid token gives 303 and clears the cookie; a valid cookie gives 200 and CurrentUser returns the expected id and role.
   - AuthMiddleware (the API middleware) with only a valid cookie and no Authorization header gives 401 (cookies must not authenticate API requests).
   - WebRequireRoles calls onForbidden for a disallowed role and passes an allowed one.
   - CSRFMiddleware: a GET sets a csrf_token cookie with HttpOnly and SameSite Lax; a POST without the form field calls onFail; a POST with a different token calls onFail; a POST whose form field equals the cookie passes.
   - SetAuthCookie sets HttpOnly and SameSite Lax; SecurityHeaders sets the four headers.

3. internal/handler/web_auth_handler_test.go: with a fake AuthService and a real renderer and middleware (the CSRF cookie and field are supplied in the requests): POST /login success gives 303 to /catalog and a Set-Cookie for libratrack_token with HttpOnly; wrong credentials give 401 and the body contains no password and no hash; POST /register with a short password gives 422 and the response body does not contain the submitted password; duplicate email gives 409; successful registration gives 303 to /login; POST /logout clears the cookie.

4. internal/handler/web_catalog_handler_test.go: with a fake CatalogRepository that records the filter it receives and a fake BookRepository: GET /catalog parses q, category_id, author_id, available=1 and page correctly, treats "page=abc" as page 1; GET /books/abc gives 404; an unknown book gives the 404 page; a reader's book page does not contain action="/staff/loans" while a librarian's does (when the book has an available copy).

5. internal/handler/web_staff_handler_test.go and web_loans_handler_test.go: with fakes: POST /staff/books with an empty title gives 422 and keeps the submitted values; duplicate ISBN (fake returns repository.ErrConflict) gives 409; success gives 303 and a flash cookie; POST /staff/books/:id/delete without confirm=yes does NOT call the repository; POST /staff/loans with an unknown reader email gives 303 to /books/{book_id} with the flash cookie "error.reader_not_found"; ErrCopyNotAvailable and ErrLoanLimitReached map to their flash keys; a book_id that is not a number redirects to /catalog; POST /staff/loans/:id/return with ErrAlreadyReturned maps to the flash key; GET /my/loans passes the user id from the token to ListByUser.

Make sure `go test ./...` passes without a database.
```

**Перевір сам:**
`go test ./... -v` — усе PASS. Тест на XSS і тест «cookie не автентифікує API» — найцінніші в цьому наборі: вони фіксують дві ключові властивості безпеки. Додай власноруч ще один сценарій: форма без `csrf_token` повертає 403-сторінку.

---

### 10.15 — Інтеграційні тести каталогу (PostgreSQL)

```
Add integration tests for the catalog in internal/repository/catalog_repository_test.go (package repository). Standard library only. Skip with t.Skip("TEST_DATABASE_URL is not set") when the variable is empty and use the existing setupTestDB helper (it already truncates all tables including loans).

Tests:
- Search by Query matches the title case-insensitively, the ISBN, and an author's full name.
- A Query containing % or _ is treated literally: create a book titled "100% Go" and another titled "1000 Go": the Query "100%" returns only the first.
- Filters CategoryID and AuthorID (a book with two authors is found by either).
- OnlyAvailable excludes a book whose copies are all borrowed or lost, and excludes a book without copies.
- TotalCopies and AvailableCopies are correct for books with mixed copy statuses, and 0/0 for a book without copies.
- Pagination: with 5 books and Limit 2 the pages have 2, 2 and 1 items, Total is 5 and TotalPages is 3; the order is by title.
- PopularBooks: create loans (through direct GORM inserts) so that one book has 3 loans, another 1 and a third none: the result is ordered by loan count, contains only books with loans, and the limit parameter is respected.

All tests must pass with TEST_DATABASE_URL set (also with -race) and be skipped when it is not set.
```

**Перевір сам:**
`TEST_DATABASE_URL=postgres://libratrack:libratrack@localhost:5432/libratrack_test?sslmode=disable go test ./... -v -race`. Без змінної репозиторні тести — SKIP.

---

### 10.16 — Упаковка для портфоліо: Docker, README, ER-діаграма, CHANGELOG

```
Package the project so that a reviewer can run it with a few commands. Create/modify these files:

1. Dockerfile (multi-stage):
   - Builder stage: golang image whose version matches the "go" directive in go.mod (alpine variant); copy go.mod and go.sum first and run go mod download, then copy the sources and build with CGO_ENABLED=0 go build -o /out/libratrack ./cmd/api (templates, static files and swagger docs are already embedded in the binary).
   - Runtime stage: alpine; create and use a non-root user; WORKDIR /app; copy the binary and the migrations/ directory (to /app/migrations); ENV MIGRATIONS_PATH=/app/migrations; EXPOSE 8080; CMD ["/app/libratrack"].
2. .dockerignore: .git, .env, .idea, .vscode, .aider*, bin, *.exe, *.out
3. docker-compose.yml (modify): keep the "db" service unchanged and add a service "app": build ".", container_name libratrack-app, depends_on db with condition service_healthy, ports "8080:8080", environment (values are interpolated from the .env file by docker compose): PORT=8080, DATABASE_URL=postgres://libratrack:libratrack@db:5432/libratrack?sslmode=disable, MIGRATIONS_PATH=/app/migrations, ALLOWED_ORIGIN, JWT_SECRET=${JWT_SECRET:?set JWT_SECRET in .env}, JWT_TTL_MINUTES=${JWT_TTL_MINUTES:-60}, ADMIN_EMAIL=${ADMIN_EMAIL:-}, ADMIN_PASSWORD=${ADMIN_PASSWORD:-}, SEED_DEMO_USERS=${SEED_DEMO_USERS:-false}, DEMO_USERS_PASSWORD=${DEMO_USERS_PASSWORD:-}, LOAN_PERIOD_DAYS, FINE_PER_DAY, FINE_MAX, MAX_ACTIVE_LOANS (each with the same default as in .env.example), COOKIE_SECURE=${COOKIE_SECURE:-false}.
4. README.md (in English): the product description (a library catalog and lending system for a school or college); features by role (reader, librarian, admin); technology table; quick start in three commands (cp .env.example .env, edit the secrets in .env, docker compose up --build) with the URLs /catalog and /swagger/index.html; how to enable the demo accounts (SEED_DEMO_USERS=true; the demo password comes from DEMO_USERS_PASSWORD in .env, do not write any password in the README); a short demo scenario (reader searches a book, librarian issues it, reader sees it in My loans, librarian accepts the return, an overdue loan and its fine); the access matrix table for the web UI and for the REST API; how to run the tests (with and without TEST_DATABASE_URL); the project structure; screenshot placeholders referencing docs/screenshots/*.png with a note that they must be added; and an ER diagram as a Mermaid erDiagram block generated from migrations/*.sql (all tables, keys and relationships).
5. CHANGELOG.md (Keep a Changelog style, English) with one section per theme: 7 (REST API with Gin, in-memory storage, Swagger), 8 (PostgreSQL, GORM, migrations, relations, copies, availability), 9 (JWT, bcrypt, RBAC, loans and fines, optional refresh tokens), 10 (server-rendered web UI, CSRF, Docker, README).

Do not modify any Go source files.
```

**Перевір сам:**
З чистого клона: `cp .env.example .env`, заміни секрети, `docker compose up --build` → відкрий `http://localhost:8080/`. Зауваж: `swag init` переписує лише `docs.go`, `swagger.json`, `swagger.yaml`; скріншоти тримай у `docs/screenshots/`. Додай реальні скріншоти в README — без них портфоліо виглядає порожнім.

---

### 10.17 — Фінальна перевірка та дрібні виправлення

```
Review the whole project and fix any remaining issues so that:

- go mod tidy works, go vet ./... reports no problems, go build ./... succeeds
- go test ./... passes (repository tests are skipped without TEST_DATABASE_URL and pass with it)
- docker compose up --build starts the database and the app; migrations are applied automatically; the app serves the web UI on port 8080
- the app refuses to start if any embedded template fails to parse
- every web page works for its role and returns the 403 page for the other roles, exactly as in the course route map
- all web forms that change state are POST with a csrf_token and end with a 303 redirect
- the REST API under /api/v1 behaves exactly as after theme 9 and is not authenticated by cookies
- no template contains inline scripts, inline event handlers or style attributes; no external resources are loaded
- there are no passwords, hashes, tokens or secrets in Go code, logs or rendered pages
- there are no non-ASCII characters in any .go file

Make only the minimal necessary changes.
```

**Перевір сам (фінальна демонстрація = «приймання» наскрізного проєкту):**
1. Читач заходить → шукає книгу (за назвою, автором, категорією, лише доступні) → бачить кількість вільних примірників.
2. Бібліотекар входить → відкриває сторінку книги → видає примірник за email читача.
3. Читач оновлює «Мої видачі» → бачить видачу й дату повернення.
4. Імітація прострочення: `docker exec -it libratrack-db psql -U libratrack -c "UPDATE loans SET borrowed_at = now() - interval '30 days', due_at = now() - interval '16 days' WHERE id = 1;"` → у «Мої видачі» читач бачить прострочення й поточний штраф; у звітах бібліотекаря — читача з простроченням.
5. Бібліотекар приймає повернення → штраф зафіксовано, примірник знову доступний.
6. Адмін змінює роль користувача на `/admin/users`.
Закоміть і постав тег `theme-10`.

---

### 10.18 — Перевірка готового коду (фінальний аудит)

```
Perform a final audit of the entire libratrack project after theme 10.

Check and report (do not rewrite working code unless something is broken):

1. Directory structure: cmd/api, internal/model, handler, service, repository, middleware, migrations/, plus the top-level web/ package (templates, static, messages.json, embed.go), Dockerfile, docker-compose.yml, README.md, CHANGELOG.md. No new packages under internal/.
2. go.mod has no new dependencies compared with theme 9 (the web layer uses html/template, embed and the existing Gin).
3. Web handlers call repositories and services directly, not the REST API; business rules (fines, limits) exist only in internal/service; web handlers contain no fine arithmetic.
4. Authentication in the browser: the JWT is stored only in an HttpOnly, SameSite=Lax cookie (Secure when COOKIE_SECURE=true); no token is ever rendered into HTML; WebAuthMiddleware reads only the cookie and AuthMiddleware reads only the Authorization header.
5. CSRF: every POST route in the web groups is protected by CSRFMiddleware with a constant-time comparison; every form in the templates has a hidden csrf_token input; the REST API is not covered by it (it does not use cookies).
6. SecurityHeaders (CSP, nosniff, frame deny, referrer policy) are applied to the web groups and static files but not to /swagger and /api; no template violates the CSP (no inline script, no on* attributes, no style attributes, no external resources).
7. html/template is used everywhere; no template.HTML, template.JS or template.URL on user data; user input is never concatenated into SQL (catalog search uses bound parameters and escapes LIKE wildcards); redirects are built only from fixed paths or numeric ids (no open redirect).
8. The route map in main.go matches the course table (public / authenticated / librarian+admin / admin only); unauthorized users get a redirect to /login, forbidden users get the 403 page; NoRoute returns JSON for /api/ paths and HTML otherwise.
9. Catalog search runs a constant number of queries per request (count, page with Preload, one aggregate for copy counts); no per-book queries in loops.
10. All user-facing texts are in templates or messages.json; Go code contains only message keys; flash cookies accept only known keys.
11. Every state-changing web action follows Post/Redirect/Get; validation errors re-render the form with status 422 and keep the submitted values (never the password).
12. Tests: template/renderer tests (including the XSS check and the CSP scan), middleware tests (including "cookie does not authenticate the API"), web handler tests (no DB) and catalog repository tests (skipped without TEST_DATABASE_URL); go test ./... passes.
13. docker compose up --build works from a clean clone; README contains the three-command quick start, the access matrices, the Mermaid ER diagram and the screenshot placeholders; CHANGELOG has sections for themes 7-10; .env is in .gitignore and .dockerignore.
14. No non-UTF8 or non-ASCII characters in any .go file.

If everything is correct - reply with a short confirmation.
If something is broken - make the minimal fix and show only the changed files.
```

**Перевір сам:**
Фінальна перевірка всього проєкту. Закоміть і постав тег `theme-10`. Додай у README реальні скріншоти й, за бажанням, коротке демо-відео або GIF.

---

### Для обговорення на занятті (компроміси, які варто проговорити)

- **JWT у `HttpOnly`-cookie чи в `localStorage`.** Cookie недоступна JavaScript (захист від викрадення токена через XSS), але потребує CSRF-захисту; `localStorage` не потребує CSRF, але токен викрадається будь-яким XSS. Тут обрано cookie + `SameSite=Lax` + double-submit CSRF-токен.
- **Три шари захисту від XSS:** автоекранування `html/template`, відсутність `template.HTML` на користувацьких даних, Content-Security-Policy без inline-скриптів. Спробуй свідомо зламати: підстав `<script>` в назву книги й подивись на результат.
- **Post/Redirect/Get:** чому після кожного POST — 303, а не 200 із сторінкою.
- **Чому веб-шар не ходить у власний REST API:** обидва «фронти» (JSON API і HTML) викликають ті самі сервіси — правила штрафів і ліміти існують в одному місці (це результат архітектурної дисципліни з теми 7).
- **Сесія й refresh-токени:** веб-сесія живе стільки ж, скільки access-токен. Для реальної експлуатації додають refresh-cookie або серверні сесії; у темі 10 цього свідомо немає.
- **Чого немає в навігації:** імені користувача (у JWT лише id і роль; щоб показати ім'я, потрібен запит до БД або додаткова cookie) — продуктове рішення.
- **Обмеження підходу без JavaScript:** підтвердження видалення — чекбокс, пошук — повне перезавантаження. Природне продовження: htmx для часткових оновлень без SPA.
- **Що поза межами курсу (hardening перед продакшеном):** rate limiting на `/login`, HTTPS і `COOKIE_SECURE=true`, бекапи БД, резервування книг (черга), нагадування про повернення, повнотекстовий пошук.
