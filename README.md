# Песочница для автоматической проверки программ студентов

Данный репозиторий содержит исходный код песочницы, используемой в целях автоматической проверки программ студентов.

### Установка
#### Предварительные настройки

##### Настройка сервера RabbitMQ 
Убедитесь, что запущен образ сервера RabbitMQ. Создать его можно с помощью команды 
```
docker run -d --hostname my-rabbit --name my-rabbit -p 8088:15672 -p 5672:5672 --net=mynet rabbitmq:management
```
Запуск сервера
```
docker start my-rabbit
```
#### Настройка переменных окружения

Первы делом необходимо задать переменные окружения в файле .env в cmd/main. Список параметров окружения приведен в таблице ниже: 

| Переменная  | Описание |
| ------ | ------ |
| DOCKER_NETWORK | Наименование сети в docker. Тип: строка |
| CODE_STORAGE_PATH | Путь к хранилищу решений на хосте. Тип: строка |
| LANG_CONFIG_FILE_PATH | Путь к файлу конфигурации компиляторов на хосте. Тип: строка |
| TASK_QUEUE | Наименование очереди решений в RabbitMQ. Тип: строка |
| QUEUE_EXCHANGE | Наименование exchange в RabbitMQ. Тип: строка |
| RESULT_QUEUE | Наименование очереди результатов в RabbitMQ. Тип: строка |
| AMQPS_SCHEME | AMQP/AMQPS URL в RabbitMQ. Тип: строка |
| DOCKER_NETWORK_ID | ID сети DOCKER_NETWORK в docker. Тип: строка |
| TARGET_FILE_STORAGE_PATH | Путь к хранилищу решений в контейнере. Тип: строка |
| RABBIT_HOST_NAME | Наименование хоста контейнера RabbitMQ. Тип: строка |
| TASK_STORAGE_URL | URL хранилища тасков, где taskID есть $taskID. Тип: строка |
| DOCKER_URL_OF_TASK_STORAGE |  URL хранилища тасков, где taskID есть $taskID для контейнера. На not-Linux машинах для доступа к localhost хоста необходимо прописывать как docker.host.internal. Тип: строка |
| IS_CONTAINER_STARTED | true или false. По умолчанию false. Тип: строка |