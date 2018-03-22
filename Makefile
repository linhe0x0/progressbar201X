# Build a development version and run a container
start:
	@echo "==> Building"
	@docker build -t sqrthree/progressbar201x-dev:latest -f Dockerfile.dev .
	@echo "==> Run a container"
	@docker run -it --rm -v $(PWD):/go/src/github.com/sqrthree/progressbar201X -p 3000:3000 sqrthree/progressbar201x-dev /bin/bash
.PHONY: start

# Build a production version image
build:
	@echo "==> Building production version"
	@docker build -t sqrthree/progressbar201x:latest .
	@echo "==> Complete"
.PHONY: build
