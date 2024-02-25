<!-- LTeX: language=en-US -->
# itchgrep
A tool for searching [itch.io](https://itch.io/) assets by name and
description.

## Running Locally
### Dependencies
- [Golang](https://go.dev/)
- [Task](https://taskfile.dev/)
- [Docker](https://www.docker.com/) (for running a local instance of DynamoDB)

### Running
The project is split up into two services:
- The `dataservice`, responsible for fetching the list of assets from [itch.io](https://itch.io/)
- The `webserver`, presenting the stored data with search tools.


Use the included [Taskfile](https://taskfile.dev/) to run these services.
- `task dataservice-local` will launch the `dataservice` with a local instance
    of DynamoDB. It will fetch all available assets on itch.io, store them in the
    local DynamoDB database. The data is persisted as an `sqlite3` file called
    `./local_data/shared-local-instance.db`.
- `task explore-data` will drop you into an `sqlite3` shell running through the
    DynamoDB Docker container. You do not need to have `sqlite3` installed for
    this.

## Testing
### Dependencies (in addition to build dependencies):
None so far :)

### Running Tests
Tests can be run by using the included [Taskfile](https://taskfile.dev/).

The task `test` runs all tests at once.
You can also run specific tests on their own:
- `test_db`: Tests the `db` package, requires `Docker` to be running.

## Techstack
- Golang (Dataservice, Web View)
- Templ (Web View)
- HTMX (Web View)
- DynamoDB (Storage)
