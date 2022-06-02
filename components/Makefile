build:
	GOOS=linux GOARCH=amd64 go build -o ./bin/offline ./offline_job
	GOOS=linux GOARCH=amd64 go build -o ./bin/register ./register_container

docker: build
	docker build --rm --no-cache -f Dockerfile -t "$(IMG)"  .

push: docker
	docker push "$(IMG)"