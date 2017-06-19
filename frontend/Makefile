dist:
	script/dist

clean:
	rm -rf dist .tmp

unit:
	npm test

e2e:
	script/e2e
# Deploy a version to the given environment(s)
#
# Example:
#
#  make deploy-staging version=cafebabe
deploy-%:
	script/deploy "$(version)" $(subst :, ,$*)

generate:
ifndef $(component)
	mkdir -p app/scripts/components/$(component)
	touch app/scripts/components/$(component)/$(component).coffee app/scripts/components/$(component)/$(component)_test.js touch app/scripts/components/$(component)/$(component).html touch app/scripts/components/$(component)/$(component).scss
endif

.PHONY: watch dist clean unit e2e deploy
