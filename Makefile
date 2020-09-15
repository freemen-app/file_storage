include .env
export

.PHONY: build
build:
	go build -v -race ./src/main.go


.PHONY: app_shell
app_shell:
	docker exec -it creative_hunter_app bash
