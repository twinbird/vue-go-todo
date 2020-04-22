vue-go-todo:
	go build

.PHONY:
	run
	clean
	dev

dev:
	go build
	PORT=8080 SECRET_KEY=asdf7q97tgpdr9y8t4990 ./vue-go-todo

run:
	go build
	PORT=8080 ENV=production SECRET_KEY=asdf7q97tgpdr9y8t4990 ./vue-go-todo

clean:
	go clean
