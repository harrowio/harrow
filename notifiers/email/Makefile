BIN_DIR=deploy/srv/harrow/bin
TEMPLATE_DIR=deploy/srv/harrow/config/mail

deploy.tar: $(TEMPLATE_DIR)/operation $(BIN_DIR)/harrow-notify-email
	tar cf $@ $$(basename $@ .tar) --transform="s!$$(basename $@ .tar)!!"

$(BIN_DIR)/harrow-notify-email: harrow-main $(BIN_DIR)
	cp $< $@

$(BIN_DIR):
	mkdir -p $@

$(TEMPLATE_DIR):
	mkdir -p $@

$(TEMPLATE_DIR)/operation: $(TEMPLATE_DIR) $(shell find ../../mail/operations -name '*.tmpl')
	-rm -r $@
	-mkdir -p $$(dirname $@)
	cp -rv ../../mail/operations $@
