<div align="center">

# pb-ftp

### Запускник FTP-сервера для PocketBook с QR-кодом подключения

[![Релизы](https://img.shields.io/badge/%D0%A0%D0%B5%D0%BB%D0%B8%D0%B7%D1%8B-%D0%A1%D0%BA%D0%B0%D1%87%D0%B0%D1%82%D1%8C-2F6FED.svg)](../../releases/latest)
[![PocketBook](https://img.shields.io/badge/PocketBook-FTP%20server-2F6FED.svg)](../../releases)
[![Go](https://img.shields.io/badge/Go-1.23-00ADD8.svg?logo=go&logoColor=white)](https://go.dev/)
[![Лицензия](https://img.shields.io/badge/%D0%9B%D0%B8%D1%86%D0%B5%D0%BD%D0%B7%D0%B8%D1%8F-GPL--2.0-blue.svg)](LICENSE)

</div>

---

**pb-ftp** запускает FTP-сервер на PocketBook и показывает на экране QR-код с адресом подключения. QR-код содержит обычный FTP URL, поэтому сервер можно использовать с любым клиентом, который умеет подключаться по FTP в локальной сети.

Проект создан как серверная часть для **[eBookSender](https://github.com/CyberCat2033/eBookSender)**, но не привязан к конкретному Android-приложению.

---

## Возможности

- Запуск FTP-сервера `vsftpd` на порту `2121`.
- Подключение без пароля по адресу вида `ftp://anonymous@<ip>:2121/mnt/ext1/`.
- QR-код на экране PocketBook с FTP-адресом устройства.
- Текстовая ошибка на экране, если сервер не удалось запустить.
- Автоматическое восстановление FTP-сервера, если PocketBook проснулась после сна.
- Локальный HTTP API `POST /rescan` на порту `2122` для обновления библиотеки после передачи файлов.
- Локальный HTTP API `GET /version` на порту `2122` для проверки установленной версии лаунчера.
- Локальный HTTP API `POST /update` на порту `2122` для безопасной активации staged-обновления лаунчера.
- Запуск стандартного сканера библиотеки PocketBook при закрытии приложения.
- Удержание сетевого подключения активным, пока приложение открыто.

---

## Как это работает

```text
PocketBook
    |
    | запуск pb-ftp
    v
FTP-сервер на :2121
    |
    | QR-код с ftp://anonymous@<ip>:2121/mnt/ext1/
    v
FTP-клиент подключается к устройству по локальной сети
    |
    v
Файлы загружаются в память PocketBook
```

---

## Установка для пользователей

### 1. Скачайте архив

1. Откройте страницу [последнего релиза](../../releases/latest).
2. Скачайте архив вида:

   ```text
   pb-ftp-vX.Y.Z.tar.gz
   ```

### 2. Распакуйте файлы

В архиве находятся:

```text
pb-ftp.app
vsftpd
vsftpd.conf
```

### 3. Скопируйте на PocketBook

Подключите PocketBook к компьютеру по USB и скопируйте все три файла в каталог приложений:

```text
/mnt/ext1/applications/
```

При подключении по USB этот каталог обычно отображается как:

```text
applications/
```

### 4. Запустите сервер

1. Отключите PocketBook от компьютера.
2. Подключите PocketBook и устройство с FTP-клиентом к одной Wi-Fi сети.
3. Откройте `pb-ftp` на PocketBook из списка приложений.
4. Используйте FTP-адрес с экрана для подключения.
5. Загрузите файлы на PocketBook.

Адрес для ручного подключения выглядит так:

```text
ftp://anonymous@<ip>:2121/mnt/ext1/
```

---

## Частые сценарии

### Запустить FTP-доступ к PocketBook

1. Запустите `pb-ftp` на PocketBook.
2. Дождитесь появления QR-кода и FTP-адреса.
3. Подключитесь к указанному адресу из FTP-клиента.
4. Загрузите нужные файлы.

### Подключиться вручную

1. Запустите `pb-ftp` на PocketBook.
2. Скопируйте FTP-адрес с экрана.
3. Введите его в FTP-клиенте.

### Обновить библиотеку после передачи

`pb-ftp` запускает обновление библиотеки при выходе из приложения. Клиенты также могут запросить обновление через `POST /rescan` на порту `2122`.

### Проверить версию лаунчера

Клиенты могут запросить установленную версию через `GET /version` на порту `2122`.

Ответ:

```json
{
  "schemaVersion": 1,
  "appName": "pb-ftp",
  "versionName": "1.0.0",
  "versionCode": 123,
  "buildId": "abcdef123456...",
  "releasedAt": "2026-06-19T12:00:00Z"
}
```

`GET /version` возвращает версию, прошитую в бинарник при сборке.

### Активировать staged-обновление лаунчера

Клиент должен сначала загрузить новый `.app` по FTP в staging-каталог:

```text
/mnt/ext1/applications/.pb-ftp-update/
```

После проверки `sha256` клиент вызывает `POST /update`:

```json
{
  "sourcePath": "/mnt/ext1/applications/.pb-ftp-update/pb-ftp.app",
  "versionName": "1.0.1",
  "versionCode": 124,
  "buildId": "abcdef123456...",
  "releasedAt": "2026-06-19T12:00:00Z",
  "sha256": "64_hex_chars"
}
```

`pb-ftp` принимает только clean absolute path внутри staging-каталога, проверяет checksum, запрещает downgrade по `versionCode`, сохраняет текущий launcher как `pb-ftp.app.previous`, атомарно переносит staged-файл в `pb-ftp.app`, запускает стандартное сканирование накопителя и перезапускает себя через новый launcher. Endpoint игнорирует неизвестные JSON-поля, чтобы Android-клиент мог добавлять необязательные метаданные без поломки старого сервера.

---

## Диагностика

- Если QR-код не появился, проверьте, что PocketBook подключен к Wi-Fi и файлы `pb-ftp.app`, `vsftpd`, `vsftpd.conf` лежат в `applications/`.
- Если FTP-клиент не подключается, убедитесь, что он находится в одной Wi-Fi сети с PocketBook.
- Если используется VPN, временно отключите его или разрешите локальные подключения в настройках клиента.
- Если сеть гостевая, проверьте, не запрещает ли роутер обмен данными между устройствами.
- Если файл передался, но не появился в библиотеке, закройте `pb-ftp` на PocketBook или запустите обновление библиотеки из клиента.

---

## Сборка для разработчиков

### Требования

| Компонент | Версия |
| --- | --- |
| Go | 1.17 |
| Docker | актуальная стабильная версия |
| PocketBook build SDK | `5keeve/pocketbook-go-sdk:6.3.0-b288-v1` |

### Сборка через скрипт

В корне проекта:

```sh
./build.sh
```

Результат:

```text
pb-ftp.app
```

### Сборка вручную

```sh
docker run --rm \
  -v "$PWD":/src \
  -w /src \
  --net=host \
  5keeve/pocketbook-go-sdk:6.3.0-b288-v1 \
  build -trimpath -ldflags "-s -w" -o pb-ftp.app ./cmd/app
```

### Тесты

```sh
go test ./...
```

### Релизы и автообновление через Android

При пуше тега вида `vX.Y.Z` workflow `.github/workflows/ci-cd.yml`:

- собирает `pb-ftp.app`;
- публикует в GitHub Release архив `pb-ftp-vX.Y.Z.tar.gz`;
- публикует стабильный файл лаунчера `pb-ftp.app` в GitHub Pages рядом с manifest;
- обновляет GitHub Pages manifest `updates/latest.json`.

Ожидаемый URL манифеста для Android-приложения:

```text
https://cybercat2033.github.io/pb-ftp/updates/latest.json
```

Обновлением лаунчера на книжке занимается Android-приложение: оно читает manifest, сравнивает версию с `GET /version`, скачивает `launcher` artifact, проверяет `sha256`, загружает его в staging-каталог по FTP и активирует через `POST /update`. Для старых версий `pb-ftp` без `POST /update` Android-приложение может один раз использовать legacy fallback: загрузить проверенный launcher напрямую в путь из `installPath`.

Файл `.app` публикуется через GitHub Pages как `updates/pb-ftp.app`, потому что GitHub Release API отклоняет release asset с расширением `.app`. Имя launcher-файла остаётся стабильным и совпадает с путём установки на PocketBook; версия хранится в manifest, `GET /version` и верхней строке интерфейса `pb-ftp`.

`versionCode` в release workflow берётся из `GITHUB_RUN_NUMBER`, а `buildId` из git SHA. Поэтому при экстренной перепубликации того же SemVer-тега клиент всё равно отличает новую сборку по монотонному `versionCode` и/или `buildId`; сам `versionName` используется только как человекочитаемая версия.

Для публикации manifest в GitHub Pages у репозитория должна быть включена Pages-публикация из GitHub Actions. Workflow уже запрашивает нужные permissions:

```yaml
pages: write
id-token: write
```

---

## Архитектура проекта

| Путь | Назначение |
| --- | --- |
| `cmd/app` | точка входа PocketBook-приложения |
| `internal/control` | локальный HTTP-сервис для запроса пересканирования библиотеки |
| `internal/netutils` | запуск FTP-сервера, остановка процесса и получение локального IP |
| `internal/rescan` | запуск стандартного сканера библиотеки PocketBook |
| `internal/ui` | отрисовка QR-кода и экранных сообщений через InkView |
| `internal/version` | embedded-версия лаунчера для Android-обновлений |
| `assets/vsftpd` | бинарник FTP-сервера для устройства |
| `assets/vsftpd.conf` | конфигурация FTP-сервера |

---

## Технологии

| Технология | Для чего используется |
| --- | --- |
| Go 1.23 | основной язык приложения |
| `github.com/dennwc/inkview` | интеграция с UI и системными возможностями PocketBook |
| `github.com/skip2/go-qrcode` | генерация QR-кода |
| `vsftpd` | FTP-сервер на устройстве |
| `5keeve/pocketbook-go-sdk` | сборка приложения под PocketBook |

---

## Связанные проекты

- **[eBookSender](https://github.com/CyberCat2033/eBookSender)** — Android-приложение, для которого изначально создавался этот FTP-сервер.

---

## Благодарности

Спасибо **[dennwc](https://github.com/dennwc)** за **[`inkview`](https://github.com/dennwc/inkview)** — Go SDK для PocketBook, на котором построена интеграция с UI и системными возможностями устройства.

Спасибо **[5keeve](https://github.com/5keeve)** за проект **[`pocketbook-go-sdk`](https://github.com/5keeve/pocketbook-go-sdk)**, который упрощает сборку приложений для PocketBook.

---

## Лицензия

Проект распространяется под лицензией **GNU General Public License v2.0**.

Полный текст лицензии находится в [LICENSE](LICENSE).
