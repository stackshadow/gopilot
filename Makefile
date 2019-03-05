
.PHONY: src/plugins/core/gitversion.go

################# We build on this system #################
./gopilot:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
	go build -ldflags="-w -s" -o ./gopilot ./src
./gopilot.strip: ./gopilot
	strip gopilot
	touch $@
clean:
	@rm -vf ./gopilot.strip
	@rm -vf ./gopilot

################# build over iron dockerimage #################
build-docker-iron: gitversion
	docker run \
	--rm \
	-v "$$PWD/../src":/go/src \
	-v "$$PWD":/go/bin \
	-w /go/src iron/go:dev go build -o /go/bin/gopilot


################# build and create image #################
docker-alpine: gitversion
	docker build \
	--tag gopilot:latest \
	--file deploy/Dockerfile.alpine.multistage \
	.

################# buld arch image #################
docker-arch: gitversion gopilot.strip
	docker build \
	--tag gopilot:latest \
	--file deploy/Dockerfile.arch \
	.

gitversion: src/plugins/core/gitversion.go
src/plugins/core/gitversion.go:
	@echo "package core" > $@
	@echo "const Gitversion = \"$$(git log -1 --pretty=format:%h)\"" >> $@
	@echo "const Gitdate = \"$$(git log -1 --pretty=format:%ai)\"" >> $@

squash:
	docker-squash -t gopilot:latest gopilot:latest

tag:
	@docker tag gopilot:latest $${REPO}gopilot:$$(uname -m)-latest
	@docker tag gopilot:latest $${REPO}gopilot:$$(git log -1 --pretty=format:%h)

push:
	@echo "> Push $${REPO}gopilot:$$(uname -m)-latest"
	@docker push $${REPO}gopilot:$$(uname -m)-latest
	@echo "> Push $${REPO}gopilot:$$(git log -1 --pretty=format:%h)"
	@docker push $${REPO}gopilot:$$(git log -1 --pretty=format:%h)


################# Run latest builded image #################
run:
	docker run -ti --rm \
	--entrypoint "" \
	gopilot:latest /app/gopilot -v
