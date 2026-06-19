# RegistryUI

Веб-интерфейс для локального **Docker Registry v2**: просмотр репозиториев, тегов,
метаданных образов (размер, дата, платформа, слои, labels) и удаление тегов.

- **Бэкенд** — Go, прокси-агрегатор перед HTTP API Docker Registry (`/v2/...`).
- **Фронтенд** — Vite + React + TypeScript, TanStack Query, React Router, Tailwind v4, shadcn/ui.

## Структура

```
.
├── main.go                  # entrypoint HTTP-сервера
├── internal/
│   ├── config/              # конфиг из переменных окружения
│   ├── registry/            # клиент Docker Registry v2
│   └── api/                 # роутер и HTTP-хендлеры
└── web/                     # фронтенд (Vite)
```

## Запуск (разработка)

Нужны два процесса.

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
# слушает :8080, ходит в registry на http://localhost:5000
```

### 3. Фронтенд

```bash
cd web
npm install
npm run dev        # http://localhost:5173, /api проксируется на :8080
```

## Переменные окружения (бэкенд)

| Переменная           | По умолчанию             | Назначение                              |
|----------------------|--------------------------|-----------------------------------------|
| `PORT`               | `:8080`                  | адрес HTTP API (можно `8080`)           |
| `REGISTRY_URL`       | `http://localhost:5000`  | базовый URL Docker Registry             |
| `REGISTRY_USERNAME`  | —                        | basic-auth логин (опционально)          |
| `REGISTRY_PASSWORD`  | —                        | basic-auth пароль (опционально)         |
| `REGISTRY_TIMEOUT`   | `15s`                    | таймаут запроса к registry              |
| `CORS_ORIGIN`        | `http://localhost:5173`  | разрешённый origin для dev-фронтенда    |

## API

| Метод    | Путь                              | Описание                       |
|----------|-----------------------------------|--------------------------------|
| `GET`    | `/api/health`                     | доступность registry           |
| `GET`    | `/api/repositories`               | список репозиториев            |
| `GET`    | `/api/tags?repo=<name>`           | теги репозитория               |
| `GET`    | `/api/tag?repo=<name>&tag=<tag>`  | детали тега (манифест + конфиг) |
| `DELETE` | `/api/tag?repo=<name>&tag=<tag>`  | удалить тег                    |
