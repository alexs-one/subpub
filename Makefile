BINDIR=${CURDIR}/bin

PROTOC = PATH="$$PATH:$(BINDIR)" protoc

SUBPUB_PORTOC_PATH = "api"

# Установка grpc пакетов и генерация grpc proto
.PHONY: .bin-deps
.bin-deps:
	$(info Installing binary dependencies...)
	GOBIN=$(BINDIR) go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.28.1 && \
	GOBIN=$(BINDIR) go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.2.0


vendor-proto/google/api:
	git clone -b master --single-branch -n --depth=1 --filter=tree:0 \
 	https://github.com/googleapis/googleapis vendor-proto/googleapis && \
 	cd vendor-proto/googleapis && \
	git sparse-checkout set --no-cone google/api && \
	git checkout
	mkdir -p  vendor-proto/google
	mv vendor-proto/googleapis/google/api vendor-proto/google
	rm -rf vendor-proto/googleapis

.PHONY: .vendor-rm
.vendor-rm:
	rm -rf vendor-proto

.PHONY: .vendor-proto
.vendor-proto: .vendor-rm  vendor-proto/google/api


protoc-generate: .bin-deps .vendor-proto
	protoc \
	-I ${SUBPUB_PORTOC_PATH} \
	-I vendor-proto \
	--plugin=protoc-gen-go=${BINDIR}/protoc-gen-go \
	--go_out internal/pkg/${SUBPUB_PORTOC_PATH} \
	--go_opt paths=source_relative \
	--plugin=protoc-gen-go-grpc=${BINDIR}/protoc-gen-go-grpc \
	--go-grpc_out internal/pkg/${SUBPUB_PORTOC_PATH} \
	--go-grpc_opt paths=source_relative \
	${SUBPUB_PORTOC_PATH}/subpub.proto
	go mod tidy

#Установка grpccurl для проверки
install-grpccurl:
	GOBIN=$(BINDIR) go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest

#Запуск 
run:
	CONFIG_FILE="./config/local.yaml" go run ./cmd/server/main.go