# RegistryUI

Веб-интерфейс для **Docker Registry v2**: просмотр репозиториев и тегов,
агрегированная статистика, метаданные образов (размер, дата, платформа, слои,
labels, entrypoint/cmd/env) и удаление тегов.

- **Бэкенд** — Go (1.25), прокси-агрегатор перед HTTP API Docker Registry (`/v2/...`)
  с сессионной авторизацией.
- **Фронтенд** — Vite + React 19 + TypeScript, TanStack Query, React Router,
  собственная CSS-дизайн-система, локализация EN/RU.

## Как это работает

На форме входа registry **выбирается из списка** — ввести произвольный URL
нельзя. Список задаётся при развёртывании через переменную `REGISTRIES`
(`GET /api/defaults` отдаёт его фронтенду). Логин/пароль никогда не
предзаполняются — пользователь вводит их сам.

После успешного `Ping` создаётся серверная сессия (in-memory, TTL 12 ч), а
клиенту выдаётся HttpOnly-cookie `rui_session`. Все registry-эндпоинты требуют
активной сессии. Бэкенд проверяет, что выбранный URL входит в `REGISTRIES`
(защита от обращения к произвольным адресам).

## Структура

```
.
├── main.go                  # entrypoint HTTP-сервера + graceful shutdown
├── internal/
│   ├── config/              # конфиг из переменных окружения + загрузка .env
│   ├── registry/            # клиент Docker Registry v2 (манифесты, blobs, stats)
│   ├── session/             # in-memory хранилище сессий
│   └── api/                 # роутер, CORS, авторизация, HTTP-хендлеры
└── web/                     # фронтенд (Vite + React)
```

## Запуск (Docker)

Образ собирается одной командой — multi-stage `Dockerfile` сам собирает фронтенд
(`npm ci && npm run build`) и Go-бинарник, а в рантайме сервер отдаёт и API, и
статику с одного порта (SPA-fallback на `index.html`).

```bash
docker build -t registryui .
docker run --rm -p 8080:8080 \
  -e REGISTRIES="Production=https://registry.example.com,Local=http://host.docker.internal:5000" \
  registryui
# UI: http://localhost:8080
```

`REGISTRIES` — список registry для выпадающего меню на форме входа (см. ниже).
Если его не задать, используется один пункт из `REGISTRY_URL`. Локально (без
Docker) фронтенд раздаётся Vite, а Go-сервер работает как API-only.

## Запуск (разработка)

### 1. Локальный Docker Registry (если ещё нет)

```bash
docker run -d -p 5000:5000 --name registry \
  -e REGISTRY_STORAGE_DELETE_ENABLED=true \
  registry:2
```

> `REGISTRY_STORAGE_DELETE_ENABLED=true` обязателен, иначе удаление тегов вернёт 405.

### 2. Go API

```bash
go run .
# слушает :8080, при логине ходит в registry (по умолчанию http://localhost:5000)
```

При старте автоматически подхватывается файл `.env` из корня (переменные,
уже заданные в окружении, не перезаписываются).

### 3. Фронтенд

```bash
cd web
npm install
npm run dev        # http://localhost:5173, /api проксируется на :8080
```

Сборка прода: `npm run build` (output в `web/dist/`).

## Переменные окружения (бэкенд)

| Переменная           | По умолчанию             | Назначение                                          |
|----------------------|--------------------------|-----------------------------------------------------|
| `PORT`               | `:8080`                  | адрес HTTP API (можно `8080` — двоеточие добавится) |
| `REGISTRIES`         | —                        | список registry для формы входа: `Имя=URL` через запятую |
| `REGISTRY_URL`       | `http://localhost:5000`  | резервный одиночный registry, если `REGISTRIES` пуст |
| `REGISTRY_TIMEOUT`   | `15s`                    | таймаут запроса к registry (`30s`, `1m` или `30`)   |
| `CORS_ORIGIN`        | `http://localhost:5173`  | разрешённый origin для dev-фронтенда                |
| `STATIC_DIR`         | `web/dist`               | каталог собранного фронтенда; если нет — API-only   |

Пример `REGISTRIES` (имя слева опционально, по умолчанию — хост URL):

```
REGISTRIES="Production=https://registry.example.com,Local=http://localhost:5000"
```

## API

Эндпоинты `/api/health`, `/api/stats`, `/api/repositories`, `/api/repository`,
`/api/tags`, `/api/tag` требуют валидной сессии (cookie `rui_session`).

### Авторизация и bootstrap

| Метод    | Путь            | Описание                                         |
|----------|-----------------|--------------------------------------------------|
| `GET`    | `/api/defaults` | значения для предзаполнения формы логина         |
| `POST`   | `/api/session`  | логин: `{registryUrl, username, password}` → cookie |
| `GET`    | `/api/session`  | текущая сессия (`registryUrl`, `username`)       |
| `DELETE` | `/api/session`  | логаут                                           |

### Registry

| Метод    | Путь                              | Описание                                  |
|----------|-----------------------------------|-------------------------------------------|
| `GET`    | `/api/health`                     | доступность registry                      |
| `GET`    | `/api/stats`                      | сводка: репозитории, теги, размер хранилища |
| `GET`    | `/api/repositories`               | список репозиториев                       |
| `GET`    | `/api/repository?repo=<name>`     | карточка репозитория (кол-во тегов, размер, дата) |
| `GET`    | `/api/tags?repo=<name>`           | теги репозитория                          |
| `GET`    | `/api/tag?repo=<name>&tag=<tag>`  | детали тега (манифест + конфиг + слои)     |
| `DELETE` | `/api/tag?repo=<name>&tag=<tag>`  | удалить тег                               |