.DEFAULT_GOAL := docker
BIN=azure-loganalytics-python

docker:
	docker build -t turbonomic/$(BIN) --build-arg GIT_COMMIT=$(shell git rev-parse --short HEAD) .


