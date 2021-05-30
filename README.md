# rainbow-road
Rainbow road is a server and client pair that will get the star count for an arbitrary amount of GitHub repos and return them to the client. The server can also be used as an API.

To run the project, first build and run the server using either go binaries, docker image or a docker-desktop kubernetes cluster using [tilt](https://tilt.dev/). Then build and use the client to make requests to the server.

## Client
### Building and Running
Build the client with the following make command:
```
make build-client // builds client
```
Set the RAINBOW_ROAD_SERVER environment variable. Regardless of how you choose to run your server from the section below, the address will be `http://localhost:9999`
```
export RAINBOW_ROAD_SERVER=http://localhost:9999
```
### Usage
The stars command takes in a list of GitHub repos and returns the amount of times that repo has been stared.
```
➜  rainbow-road git:(main) ✗ ./stars kubernetes/kubernetes istio/istio puppetlabs/puppet
REPO                                              STARS
kubernetes/kubernetes                             77649
istio/istio                                       27087
puppetlabs/puppet                                 6172
```
```
➜  rainbow-road git:(main) ✗ ./stars
Usage: stars <git-repo-1> <git-repo-2> ...
```
Errors, on a per repo basis are passed to the client and shown in the stars column:
```
➜  rainbow-road git:(main) ✗ ./stars kubernetes/kubernetes istio/itio
REPO                                              STARS
kubernetes/kubernetes                             77649
istio/itio                                        Repo Not found: istio/itio
```
Stars will not call the server if the user is making a malformed request:
```
➜  rainbow-road git:(main) ✗ ./stars kubernetes/kubernetes istio
Error: Invalid repo name istio. Hint: <org>/<repo-name>
```
  
## Server
### Building and Running
If you have a GitHub token, set the GITHUB_TOKEN environment variable. Regardless of how you choose to run your server, this variable will be sourced from here. If you do not have a GitHub token, you can still use rainbow-road, but will be rate limited by GitHub and receive non deterministic results. 
```
export GITHUB_TOKEN = <my github token>
```
#### Without Docker:

The following build commands are available from the root dir of the project:
```
make build // builds both the client and server binaries
make build-client // builds client
make build-server // builds server
```
After building the binaries start the server locally:
```
➜  rainbow-road git:(main) ✗ ./rainbow-road-server


Starting Rainbow Road Server!
            .
           ,O,
          ,OOO,
    'oooooOOOOOooooo'
      'OOOOOOOOOOO'
        'OOOOOOO'
        OOOO'OOOO
       OOO'   'OOO
      O'         'O

 Listening on port: 9999
```
#### Docker:
Build the docker image with the following make command.
```
make build-docker
```
Ensure you have set a GITHUB_TOKEN
```
export GITHUB_TOKEN = <my github token>
```
Run the docker image with the following make command. This will load the GitHub token and expose the server at `http://localhost:9999`
```
make run-docker
```
#### Docker Desktop Kubernetes Cluster w/ Tilt
Tilt is a tool used for local development on kubernetes that offers, hot reloading on file save among other things. Here we use it to create a local-dev kubernetes setup.
##### Prerequisites: 

 - Ensure you have a working [docker desktop kubernetes cluster](https://docs.docker.com/desktop/kubernetes/)
 - Download and install [Tilt](https://docs.tilt.dev/install.html)
 
Ensure you have set a GITHUB_TOKEN
```
export GITHUB_TOKEN = <my github token>
```
Start tilt with the following command. This will load the GitHub token into a secret and expose the server at `http://localhost:9999`
```
➜  rainbow-road git:(main) ✗ tilt up
Tilt started on http://localhost:10350/
v0.20.3, built 2021-05-14

(space) to open the browser
(s) to stream logs (--stream=true)
(t) to open legacy terminal mode (--legacy=true)
(ctrl-c) to exit
```
Press the (space bar) to open a web ui to get pod status, logs, build status etc. or press (t) to get a terminal ui.

To terminate the tilt session run:
```
➜  rainbow-road git:(main) ✗ tilt down
Beginning Tiltfile execution
Successfully loaded Tiltfile (21.81092ms)
Deleting kubernetes objects:
→ Deployment/rainbow-road-api
```

### Testing
Use the following make commands from the root directory to run tests:
```
make test // run all tests
make test-client // run client tests
make test-server // run server tests
```


### Metrics
Prometheus style metrics are served via `/metrics`

Currently there are standard go metrics and some http request metrics