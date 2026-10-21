# Workspaces Backend

Бэкенд web-IDE на Go. Пользователь выбирает набор пакетов nixpkgs, бэкенд собирает под него Docker-образ, запускает контейнер и отдаёт его терминал в браузер по WebSocket. Рядом хранятся проекты и файлы пользователей.

## Содержание

- [Быстрый старт](#быстрый-старт)
- [Архитектура](#архитектура)
- [Структура репозитория](#структура-репозитория)
- [Конфигурация](#конфигурация)
- [Платформа](#платформа)
- [Как устроен модуль и как добавить новый](#как-устроен-модуль-и-как-добавить-новый)
- [Модули](#модули)
  - [projects](#модуль-projects)
  - [users](#модуль-users)
  - [Шаблон для нового модуля](#шаблон-раздела-для-нового-модуля)
- [Работа с БД](#работа-с-бд)
- [Файлы проектов на диске](#файлы-проектов-на-диске)
- [Разработка](#разработка)
- [Решение проблем](#решение-проблем)

## Быстрый старт

### Требования

- Go (версия из `go.mod`)
- Docker (доступный текущему пользователю: сокет по умолчанию или `DOCKER_HOST`)
- PostgreSQL (запускается в отдельном Docker-контейнере)
- `make`

### 1. PostgreSQL

Пример (подставьте свои значения и те же самые в `.env`):

```bash
docker run -d --name workspaces-db \
  -e POSTGRES_USER=<user> \
  -e POSTGRES_PASSWORD=<password> \
  -e POSTGRES_DB=<db> \
  -p 5432:5432 \
  postgres
```

### 2. Конфигурация

Скопируйте пример и при необходимости поправьте значения (все переменные обязательны, подробности в разделе [Конфигурация](#конфигурация)):

```bash
cp .env.example .env
```

Значения в примере (`postgres` / `password`) подходят для локальной разработки. Если вы поднимали Postgres командой выше с другими `POSTGRES_*`, укажите те же в `.env`.

### 3. Запуск

```bash
make run
```

Цель выполняет `go mod tidy`, собирает `bin/backend` из `./cmd/` и запускает его. Все маршруты доступны под префиксом `/api/v1`, например `http://localhost:8000/api/v1/proj/1`.

Если Postgres недоступен, приложение завершится с паникой при старте: подключение проверяется сразу (`Ping`).

## Архитектура

Приложение собрано из модулей. Каждый модуль реализует интерфейс `platform.Module`, получает общие зависимости (`module.Deps`: конфиг и подключение к БД) и сам регистрирует свои маршруты. `cmd/main.go` только создаёт модули и монтирует их под `/api/v1`.

```
HTTP (chi + httplog + cors)
        │
        ▼
     handler ──► service ──► repository ──► PostgreSQL (go-pg)
                    │
                    ├──► util.Docker ──► Docker Engine API (сборка образов, терминал)
                    └──► PROJECT_PATH (файлы проектов на диске)
```

Слои внутри модуля:

| Слой | Ответственность |
|---|---|
| `handler` | HTTP и WebSocket: разбор запроса, формирование ответа |
| `service` | бизнес-логика, оркестрация репозиториев и Docker |
| `repository` | доступ к БД (go-pg) |
| `models` | модели данных, тела запросов и их валидация |
| `custom_errors` | ошибки модуля и их соответствие HTTP-статусам |
| `util` | вспомогательный код, в том числе клиент Docker |

## Структура репозитория

```
.
├── cmd/
│   └── main.go                  # точка входа: конфиг, БД, модули, HTTP-сервер
├── internal/
│   └── platform/                # общая платформа для всех модулей
│       ├── config/              # загрузка конфигурации из окружения и .env
│       ├── database/            # подключение к PostgreSQL
│       ├── httpserver/          # роутер, логирование, CORS, запуск, формат ошибок
│       └── module/              # интерфейс Module и Deps
├── projects/                    # модуль projects
│   ├── projects.go              # публичный вход: New(deps) platform.Module
│   └── internal/                # реализация модуля (скрыта от других пакетов)
│       ├── custom_errors/
│       ├── handler/             # project, file, workspace (WebSocket)
│       ├── models/
│       ├── module/              # сборка модуля: module.go, routes.go, schema.go
│       ├── repository/
│       ├── service/
│       └── util/                # docker.go, terminal.go, hash.go, query.go, flake.go
├── scripts/                     # build_image.sh, run_container.sh (скрипты для удобной сборки и запуска в проде)
├── Dockerfile                   # Dockerfile для сборки образа приложения
├── Makefile
├── tests/
└── tmp/                         # файлы проектов пользователей (PROJECT_PATH)
```

## Конфигурация

Конфиг загружается в `internal/platform/config`: сначала читается `.env` (если файл есть), затем переменные окружения. **Все переменные обязательны**: если какая-то не задана, приложение завершится с ошибкой `<NAME> is not set`.

| Переменная | Описание |
|---|---|
| `API_PORT` | порт HTTP-сервера |
| `DATABASE_HOST` | хост PostgreSQL |
| `DATABASE_PORT` | порт PostgreSQL |
| `DATABASE_USER` | пользователь PostgreSQL |
| `DATABASE_PASSWORD` | пароль PostgreSQL |
| `DATABASE_DB` | имя базы данных |
| `PROJECT_PATH` | каталог, где хранятся файлы проектов пользователей |
| `DOCKER_HOST` | (необязательно) адрес Docker Engine; читается клиентом Docker из окружения, по умолчанию локальный сокет |

Значения, которые пока заданы в коде и которые стоит вынести в конфиг: ревизия nixpkgs, TTL workspace (8 часов), лимиты контейнера (512 МБ памяти, 256 процессов), список разрешённых origin для WebSocket.

## Платформа

Общий код в `internal/platform`.

**HTTP-сервер (`httpserver`)**

- Роутер на chi. Логирование запросов через `httplog` (slog, текстовый формат, схема ECS), паники в обработчиках перехватываются. Имя сервиса в логах задаётся в `NewRouter(...)`.
- CORS: разрешены любые `http://*` и `https://*` origin, методы `GET`, `POST`, `PUT`, `DELETE`, `OPTIONS`, заголовки `Accept`, `Authorization`, `Content-Type`, `X-CSRF-Token`.
- `Run` блокируется до отмены контекста (SIGINT или SIGTERM) и затем останавливает сервер корректно, с таймаутом 10 секунд. `ReadHeaderTimeout` 5 секунд.

**Формат ошибок**

Все ошибки API имеют одинаковый вид, и по HTTP, и по WebSocket:

```json
{ "message": "Failed to get project: ...", "timestamp": "2026-10-13T00:41:58.472625+03:00" }
```

- HTTP: `httpserver.ResponseError(w, message, status)`.
- WebSocket (после апгрейда писать в `http.ResponseWriter` нельзя): `httpserver.WSError(conn, message)`.

**Соответствие ошибок и статусов** (`projects/internal/custom_errors`, функция `StatusFromError`):

| Ошибка | HTTP-статус |
|---|---|
| `ErrProjectNotFound` | 400 |
| `ErrProjectAlreadyExists` | 409 |
| `ErrInvalidProject` | 400 |
| `ErrInvalidFile` | 400 |
| `ErrFileExists` | 400 |
| `ErrInvalidQuery` | 400 |
| `io.ErrUnexpectedEOF` | 400 |
| `ErrServerError` и всё остальное | 500 |

## Как устроен модуль и как добавить новый

Все модули строятся по одной схеме, поэтому новый модуль делается копированием структуры существующего.

**Контракт.** Модуль реализует `platform.Module` (см. `internal/platform/module/module.go`):

- `Name() string`: имя модуля, пишется в лог при регистрации.
- `Routes(r chi.Router)`: регистрирует маршруты модуля. `main.go` вызывает его внутри `r.Route("/api/v1", ...)`, то есть пути в модуле указываются без префикса.

Зависимости (`module.Deps`): `Cfg *config.Config` и `DB *pg.DB`.

**Раскладка пакетов модуля `<name>`:**

```
<name>/
├── <name>.go                    # публичный вход: func New(deps platform.Deps) platform.Module
└── internal/
    ├── custom_errors/           # ошибки модуля и StatusFromError
    ├── handler/                 # HTTP-обработчики
    ├── models/                  # модели и тела запросов (Validate)
    ├── module/
    │   ├── module.go            # struct Module, New: сборка repo -> service -> handler
    │   ├── routes.go            # (m *Module) Routes
    │   └── schema.go            # createSchema: создание таблиц
    ├── repository/
    ├── service/
    └── util/
```

Реализация лежит в `internal/`, поэтому другие модули не могут импортировать её напрямую: наружу виден только `<name>.New`.

**Шаги:**

1. Создайте каталог `<name>/` со структурой выше.
2. В `internal/module/module.go` создайте репозитории, сервисы и обработчики. В `New` вызовите `createSchema(deps.DB)` и **проверьте возвращаемую ошибку**.
3. В `internal/module/schema.go` перечислите модели для `CreateTable` в порядке зависимостей: сначала таблицы, на которые ссылаются внешние ключи.
4. В `internal/module/routes.go` зарегистрируйте маршруты.
5. В `<name>/<name>.go` добавьте `New(deps)`, возвращающий `platform.Module`.
6. Добавьте модуль в список `modules` в `cmd/main.go`.

## Модули

### Модуль projects

Проекты, файлы проектов и workspace (запуск окружения).

**Маршруты** (префикс `/api/v1`):

| Метод | Путь | Тело / параметры | Ответ |
|---|---|---|---|
| GET | `/proj/{id}` | `id` в пути | проект, 200 |
| POST | `/proj` | JSON `CreateProjectRequest` | созданный проект, 201 |
| GET | `/files?file_id=<id>` | query-параметр `file_id` | файл, 200 |
| GET | `/files/{project_id}` | `project_id` в пути | список файлов проекта, 200 |
| POST | `/files` | JSON `{"project_id": 1, "filename": "main.rs"}` | созданный файл, 201 |
| PUT | `/files` | JSON `{"id": 1, "text": "..."}` | обновлённый файл, 200 |
| GET | `/ws` | WebSocket | см. ниже |

Ошибки возвращаются в [общем формате](#платформа). Поля проекта и `CreateProjectRequest`: по `models/project.go`.

**Файл** (`File`):

```json
{ "id": 1, "project_id": 1, "path": "src/main.rs", "text": "fn main() {}" }
```

**Workspace.** Основная функциональность модуля.

1. Клиент подключается к `/api/v1/ws` и первым сообщением отправляет `{"user_id": 1, "project_id": 1, "dependencies": ["python313"]}`.
2. Бэкенд собирает образ на базе `nixos/nix` с указанными пакетами nixpkgs и присылает статусы (`building`, `starting`, `ready`) и лог сборки.
3. После `ready` сокет работает как терминал контейнера: бинарные кадры это ввод и вывод, текстовый JSON `{"cols": 120, "rows": 30}` меняет размер терминала.
4. Закрытие сокета останавливает и удаляет контейнер.

Подробный протокол для фронтенда: документ «Workspaces: WebSocket API для фронтенда» (TODO: добавить ссылку или перенести в `docs/workspaces-ws.md`).

**Образы.** Тег образа детерминирован: хеш от базового образа, ревизии nixpkgs и отсортированного списка пакетов. Одинаковый набор зависимостей даёт тот же образ, повторная сборка берётся из кеша слоёв. Ревизия nixpkgs зафиксирована в коде, поэтому окружение воспроизводимо.

### Модуль users

Планируется. Раздел заполняется по мере разработки.

- Назначение: Регистрация и авторизация пользователей, генерация JWT токенов
- Маршруты: TODO
- Модели и таблицы: TODO
- Связь с `projects`: сейчас `user_id` в запросе на workspace приходит от клиента, без проверки.
- Аутентификация и авторизация: TODO (на `/ws` проверок нет, origin не проверяется)

### Шаблон раздела для нового модуля

```markdown
### Модуль <name>

Назначение: <одно-два предложения>.

**Маршруты** (префикс `/api/v1`):

| Метод | Путь | Тело / параметры | Ответ |
|---|---|---|---|

**Модели и таблицы:** <список>

**Зависимости от других модулей и сервисов:** <список>

**Особенности и ограничения:** <список>
```

## Работа с БД

Используется go-pg. Схема создаётся при старте модуля (`createSchema`, `CreateTable` с `IfNotExists` и внешними ключами). Порядок важен: сначала таблицы, на которые есть внешние ключи.

Таблицы модуля `projects`:

| Таблица | Назначение |
|---|---|
| `projects` | проекты |
| `files` | файлы проектов |
| `workspace_images` | собранные образы: хеш зависимостей, ревизия nixpkgs, тег, статус, `last_used_at` |
| `workspaces` | запущенные окружения: пользователь, образ, `container_id`, статус, `expires_at` |

Статусы workspace: `pending` → `running` → `stopped` или `failed`. Статусы образа: `building`, `ready`, `failed`.

`CreateTable IfNotExists` не изменяет существующие таблицы. После изменения моделей на dev-окружении таблицы нужно пересоздать, для постоянной работы нужны миграции (TODO).

## Файлы проектов на диске

Файлы проектов пользователей лежат в каталоге `PROJECT_PATH` (в dev-окружении это `tmp/`).

```
tmp/
└── <id проекта>/
    └── <имя проекта>/
        └── src/
            └── main.rs
```

Файлы попадают в контейнер с помощью docker mounts

## Разработка

```bash
make build        # go mod tidy + сборка bin/backend
make run          # то же + запуск
go vet ./...      # статический анализ
```

Соглашения:

- Ошибки HTTP возвращаются через `httpserver.ResponseError`, ошибки WebSocket через `httpserver.WSError`. Формат везде одинаковый.
- После апгрейда до WebSocket писать в `http.ResponseWriter` нельзя (соединение захвачено).
- Запись в WebSocket-соединение должна идти из одной горутины или под мьютексом (требование gorilla/websocket).
- Очистка после закрытия сессии (удаление контейнера, смена статуса в БД) выполняется со своим контекстом, а не с контекстом запроса: к этому моменту он уже отменён.

## Решение проблем

**`docker pull nixos/nix:...` обрывается с `EOF`.** Нестабильное соединение с Docker Hub на длинных загрузках. Уменьшите параллельность (`"max-concurrent-downloads": 1` в `daemon.json`) и повторяйте pull в цикле:

```bash
until docker pull nixos/nix:2.24.9; do echo retry; sleep 3; done
```

**`relation "workspace_images" does not exist`.** Таблицы создаются в неверном порядке или `createSchema` завершился с ошибкой, которую никто не проверил. Образ должен создаваться раньше workspace.

**В логах `response.WriteHeader on hijacked connection`.** Где-то после апгрейда WebSocket используется `http.ResponseWriter`. В этой фазе ошибки отправляются через `httpserver.WSError`.

**`<NAME> is not set` при старте.** Не задана обязательная переменная окружения, проверьте `.env`.

**Паника при старте на `Ping`.** PostgreSQL недоступен: проверьте, что контейнер с базой запущен и `DATABASE_HOST`, `DATABASE_PORT` и учётные данные в `.env` верны.

**Остались контейнеры с меткой `ide.managed`.** Автоматической очистки осиротевших контейнеров после падения сервера пока нет:

```bash
docker ps -a --filter label=ide.managed=true
docker rm -f $(docker ps -aq --filter label=ide.managed=true)
```

## Лицензия

MIT License

Copyright (c) 2026 https://github.com/Gretsir12

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
