# rainbow-road

## Server

### Build
```
docker build . -t rainbow-road:latest
```

### Run
```
export GIT_TOKEN=<ya git token>
docker run -p 9999:9999 --env GIT_TOKEN=$GIT_TOKEN rainbow-road:latest
```

### Test
```
cd server
go test
```