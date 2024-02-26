<!-- LTeX: language=en-US -->
# itchgrep
A tool for searching [itch.io](https://itch.io/) assets by name and
description.

## Running Locally

> This project is built and maintained on Linux. While I don't think it's
> generally impossible to run on Windows, but the
> [Taskfile](https://taskfile.dev/) is written using Linux commands.

### Dependencies
- [Golang](https://go.dev/)
- [Task](https://taskfile.dev/)
- [Docker](https://www.docker.com/) (for running a local fake instance of GCS (Google Cloud Storage))

### Running
The project is split up into two services:
- The `dataservice`, responsible for fetching the list of assets from [itch.io](https://itch.io/)
- The `webserver`, presenting the stored data with search tools.

Use the included [Taskfile](https://taskfile.dev/) to run these services.
- `task local-dataservice` will launch the `dataservice` with a local instance
    of GCS. It will fetch all available assets on itch.io, store them in the
    local GCS. The data is persisted as a `.json` file called
    `./local_data/itchgrep-data/assets.json`.
- `task local-webserver` will build and run the web server in a Docker
    container together with the local GCS in a separate container. `Templ`
    templates are not copied during the build, but generated inside the
    container.
- `task templ` will generate `.go` files from any `.templ` files. This is not
    required for building/running, but to provide code completion and stop the
    language server from complaining.

## Testing
### Dependencies (in addition to build dependencies):
None so far :)

### Running Tests
Tests can be run by using the included [Taskfile](https://taskfile.dev/).

- `task test`: Runs all of the test tasks below.
- `task test-storage`: Tests the `storage` package, requires `Docker` to be running.

## Techstack
- Golang (Dataservice, Web View)
- Templ (Web View)
- HTMX (Web View)
- DynamoDB (Storage)
