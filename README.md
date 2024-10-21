# Draftsmith API

## Introduction

<p><img src="./assets/logo.png" style="float: left; width: 80px" /></p>



Draftsmith is notetaking tool that uses PostgreSQL as a backend to implement full text search and semantic search via an openAI API. This is an API that is used to interact with the database and can be implemented by any front-end application, basically a bring your own GUI.

See the PyQt6 GUI implementation [here (TODO)] and the Neovim Extension [here (TODO)].

It is designed to be a simple, fast, and reliable way to take notes, whilst remaining feature complete and open source.

## Installation

Installation is supported via docker (although it is possible to run the server locally). See the [installation documentation](https://ryangreenup.github.io/draftsmith_api/installation.html) for more information:


```sh
git clone https://github.com/RyanGreenup/draftsmith_api
docker compose build
docker compose up db -d
sleep 5  # wait for db to start
docker compose run app ./draftsmith_api --db_host=db cli init
docker compose down
docker compose up
curl http://localhost:37238/notes/tree | jq
```

## Related Software

- [PostgreSQL-Browser for Browsing the Database](https://github.com/RyanGreenup/PostgreSQL-Browser).
- Neovim Extension (TODO)
- PyQt6 GUI (TODO)
    - [Draftsmith](https://github.com/RyanGreenup/draftsmith_api)
        - This is still file based, migration toward the API is under way [Draftsmith QT /  Move toward REST API with PostgresQL backend #1 ](https://github.com/RyanGreenup/Draftsmith/issues/1)
- Flask Server (TODO)


## Footnotes

[^1729388462]: This has a mnemonic:


    | Letter | T9 |
    |--------|----|
    | d      | 3  |
    | r      | 7  |
    | a      | 2  |
    | f      | 3  |
    | t      | 8  |

