BIN_DIR=deploy/srv/harrow/bin

deploy.tar: $(BIN_DIR)/harrow-notify-job
	tar cf $@ $$(basename $@ .tar) --transform="s!$$(basename $@ .tar)!!"

$(BIN_DIR)/harrow-notify-job: harrow-main $(BIN_DIR)
	cp $< $@

$(BIN_DIR):
	mkdir -p $@
