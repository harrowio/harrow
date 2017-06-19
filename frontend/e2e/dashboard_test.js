var expect = require('chai').expect
var helper = require('./helpers/core')

describe('Dashboard', function () {
  this.timeout(15000)
  var nightmare
  before(function () {
    nightmare = helper.nightmare(this.test.fullTitle())
    helper.login(nightmare, 'vagrant@localhost', 'vagrant')
  })

  it('has a project and tasks', function (done) {
    nightmare
      .evaluate(function () {
        var response = {}
        response.organizationName = document.querySelector('.organization .sectionHeader__title').textContent
        response.projectName = document.querySelector('.project h3').textContent
        return response
      })
      .then(function (response) {
        expect(response.organizationName).to.equal('Harrow')
        expect(response.projectName).to.equal('Harrow API')
        done()
      })
      .catch(done)
  })
})
