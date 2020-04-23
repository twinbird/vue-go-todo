onDockerDev:
	docker-compose exec app /bin/bash -c "cd /go/src/github.com/twinbird/vue-go-todo/ && make dev"

vue-go-todo:
	go build

.PHONY:
	run
	clean
	dev
	onDockerDev

dev:
	go build
	PORT=80 SECRET_KEY=asdf7q97tgpdr9y8t4990 ./vue-go-todo

run:
	go build
	PORT=80 ENV=production SECRET_KEY=asdf7q97tgpdr9y8t4990 ./vue-go-todo

clean:
	go clean
