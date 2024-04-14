# Сервис баннеров

# Запуск

~~~zsh
git clone https://github.com/realPointer/banners 
cd banners
make compose-up
~~~

# Запросы

## Дополнительные endpoints вне основного API

### Регистрация
~~~zsh
curl --location 'localhost:8080/v1/auth/register' \
--header 'Authorization: token' \
--header 'Content-Type: text/plain' \
--data '{
  "username": "username",
  "password": "password",
  "role": "admin"
}'
~~~

---

### Авторизация

```zsh
curl --location 'localhost:8080/v1/auth/login' \
--header 'Authorization: token' \
--header 'Content-Type: text/plain' \
--data '{
  "username": "username",
  "password": "password"
}'
```
Пример ответа:
~~~json
{
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3MTMxMzI1NzIsIlVzZXJuYW1lIjoiQW5kcmV3QWFhYSIsIlJvbGUiOiJhZG1pbiJ9.09XRDNSmHPU0ImkSxGjHvuu-uv69jpkv2ldoREjoLn4"
}
~~~

### Удаление баннеров по тэгу

```zsh
curl --location --request DELETE 'localhost:8080/v1/banner/tag/{tagID}' \
--header 'Authorization: token'
```

### Удаление баннеров по фиче

```zsh
curl --location --request DELETE 'localhost:8080/v1/banner/feature/{featureID}' \
--header 'Authorization: token'
```

## Задания
Основное задание:
- [x] Использован основной API
- [x] Реализована авторизация и аутентификация
- [x] Сделан интеграционный тест на получения баннера пользователем
- [x] Имплементирован кэш для использования флага `use_last_revision`
- [x] Имеется параметр отключения баннеров

Дополнительные задания:
- [x] Версионирование баннеров
- [x] Добавлены методы для удаления баннеров по фиче или тэгу
- [x] Описан линтер
