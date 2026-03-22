# 👻 Ghost Reaction

> Бот на Go + CGo, который автоматически ставит реакции на каждое сообщение в Telegram-чате. Реакцию выбирает AI.

[![Go](https://img.shields.io/badge/Go-1.20+-00ADD8?logo=go)](https://golang.org)
[![TDLib](https://img.shields.io/badge/TDLib-1.8.62-blue)](https://github.com/tdlib/td)
[![License](https://img.shields.io/badge/license-MIT-green)](LICENSE)

---

## 📖 О проекте

**Ghost Reaction** использует официальную библиотеку [TDLib](https://github.com/tdlib/td) через CGo для взаимодействия с Telegram API. Бот слушает указанный чат и на каждое новое сообщение ставит реакцию, которую подбирает AI на основе контекста сообщения.

---

## ⚙️ Требования

- **Go** 1.20+
- **GCC / Clang** (C компилятор)
- **TDLib** (собранная из исходников)
- **Зависимости для сборки TDLib:**

```bash
sudo apt install build-essential cmake libssl-dev zlib1g-dev gperf
```

---

## 🗂️ Структура проекта

```
.
├── cmd/
│   └── main.go                  # Точка входа
├── internal/
│   ├── model/                   # Структуры данных
│   │   ├── commonResp.go
│   │   ├── initData.go
│   │   ├── msgResp.go
│   │   ├── msgType.go
│   │   └── reactionResp.go
│   ├── service/                 # Бизнес-логика
│   │   ├── auth.go
│   │   ├── init.go
│   │   ├── msgHelpers.go
│   │   └── reactMsg.go
│   └── tdlib/
│       ├── include/             # Заголовочные файлы TDLib для CGo
│       │   ├── td/telegram/
│       │   │   ├── GitCommitHash.cpp
│       │   │   └── tdjson_export.h
│       │   └── td_json_client.h
│       └── tdlib.go
├── lib/                         # Скомпилированные библиотеки TDLib (.so / .a)
├── go.mod
├── go.sum
└── LICENSE
```

> **Примечание:** директории `lib/`, `internal/tdlib/include/`, `tdlib-db/` и `tdlib-files/` добавлены в `.gitignore` — их нужно создать вручную.

---

## 🚀 Установка и запуск

### 1. Клонируйте репозиторий

```bash
git clone https://github.com/GAZIZlikesLollipop/ghost-reaction
cd ghost-reaction
```

### 2. Соберите TDLib

Перейдите на официальную страницу сборки TDLib и сгенерируйте инструкции под вашу ОС и компилятор:

👉 **https://tdlib.github.io/td/build.html**

### 3. Подготовьте директории

```bash
mkdir -p internal/tdlib/include/td/telegram lib
```

### 4. Перенесите файлы TDLib

После сборки TDLib скопируйте нужные файлы в проект:

```bash
# Основной заголовочный файл
cp td/td/telegram/td_json_client.h internal/tdlib/include/

# Сгенерированные заголовки (из папки сборки)
cp -r td/build/td/telegram internal/tdlib/include/td/

# Скомпилированные библиотеки
cp td/build/libtdjson.so lib/                   # динамическая
cp td/build/libtdjson.so.1.8.62 lib/

# Статические библиотеки (нужны для статической линковки)
cp td/build/lib*.a lib/
```

### 5. Получите API-ключи

- **Telegram** — зарегистрируйте приложение на [my.telegram.org/apps](https://my.telegram.org/apps) и получите `app_id` и `app_hash`.
- **Gemini** — получите API-ключ на [aistudio.google.com/app/api-keys](https://aistudio.google.com/app/api-keys).

### 6. Задайте переменные окружения

```bash
export APP_ID="ваш_app_id"
export APP_HASH="ваш_app_hash"
export API_KEY="ваш_gemini_api_key"
export AI_MODEL="gemini-2.5-flash-lite"         # или другая модель Gemini
export REQUEST_DELAY="10"                        # задержка между запросами в секундах (по умолчанию 10)
```

### 7. Запустите

```bash
go run cmd/main.go
```

---

## 🔨 Сборка

Перед сборкой убедитесь, что все статические `.a`-библиотеки TDLib находятся в директории `lib/`, и задайте переменную `CC`, указывающую на ваш C-компилятор.

### Linux

```bash
export CC=gcc   # или clang

CGO_ENABLED=1 \
CGO_LDFLAGS="-L$(pwd)/lib -Wl,--start-group \
  -ltdjson_static -ltdjson_private \
  -ltdclient -ltdcore -ltdapi -ltdmtproto \
  -ltdnet -ltddb -ltdactor -ltdutils \
  -ltdsqlite -ltde2e \
  -lssl -lcrypto -lc++ -lz -lm -ldl -lpthread \
  -Wl,--end-group" \
go build \
  -ldflags="-linkmode external -extldflags '-static'" \
  -o grBin \
  cmd/main.go
```

> Флаги `-Wl,--start-group ... -Wl,--end-group` необходимы, так как статические библиотеки TDLib имеют циклические зависимости между собой.

### Windows

Переходите по следующей ссылке и скачивайте [убунту](https://ubuntu.com/download/desktop) 🫠

---

## ❓ Частые проблемы

| Проблема | Решение |
|---|---|
| `cannot find -ltdjson` | Убедитесь, что `libtdjson.so` / `libtdjson_static.a` находятся в `lib/` |
| `td_json_client.h: No such file` | Проверьте путь `internal/tdlib/include/td_json_client.h` |
| Ошибка авторизации | Проверьте правильность `APP_ID` и `APP_HASH` |
| `CGO_ENABLED` ошибка | CGo включён по умолчанию; убедитесь, что GCC установлен и переменная `CC` указывает на корректный путь к компилятору |
| Ошибка линковки `-lc++` | Установите `libc++-dev` или замените на `-lstdc++` для GCC |

---

## 📄 Лицензия

MIT — делайте что хотите, но звёздочку поставьте 🌟
