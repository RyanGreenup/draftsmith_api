# Draftsmith API

## Introduction

<p><img src="./assets/logo.png" style="float: left; width: 80px" /></p>



Draftsmith is notetaking tool that uses PostgreSQL as a backend to implement full text search and semantic search via an openAI API. This is an API that is used to interact with the database and can be implemented by any front-end application, basically a bring your own GUI.

See the PyQt6 GUI implementation [here (TODO)] and the Neovim Extension [here (TODO)].

It is designed to be a simple, fast, and reliable way to take notes, whilst remaining feature complete and open source.

## Installation


## Related Software

- [PostgreSQL-Browser for Browsing the Database](https://github.com/RyanGreenup/PostgreSQL-Browser).
- Neovim Extension (TODO)
- PyQt6 GUI (TODO)
    - [Draftsmith](https://github.com/RyanGreenup/draftsmith_api)
        - This is still file based, migration toward the API is under way [Draftsmith QT /  Move toward REST API with PostgresQL backend #1 ](https://github.com/RyanGreenup/Draftsmith/issues/1)
- Flask Server (TODO)
- CLI
    - [Draftsmith CLI](https://github.com/RyanGreenup/draftsmith_cli)
        - This is a basic CLI that implements a Python API Client and exposes some basic functionality through a Typer CLI



## Installation

To get started with NoteMaster:

### Backend

Installation is supported via docker (although it is possible to run the server locally). See the [installation documentation](https://ryangreenup.github.io/draftsmith_api/installation.html) for more information:


```sh
git clone https://github.com/RyanGreenup/draftsmith_api
cd draftsmith_api
docker compose build
docker compose up db -d
sleep 5  # wait for db to start
docker compose run app ./draftsmith_api --db_host=db cli init
docker compose down
docker compose up
curl http://localhost:37238/notes/tree | jq
```

### Interfaces

2. **CLI Client**
   - Install Python and dependencies.
       - `pipx install git+https://github.com/RyanGreenup/draftsmith_cli --force`

3. [**PyQt GUI**](https://github.com/RyanGreenup/draftsmith)
   - `pipx install https://github.com/RyanGreenup/draftsmith`

4. **Web UI (Flask)**
   - To Be Implemented

## Usage

The API is documented in the [API Documentation](https://ryangreenup.github.io/draftsmith_api/usage.html).

## Contribution

I warmly welcome contributions from the community!

- **Fork the Repository**: Make the changes you’d like to see.
- **Submit a Pull Request**: I'd be thrilled to review and merge all contributions.
    - Please ensure that your code has tests and adheres to the `ruff` / `black` formatter.
- **Report Issues**: Help me by reporting bugs or suggesting features via the issue tracker.

## Roadmap

- [ ] Implement Flask or Django Web UI.
- [ ] Develop a Jetpack / Flutter mobile application for Android.
- [ ] Implement the PyQT GUI.
- [ ] Implement Structured Data and features similar to [Semantic Mediawiki](https://www.semantic-mediawiki.org/wiki/Semantic_MediaWiki), notion and [Dokuwiki's Struct Plugin](https://www.dokuwiki.org/plugin:struct) features.
- [ ] Implement real-time collaboration features.
- [ ] Implement Semantic Search and Tagging via OpenAI API such as [ollama](https://ollama.com/)
    - See [this issue](https://github.com/RyanGreenup/draftsmith_api/issues/2)
- [ ] Implement RAG over notes
- [ ] Implement a Kanban board view
- [ ] Implement a timeline view
- [ ] Implement a mindmap view
- [ ] Implement a table view for structured data
- [ ] Implement a graph view
- [ ] Implement a calendar view

## Join Our Community

Connect with me on Discord (`Eisenvig`) / Matrix (`@eisenvig:matrix.org`), or follow me on [Mastodon](`@ryangreenup@mastodon.social`) for the latest updates.

---

Structured thinking is increasingly important now that LLM's are making information more accessible than ever before.


## Footnotes

[^1729388462]: This has a mnemonic:


    | Letter | T9 |
    |--------|----|
    | d      | 3  |
    | r      | 7  |
    | a      | 2  |
    | f      | 3  |
    | t      | 8  |

