.PHONY: start build

NOW = $(shell date -u '+%Y%m%d%I%M%S')


APP             = server
SERVER_BIN  	= ./cmd/${APP}/${APP}
RELEASE_ROOT 	= release
RELEASE_SERVER 	= release/${APP}


ifeq ($(GOHOSTOS), windows)
	#the `find.exe` is different from `find` in bash/shell.
	#to see https://docs.microsoft.com/en-us/windows-server/administration/windows-commands/find.
	#changed to use git-bash.exe to run find cli or other cli friendly, caused of every developer has a Git.
	#Git_Bash= $(subst cmd\,bin\bash.exe,$(dir $(shell where git)))
	Git_Bash=$(subst \,/,$(subst cmd\,bin\bash.exe,$(dir $(shell where git))))
	INTERNAL_PROTO_FILES=$(shell $(Git_Bash) -c "find internal -name *.proto")
	API_PROTO_FILES=$(shell $(Git_Bash) -c "find api -name *.proto")
else
	INTERNAL_PROTO_FILES=$(shell find internal -name *.proto)
	API_PROTO_FILES=$(shell find api -name *.proto)
endif

all: start

.PHONY: api
api:
	protoc --proto_path=./third_party      \
		   --proto_path=./api             \
		   --go_out=.                \
		   --go-gin_out=.            \
		   --go_error_out=.          \
		   --openapi_out=fq_schema_naming=true,default_response=false:.  \
		   $(INTERNAL_PROTO_FILES)

build:
	@go build -ldflags "-w -s" -o $(SERVER_BIN) ./cmd/${APP}/main.go

start:
	@go run ./cmd/${APP}/main.go

test:
	cd ./test && go test -v

clean:
	rm -rf $(SERVER_BIN)

docker:
	docker build -t ${APP}:v1 .

