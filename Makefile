# Build a development version and run a container
dev:
	@docker build -t sqrthree/progressbar201x-dev:latest -f Dockerfile.dev .
	@echo "==> Building"
	@echo "==> Run a container"
	@docker run -it --rm -v $(PWD):/go/src/github.com/sqrthree/progressbar201X -p 3000:3000 sqrthree/progressbar201x-dev /bin/bash
.PHONY: dev

# Build a production version image
build:
	@echo "==> Building production version"
	@docker build -t sqrthree/progressbar201x:latest .
	@echo "==> Complete"
.PHONY: build

start:
	@echo "==> Start a container with production version"
	@docker run --name progressbar201x -v $(PWD)/config.yml:/root/config.yml -v $(PWD)/article_template.html:/root/article_template.html -p 3000:3000 -d sqrthree/progressbar201x
	@echo "==> Done"
.PHONY: start
