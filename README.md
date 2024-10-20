# Draftsmith API

## Introduction

<p><img src="./assets/logo.png" style="float: left; width: 80px" /></p>



Draftsmith is notetaking tool that uses PostgreSQL as a backend to implement full text search and semantic search via an openAI API. This is an API that is used to interact with the database and can be implemented by any front-end application, basically a bring your own GUI.

See the PyQt6 GUI implementation [here (TODO)] and the Neovim Extension [here (TODO)].

It is designed to be a simple, fast, and reliable way to take notes, whilst remaining feature complete and open source.

## Installation


1. Clone the repository
    ```sh
    git clone TODO
    ```
2. Start the Docker Container

    ```sh
    docker-compose up
    ```

The binary is embedded in the docker container, which automatically starts the server on port `37238` [^1729388462]

## Development

Use [PostgreSQL-Browser for Browsing the Database](https://github.com/RyanGreenup/PostgreSQL-Browser).

Create the database:

```sh
PGPASSWORD=postgres \
    psql -h localhost -p 5432 -U postgres -f ./src/cmd/draftsmith.sql
pg_browser postgres  --host localhost --password postgres --username postgres --port 5432
```




## Footnotes

[^1729388462]: This has a mnemonic:


    | Letter | T9 |
    |--------|----|
    | d      | 3  |
    | r      | 7  |
    | a      | 2  |
    | f      | 3  |
    | t      | 8  |

