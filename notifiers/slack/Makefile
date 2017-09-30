BIN_DIR=deploy/srv/harrow/bin

deploy.tar: $(BIN_DIR)/harrow-notify-slack
	tar cf $@ $$(basename $@ .tar) --transform="s!$$(basename $@ .tar)!!"

$(BIN_DIR)/harrow-notify-slack: harrow-main $(BIN_DIR)
	cp $< $@

$(BIN_DIR):
	mkdir -p $@
