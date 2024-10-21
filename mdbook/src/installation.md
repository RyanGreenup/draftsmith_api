
## Installation


1. Clone the repository
    ```sh
    git clone https://github.com/RyanGreenup/draftsmith_api
    ```
2. Start the Docker Container

    ```sh
    docker-compose up
    ```

The binary is embedded in the docker container, which automatically starts the server on port `37238` [^1729388462]
1. Clone the Repository

    ```sh
    git clone TODO
    ```

2. Initialize the Database

    ```bash
    # Build the containers
    docker compose build
    docker compose up db -d
    sleep 5  # wait for db to start

    # Populate the database
    docker compose run app ./draftsmith_api --db_host=db cli init
    docker compose down

    # Start the server
    docker compose up

    # Test the server
    curl http://localhost:37238/notes/tree | jq
    ```

After that, it is sufficient to run `docker compose up` to start the server.

## Updating

1. Git Pull
2. Rebuild the Docker Container

    ```sh
    docker compose build
    ```
3. Restart the Docker Container

    ```sh
    docker compose up
    ```

## Debugging

### Enter the container

To jump into the container and have a look around, you can use the following command:

```sh
docker compose down
docker compose up db -d
docker compose run app tail -f /dev/null -d
docker compose exec app sh
```

This will start the database and the application, and then open a shell in the container.

### Entry Point Script

Here is an entrypoint script that can be used to wait for the database to be ready before starting the application:

```sh
#!/bin/bash

# A command to wait for the database to be ready
until psql -h "db" -U "postgres" -c '\q'; do
  >&2 echo "Postgres is unavailable - sleeping"
  sleep 1
done

>&2 echo "Postgres is up - executing command"

# Now run the actual application setup command
./draftsmith_api --db_host=db cli init

# Finally start the main application
exec /app/draftsmith_api

```

### Initializing the Database

Create the database:

```sh
PGPASSWORD=postgres \
    psql -h localhost -p 5432 -U postgres -f ./src/cmd/draftsmith.psql
pg_browser postgres  --host localhost --password postgres --username postgres --port 5432
```

Alternatively the go binary can be used to initialize the database:

```sh
./draftsmith_api --db_host=db cli init
```

See also [PostgreSQL-Browser for Browsing the Database](https://github.com/RyanGreenup/PostgreSQL-Browser).


## Footnotes

[^1729388462]: This has a mnemonic:


    | Letter | T9 |
    |--------|----|
    | d      | 3  |
    | r      | 7  |
    | a      | 2  |
    | f      | 3  |
    | t      | 8  |

