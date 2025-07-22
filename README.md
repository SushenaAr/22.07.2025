# 22.07.2025
Для запуска в корневой папке проекта создать папки resources->(downloads, zip)
## Эндпоинты:
### GET /resources/zip/:filename  
Получить готовый зип файл (перейти из последнего эндпоинта)
### POST /task  
Создать таску на создание зип файла  
Response - id таски  
### PATCH /task  
Добавить файл в таску  
Query params - id(получить в ответе из второго эндпоинта) и url(ссылка на файл с инета. должна оканчиваться на filename.ext)  
Response - ok  
### GET /task/status - получить статус задачи
Query params - id(получить в ответе из второго эндпоинта)  
Response - количество загруженных файлов | количество загруженных файлов и ссылка на первый эндпоинт  
## .env параметры
AppEnv         string - отвечает за логи (development, local) default:development  
Port           string - порт приложения default:8080  
AdvertisedAddr string - адрес для создания ссылок default:127.0.0.1:8080  
Protocol       string - тоже нужен для создания ссылки(http, https) default:http  
Но все запускается из коробки по адресу localhost:8080
## Тесты, к сожалению, написать не успел
## Папочная структура дефолт для REST API
