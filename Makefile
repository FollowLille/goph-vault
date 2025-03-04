VERSION = 1.0.0
BUILD_DATE = $(shell powershell -Command Get-Date -Format "yyyy-MM-dd")

# Платформы для сборки
PLATFORMS = windows linux darwin
ARCHITECTURES = amd64

build:
	@echo "Building GophVault..."
	$(foreach GOOS, $(PLATFORMS), \
		$(foreach GOARCH, $(ARCHITECTURES), \
			$(shell set GOOS=$(GOOS) & set GOARCH=$(GOARCH) & \
			go build -ldflags="-X github.com/FollowLille/goph-vault/internal/commands.version=$(VERSION) -X github.com/FollowLille/goph-vault/internal/commands.buildDate=$(BUILD_DATE)" \
			-o bin/goph-vault-$(GOOS)-$(GOARCH)$(if $(filter windows, $(GOOS)),.exe,) cmd/client/main.go) \
		) \
	)
	@echo "Build completed!"

clean:
	@echo "Cleaning up..."
	rm -rf bin/
	@echo "Cleanup completed!"

run: build
	@echo "Running GophVault..."
	./bin/goph-vault-$(shell go env GOOS)-$(shell go env GOARCH)$(if $(filter windows, $(shell go env GOOS)),.exe,)

test:
	@echo "Running tests..."
	go test ./...
	@echo "Tests completed!"