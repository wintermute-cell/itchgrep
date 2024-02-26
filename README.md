<!-- LTeX: language=en-US -->
# itchgrep
A tool that helps you find more and better assets for your games.
It enables searching [itch.io](https://itch.io/) game assets with text queries
instead of by tags.

You can find this service hosted on [TODO](todo).

## Running Locally

If you want to [contribute](#contributing), or just run the project locally for your own use,
follow the instructions below.

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
- ![Go](https://img.shields.io/badge/Go-00ADD8?style=for-the-badge&logo=go&logoColor=white) Go
- ![Templ](https://img.shields.io/badge/Templ-000000?style=for-the-badge&logo=templ&logoColor=white) Templ
- ![HTMX](https://img.shields.io/badge/HTMX-FF5733?style=for-the-badge&logo=htmx&logoColor=white) HTMX
- ![Google Cloud Storage](https://img.shields.io/badge/Google_Cloud_Storage-4285F4?style=for-the-badge&logo=google-cloud&logoColor=white) Google Cloud Storage


## Contributing
TODO, Outline:
- major features -> ask first as issue
- use go fmt to format code
- beginners welcome, take your time, ask if unsure
- feature request as issues
- keep contribution related discussion in issues, no other comms channels

