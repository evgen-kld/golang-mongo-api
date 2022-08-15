<h1>Golang MongoDB API</h1>

Приложение реализует API с использованием пакета gorilla/mux.  

Сервер предоставляет 4 эндпоинта:  
GET `/` - получить все данные  
POST `/` (body: json) - создать новую запись  
DELETE `/<id>` - удалить запись по её id  
PUT `/<id>` (body: json) - изменить запись по её id  

Хранение данных происходит в базе данных MongoDB и в кеше (in memory).
Каждые 100 секунд приложение обращается к БД и обновляет кеш.  

Структура записи в БД:
`{"_id": "62fa86cff6534d59eceba870", "firstname": "Ivan", "lastname": "Ivanov"}`