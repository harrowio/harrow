var expect = require('chai').expect
var helper = require('./helpers/core')

describe('Repository', function () {
  this.timeout(60000)
  var nightmare
  before(function () {
    nightmare = helper.nightmare(this.test.fullTitle())
    helper.login(nightmare, 'vagrant@localhost', 'vagrant')
  })

  describe('Add SSH Repository', function () {
    it('successfully adds SSH repository', function (done) {
      var step1 = nightmare
        .click('.project:nth-of-type(1) .project__content')
        .wait('.app__content.project__tasks')
        .click('.layout__header a.tab[ui-state="ctrl.editItem.stateName"]')
        .wait('form[translation-root="forms.project"]')
        .click('a.navigation__item[href$="repositories"]')
        .wait('a[ui-sref="createRepository"]')
        .click('a[ui-sref="createRepository"]')
        .wait('form[translation-root="forms.wizard.connect"]')
        .type('input[name="url"]', 'git@bitbucket.com:harrowio/mdpf-intergration-test.git')
        .click('.card__footer .btn--primary')
        .wait('[ng-if="ctrl.credential.subject.publicKey"]')
        .evaluate(function () {
          var response = {}
          response.sshKey = document.querySelector('[ng-if="ctrl.credential.subject.publicKey"] pre').textContent
          return response
        })
        .then(function (response) {
          helper.addKeyToBitBucket(response.sshKey)
        })
      step1.then(function () {
        nightmare
          .click('.card__footer .btn--primary')
          .wait('[ng-repeat="repository in ctrl.repositories"]')
          .evaluate(function () {
            var response = {}
            response.status = document.querySelector('[ng-repeat="repository in ctrl.repositories"]:nth-of-type(1) .card__item__icon svg').getAttribute('svg-icon').replace('icon-', '')
            return response
          })
          .then(function (response) {
            expect(response.status).to.equal('complete')
            done()
          })
      }).catch(done)
    })
  })
})
