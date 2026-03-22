# 👻 Ghost Reaction

> Бот на Go + CGo, который автоматически ставит реакции на каждое сообщение в Telegram-чате. Реакцию выбирает AI.

[![Go](https://img.shields.io/badge/Go-CGo-00ADD8?logo=go)](https://golang.org)
[![TDLib](https://img.shields.io/badge/TDLib-official-blue)](https://github.com/tdlib/td)
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

Создайте директории `lib` корне проекта и `include` в `internal/tdlib`:

```bash
mkdir -p internal/tdlib/include/td lib
```

### 4. Перенесите файлы библиотеки

```bash
# Заголовочный файл
cp td/td/telegram/td_json_client.h include/

# Сгенерированные заголовки (из папки сборки)
cp -r td/build/td/telegram include/td

# Скомпилированная библиотека
cp td/build/libtdjson.so lib/   # или .dylib на macOS
```

### 5. Получите Telegram и Gemini API credentials

Зарегистрируйте приложение на [my.telegram.org/apps](https://my.telegram.org/apps) и получите `app_id` и `app_hash`.
Далее получите API key на [сайте](https://aistudio.google.com/app/api-keys).

### 6. Задайте переменные окружения

```bash
export APP_ID="ваш_app_id"
export APP_HASH="ваш_app_hash"
export API_KEY="ваш_gemini_api_key"
export AI_MODEL="модель_гемини" # например gemini-2.5-flash-lite
export REQUEST_DELAY="задержка_между_запросами_в_секундах" # опционально (по умолчанию 10с)
```

### 7. Запустите

```bash
go run cmd/main.go
```

---

## 🗂️ Структура проекта

```
.
├ cmd/main.go          # Точка входа
├ internal/
│   ├ model/           # Структуры данных (сообщения, реакции, ответы)
│   ├ service/         # Бизнес-логика (авторизация, работа с сообщениями)
│   └ tdlib/
│       └ include/     # Заголовочные файлы TDLib (.h) для cgo
├ lib/                 # Скомпилированные бинарники TDLib (.so / .a)
└ go.mod
└── README.md
```

---

## ❓ Частые проблемы

| Проблема | Решение |
|---|---|
| `cannot find -ltdjson` | Убедитесь что `libtdjson.so` лежит в `lib/` |
| `td_json_client.h: No such file` | Проверьте путь `include/td_json_client.h` |
| Ошибка авторизации | Проверьте правильность `APP_ID` и `APP_HASH` |
| `CGO_ENABLED` ошибка | CGo включён по умолчанию, убедитесь что GCC установлен и переменная CC указывает на путь к компилятору C |

---

## 📄 Лицензия

MIT — делайте что хотите, но звёздочку поставьте 🌟
