# Magicompose

Convert your docker-compose.yaml file to a [Dagger](https://dagger.io) module.

## Usage

Let's say you have the following docker-compose.yaml file:

```yaml
version: '3.7'

services:
  postgres:
    image: postgres:12
    restart: always
    environment:
      - POSTGRES_USER=admin
      - POSTGRES_PASSWORD=changeme
    volumes:
      - ./postgres/postgres.conf:/usr/local/etc/postgres/postgres.conf
      - ./postgres/:/docker-entrypoint-initdb.d/
      - postgres-data:/var/lib/postgresql/data
    command: postgres -c config_file=/usr/local/etc/postgres/postgres.conf
    ports:
      - '5432:5432'
  redis:
    image: redis:7
    restart: always
    command: redis-server
    ports:
      - '6379:6379'

volumes:
  postgres-data:
```

```
dagger -m https://github.com/TomChv/magicompose call --file docker-compose.yml generate -o .dagger

dagger -m .dagger functions
Name       Description
postgres   -
redis      -

# Run the service
dagger -m .dagger call redis up
```