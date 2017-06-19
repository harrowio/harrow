var exports = module.exports = {}
exports.createAccount = function createAccount (nightmare) {
  // create user account
  nightmare
    .click('.card--wizard .btn[href="#/a/signup"]')
    .wait('form[translation-root="forms.signup"]')
    .type('form.card.card--wizard input[name="name"]', 'Joe Bloggs')
    .type('form.card.card--wizard input[name="email"]', 'test-' + new Date().getTime() + '@harrow.io')
    .type('form.card.card--wizard input[name="password"]', 'changeme123')
    .type('form.card.card--wizard input[name="passwordConfirmation"]', 'changeme123')
    .click('form.card.card--wizard input[name="acceptTermsAndConditions"]')
    .click('form.card.card--wizard .btn--primary')
    .wait('form[translation-root="forms.wizard.create"]')
}
exports.createOrganizationAndProject = function createOrganizationAndProject (nightmare) {
  // create organization and project
  nightmare
    .type('.card--wizard input[name="organizationName"]', 'ACME Corp')
    .type('.card--wizard input[name="projectName"]', 'Meep Meep App')
    .click('.card--wizard .btn--primary')
    .wait('form[translation-root="forms.wizard.connect"]')
}
exports.addRepository = function addRepository (nightmare) {
  // add respository
  nightmare
    .type('form[translation-root="forms.wizard.connect"] input[name="url"]', 'https://github.com/capistrano/capistrano.git')
    .click('form[translation-root="forms.wizard.connect"] .btn--primary')
    .wait('.stencil')
}
exports.addStencil = function addStencil (nightmare) {
  // add stencils
  nightmare
    .click('.stencil:nth-of-type(2)')
    .wait('h3[translate="forms.wizard.finished.header"]')
}
