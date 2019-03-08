


################# We build on this system #################
src/github.com/gorilla/websocket:
	GOPATH=$${PWD} go get -v ./src
./gopilot: src/github.com/gorilla/websocket
	#GOARCH=amd64
	GOPATH=$${PWD} CGO_ENABLED=0 GOOS=linux \
	go build -ldflags="-w -s" -o ./gopilot ./src
./gopilot.strip: ./gopilot
	strip gopilot
	touch $@
clean:
	@rm -vf ./gopilot ./gopilot.strip
	@rm -vfR ./src/github.com
	@rm -vfR ./src/gopkg.in
	@make -C deploy/arch clean
	@make -C deploy/docker clean

install:
	install --directory $${DESTDIR}/etc/gopilot
	install --directory $${DESTDIR}/usr/bin
	install ./gopilot $${DESTDIR}/usr/bin/gopilot

gitversion: src/plugins/core/gitversion.go
src/plugins/core/gitversion.go:
	@echo "package core" > $@
	@echo "const Gitversion = \"$$(git log -1 --pretty=format:%h)\"" >> $@
	@echo "const Gitdate = \"$$(git log -1 --pretty=format:%ai)\"" >> $@

squash:
	docker-squash -t gopilot:latest gopilot:latest



################# Run latest builded image #################
run:
	docker run -ti --rm \
	--entrypoint "" \
	gopilot:latest /app/gopilot -v
