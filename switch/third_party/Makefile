REPO=swagger-ui.repo

swagger-ui: ${REPO}/dist update
	rm -rf $@
	cp -R $< $@
update: ${REPO}
	cd ${REPO} && git pull
${REPO}/dist:
	git clone https://github.com/swagger-api/swagger-ui ${REPO}

