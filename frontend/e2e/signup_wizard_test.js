var expect = require('chai').expect
var helper = require('./helpers/core')
var wizard = require('./helpers/wizard')

describe('Happy Path', function () {
  this.timeout(15000); // Set timeout to 15 seconds, instead of default 2 seconds
  var nightmare
  before(function () {
    nightmare = helper.nightmare(this.test.fullTitle())
  })
  describe('Sign Up', function () {
    before(function () {
      wizard.createAccount(nightmare)
    })

    it('greets the user with the wizard', function (done) {
      nightmare
        .evaluate(function () {
          return document.querySelector('form.card.card--wizard .card__content h2').textContent
        })
        .then(function (heading) {
          expect(heading).to.equal('Hey Joe Bloggs')
          done()
        })
        .catch(done)
    })

    describe('Wizard', function () {
      before(function () {
        wizard.createOrganizationAndProject(nightmare)
      })

      it('creates a Organization and project', function (done) {
        nightmare
          .evaluate(function () {
            return document.querySelector('.card__content h3').textContent
          })
          .then(function (heading) {
            expect(heading).to.equal('Connect your code repository')
            done()
          }).catch(done)
      })

      describe('Repository', function () {
        before(function () {
          wizard.addRepository(nightmare)
        })

        it('shows stencil options', function (done) {
          nightmare
            .evaluate(function () {
              var out = {
                heading: document.querySelector('.card__content h3').textContent,
                stencils: []
              }
              var stencils = document.querySelectorAll('.stencil h4')
              for (var i = 0; i < stencils.length; i++) {
                out.stencils.push(stencils[i].textContent)
              }
              return out
            })
            .then(function (response) {
              expect(response.heading).to.equal('Choose your setup to populate Harrow with defaults')
              expect(response.stencils[0]).to.equal('Ruby on Rails with Capistrano')
              expect(response.stencils[1]).to.equal('Linux & Shell (Bash)')
              done()
            }).catch(done)
        })

        describe('Finished', function () {
          before(function () {
            wizard.addStencil(nightmare)
          })
          it('displays a congratulations screen', function (done) {
            nightmare
              .evaluate(function () {
                var out = {
                  heading: document.querySelector('.card__content h3').textContent
                }
                return out
              })
              .then(function (response) {
                expect(response.heading).to.equal('Success! You’ve created your first project')
                done()
              }).catch(done)
          })
        })
      })
    })
  })
})
