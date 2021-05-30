build:
	go build -o rainbow-road-server server/server.go && \
	go build -o stars client/client.go
	echo "built ./rainbow-road-server and ./stars (client)"
build-server:
	go build -o rainbow-road-server server/server.go
	echo "built ./rainbow-road-server"
build-client:
	go build -o stars client/client.go
	echo "built ./stars (client)"
build-docker:
	docker build . -t rainbow-road:latest
run-docker:
	docker run -p 9999:9999 --env GITHUB_TOKEN=$GITHUB_TOKEN rainbow-road:latest
test:
	go test ./client -v && go test ./server -v	
test-client:
	go test ./client -v
test-server:
	go test ./server -v
