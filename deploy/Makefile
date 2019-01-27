

################# We build on this system #################
build-host:
	CGO_ENABLED=0 GOOS=linux go build -o ./gopilot ../src
gopilot.strip: build-host
	strip gopilot
	touch $@
clean:
	rm -f gopilot.strip
	rm -f gopilot
	rm -fR ./src

################# build over iron dockerimage #################
build-docker-iron:
	docker run \
	--rm \
	-v "$$PWD/../src":/go/src \
	-v "$$PWD":/go/bin \
	-w /go/src iron/go:dev go build -o /go/bin/gopilot


################# build and create image #################
./src:
	cp -R ../src $@
image-alpine: ./src
	docker build \
	--tag gopilot:latest \
	--file Dockerfile.scratch \
	.

################# buld arch image #################
image-arch: gopilot.strip
	docker build \
	--tag gopilot:latest \
	--file Dockerfile.arch \
	--no-cache \
	.

################# Run latest builded image #################
run:
	docker run -ti --rm \
	--entrypoint "" \
	gopilot:latest /app/gopilot -v
