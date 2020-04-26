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
	ENV=development SECRET_KEY=asdf7q97tgpdr9y8t4990 DATABASE_URL="host=postgres user=root dbname=todo_app password=root sslmode=disable" REDIS_URL="redis:6379" ./vue-go-todo

run:
	go build
	SECRET_KEY=asdf7q97tgpdr9y8t4990 DATABASE_URL="host=postgres user=root dbname=todo_app password=root sslmode=disable" REDIS_URL="redis:6379" ./vue-go-todo

clean:
	go clean
