<!-- LTeX: language=en-US -->
# itchgrep
A tool that helps you find more and better assets for your games.
It enables searching [itch.io](https://itch.io/) game assets with text queries
instead of by tags.

You can find this service hosted on [TODO](todo).

If you want to show some love for what I do, or want help pay the server, you can...<br>
[!["Buy Me A Coffee"](https://www.buymeacoffee.com/assets/img/custom_images/yellow_img.png)](https://www.buymeacoffee.com/winterv)

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
Tests can be run by using the included [Taskfile](https://taskfile.dev/).

- `task test`: Runs all of the test tasks below.
- `task test-storage`: Tests the `storage` package, requires `Docker` to be running.

## Techstack
![Go](https://img.shields.io/badge/Go-00ADD8?style=for-the-badge&logo=go&logoColor=white)
![Templ](https://img.shields.io/badge/Templ-000000?style=for-the-badge&logo=templ&logoColor=white)
![HTMX](https://img.shields.io/badge/HTMX-FF5733?style=for-the-badge&logo=htmx&logoColor=white)
![Google Cloud](https://img.shields.io/badge/Google_Cloud-4285F4?style=for-the-badge&logo=google-cloud&logoColor=white)
![Docker](https://img.shields.io/badge/Docker-2496ED?style=flat-square&logo=docker&logoColor=white)

## Contributing
- before posting a pull request, please use [`go fmt`](https://go.dev/blog/gofmt) to format your code.
- beginners to open source are welcome. if you'd like to contribute, but don't
    understand something, you're welcome to ask using an issue.
- please post feature requests as one issue per feature.
- before working on a larger contribution, please open an issue to ask if the
    feature you want to implement would be welcome.
- to maintain a transparent workflow, please keep all discourse regarding work
    on this repository in the github issues, don't message me through other
    channels to discuss this.
