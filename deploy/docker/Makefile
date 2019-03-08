.PHONY: src/plugins/core/gitversion.go docker-build-context.tar.gz

docker-build-context.tar.gz:
	tar \
	-C ../../ \
	-caf $@ \
	./deploy/docker \
	./src/core \
	./src/plugins \
	./src/copilotg.go

docker-alpine: docker-build-context.tar.gz
	docker build \
	--tag gopilot:latest \
	--file deploy/docker/alpine.multistage.Dockerfile \
	- < docker-build-context.tar.gz

clean:
	@rm -vf *.gz

tag:
	@docker tag gopilot:latest $${REPO}gopilot:$$(uname -m)-latest
	@docker tag gopilot:latest $${REPO}gopilot:$$(git log -1 --pretty=format:%h)

push:
	@echo "> Push $${REPO}gopilot:$$(uname -m)-latest"
	@docker push $${REPO}gopilot:$$(uname -m)-latest
	@echo "> Push $${REPO}gopilot:$$(git log -1 --pretty=format:%h)"
	@docker push $${REPO}gopilot:$$(git log -1 --pretty=format:%h)
