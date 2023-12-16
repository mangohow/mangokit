user	:=	$(shell whoami)
rev 	:= 	$(shell git rev-parse --short HEAD)
os		:=	$(shell expr substr $(shell uname -s) 1 5)

# GOBIN > GOPATH > INSTALLDIR
# Mac OS X
ifeq ($(shell uname),Darwin)
GOBIN	:=	$(shell echo ${GOBIN} | cut -d':' -f1)
GOPATH	:=	$(shell echo $(GOPATH) | cut -d':' -f1)
endif

# Linux
ifeq ($(os),Linux)
GOBIN	:=	$(shell echo ${GOBIN} | cut -d':' -f1)
GOPATH	:=	$(shell echo $(GOPATH) | cut -d':' -f1)
endif

# Windows
ifeq ($(os),MINGW)
GOBIN	:=	$(subst \,/,$(GOBIN))
GOPATH	:=	$(subst \,/,$(GOPATH))
GOBIN :=/$(shell echo "$(GOBIN)" | cut -d';' -f1 | sed 's/://g')
GOPATH :=/$(shell echo "$(GOPATH)" | cut -d';' -f1 | sed 's/://g')
endif
BIN		:= 	""

TOOLS_SHELL="./hack/tools.sh"
# golangci-lint
LINTER := bin/golangci-lint

# check GOBIN
ifneq ($(GOBIN),)
	BIN=$(GOBIN)
else
# check GOPATH
	ifneq ($(GOPATH),)
		BIN=$(GOPATH)/bin
	endif
endif

all:
	@cd cmd/mangokit && go build && cd - &> /dev/null
	@cd cmd/protoc-gen-go-error && go build && cd - &> /dev/null
	@cd cmd/protoc-gen-go-gin && go build && cd - &> /dev/null

.PHONY: install
install: all
ifeq ($(user),root)
#root, install for all user
	@cp ./cmd/mangokit/mangokit /usr/bin
	@cp ./cmd/protoc-gen-go-error/protoc-gen-go-error /usr/bin
	@cp ./cmd/protoc-gen-go-gin/protoc-gen-go-gin /usr/bin
else
#!root, install for current user
	$(shell if [ -z '$(BIN)' ]; then read -p "Please select installdir: " REPLY; mkdir -p $${REPLY};\
	cp ./cmd/mangokit/mangokit $${REPLY}/;cp ./cmd/protoc-gen-go-error/protoc-gen-go-error $${REPLY}/;cp ./cmd/protoc-gen-go-gin/protoc-gen-go-gin $${REPLY}/;else mkdir -p '$(BIN)';\
	cp ./cmd/mangokit/mangokit '$(BIN)';cp ./cmd/protoc-gen-go-error/protoc-gen-go-error '$(BIN)';cp ./cmd/protoc-gen-go-gin/protoc-gen-go-gin '$(BIN)'; fi)
endif
	@which protoc-gen-go &> /dev/null || go get google.golang.org/protobuf/cmd/protoc-gen-go
	@which protoc-gen-go-grpc &> /dev/null || go get google.golang.org/grpc/cmd/protoc-gen-go-grpc
	@which protoc-gen-validate  &> /dev/null || go get github.com/envoyproxy/protoc-gen-validate
	@echo "install finished"

.PHONY: uninstall
uninstall:
	$(shell for i in `which -a kratos | grep -v '/usr/bin/mangokit' 2>/dev/null | sort | uniq`; do read -p "Press to remove $${i} (y/n): " REPLY; if [ $${REPLY} = "y" ]; then rm -f $${i}; fi; done)
	$(shell for i in `which -a protoc-gen-go-grpc | grep -v '/usr/bin/protoc-gen-go-error' 2>/dev/null | sort | uniq`; do read -p "Press to remove $${i} (y/n): " REPLY; if [ $${REPLY} = "y" ]; then rm -f $${i}; fi; done)
	$(shell for i in `which -a protoc-gen-validate | grep -v '/usr/bin/protoc-gen-go-error' 2>/dev/null | sort | uniq`; do read -p "Press to remove $${i} (y/n): " REPLY; if [ $${REPLY} = "y" ]; then rm -f $${i}; fi; done)
	@echo "uninstall finished"

