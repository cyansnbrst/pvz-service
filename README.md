# Сервис для работы с ПВЗ

Сервис предоставляет интерфейс для работы сотрудников с ПВЗ, приёмками и товарами.

## Требования
- Docker
- migrate

## Установка и запуск
1. Склонируйте репозиторий, перейдите в папку, где лежит Makefile.
2. Запустите Docker.
3. Запустите базовую инфраструктуру сервиса:
    ```bash
    make up
    ```

4. Примените миграции:
    ```bash
    make migrate/up
    ```

5. Если что-то не работает, проверить конфликтующие порты (:8080, :5432, :3000, :9000, :8085) и поменять их в .env файле и докер композе.

## Тесты и линтеры
Конфигурация линтеров описана в `.golangci.yml`. 

В этом проекте используются линтеры для:

- Форматирования кода (`gofmt`, `goimports`).
- Поиска ошибок и багов (`govet`, `staticcheck`, `errcheck`).
- Анализа производительности и безопасности (`gosec`, `gocritic`, `prealloc`).
- Обнаружения дублирования и стилистических проблем (`revive`, `nestif`, `misspell`).

Запуск тестов и линтеров (необходим Docker, лучше заранее поставить образ postgres:15-alpine3.18): 

```bash
make audit
```

Проверка покрытия: 

```bash
make сoverage
```

## Пример API-запросов
Примеры запросов можно найти в файле `PVZ Service.postman_collection.json`

Также доступен Swagger UI: http://localhost:8085/

### Проблема 1. Хардкод городов/ролей/типов при проверке на их валидность.
Для валидации городов и ролей используется хардкод, а не хранение в БД. Хотя такой подход снижает гибкость, он оправдан в текущих условиях:
- Роли практически не изменяются
- В системе всего три города
- Высокие требования к производительности
При расширении географии сервиса данные могут быть перенесены в БД с кэшированием, чтобы избежать лишних запросов.

### Проблема 2. Разделение путей на требующие и не требующие авторизации 
Проверка авторизации реализована в middleware через явный список публичных эндпоинтов. Решение принято по двум причинам:
1. Кодогенерация OpenAPI усложняет группировку роутов
2. В системе всего три публичных эндпоинта

### Проблема 3. Тестирование хендлеров
Для тестирования хендлеров использовались интеграционные тесты, а не unit-тесты. Такой подход был выбран потому что:
- Хендлеры работают с внешними зависимостями
- Интеграционные тесты лучше проверяют реальное поведение системы
- Покрывают весь workflow от запроса до ответа
Unit-тесты не давали бы реальной картины работы системы.

### Проблема 4. Proto-файл
В задании написано, что GRPC-хендлер должен отдавать все существующие в системе ПВЗ. То же самое описано и в proto файле, GetPVZListResponse содержит просто список ПВЗ (repeated PVZ pvzs). В файле же присутствует еще enum ReceptionStatus, который был удален, потому что нигде не используется.