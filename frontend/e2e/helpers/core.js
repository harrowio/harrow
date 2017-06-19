var Nightmare = require('nightmare')
var execSync = require('child_process').execSync
var request = require('request')
var measureDebug = require('debug')('e2e:benchmark')
var dbDebug = require('debug')('e2e:db')
var timer = null
Nightmare.action('startTimer', function (done) {
  timer = new Date()
  done()
})
Nightmare.action('measure', function (done) {
  var time = new Date() - timer
  measureDebug('Took %sms', time)
  done(time)
})

var wizard = require('./wizard')

var exports = module.exports = {}

exports.endpoint = 'https://www.vm.harrow.io'

exports.seedDB = function () {
  dbDebug('starting clean and seed')
  var cwd = process.env.HARROW_PATH + '/api'
  execSync('vagrant ssh default -c "make -C /harrow/api cleardb seed"', {
    cwd: cwd
  })
  dbDebug('seed complete')
}

exports.clearDB = function () {
  dbDebug('starting clean only')
  var cwd = process.env.HARROW_PATH + '/api'
  execSync('vagrant ssh default -c "make -C /harrow/api cleardb"', {
    cwd: cwd
  })
  dbDebug('clean complete')
}

exports.addKeyToBitBucket = function (key) {
  var username = process.env.HARROW_BITBUCKET_API_USER
  var token = process.env.HARROW_BITBUCKET_API_TOKEN
  if (!username || !token) {
    throw new Error('Environment variables $HARROW_BITBUCKET_USERNAME or $HARROW_BITBUCKET_TOKEN are not defined.')
  }
  request.post(
    'https://' + username + ':' + token + '@api.bitbucket.org/1.0/repositories/harrowio/mdpf-intergration-test/deploy-keys', {
      formData: {
        label: 'e2e-test-' + new Date().getTime(),
        key: key
      }
    }
  )
}

exports.nightmare = function (partition, size) {
  if (!size) {
    size = 'desktop'
  }
  if (size.includes(['mobile', 'tablet', 'desktop'])) {
    throw new Error('viewport size string must be (`mobile`, `tablet`, or `desktop`)')
  }
  partition = encodeURIComponent(partition.replace(/\s/g, '-'))

  var nightmare = Nightmare({
    show: process.env.E2E_SHOW ? true : false,
    webPreferences: {
      partition: partition
    }
  })
  if (size === 'mobile') {
    nightmare.viewport(320, 568) // iPhone 5 Vertical
  } else if (size === 'tablet') {
    nightmare.viewport(768, 1024) // iPad Vertical
  } else {
    nightmare.viewport(1024, 768)
  }
  nightmare
    .goto(this.endpoint)
    .wait('.ng-scope')
  return nightmare
}

exports.login = function (nightmare, email, password) {
  nightmare
    .type('input[name="email"]', email)
    .type('input[name="password"]', password)
    .click('.card__footer .btn--primary')
    .wait('.sectionHeader')
}

exports.createAccountWithOrganizationAndProjectWithRepositoryAndStencil = function (n) {
  wizard.createAccount(n)
  wizard.createOrganizationAndProject(n)
  wizard.addRepository(n)
  wizard.addStencil(n)
  n
    .click('.card--wizard .btn--primary')
    .wait('.sectionHeader')
}

exports.createProjectSkippingRepositoryAndStencil = function (nightmare, projectName) {
  nightmare
    .click('a[ui-sref="dashboard"]')
    .wait('.organization')
    .click('.organization:nth-of-type(1) a.project.project--empty[harrow-can="create-projects"]')
    .wait('form[translation-root="forms.wizard.create"]')
    .type('.card--wizard input[name="projectName"]', projectName)
    .click('.card--wizard .btn--primary')
    .wait('form[translation-root="forms.wizard.connect"]')
    .click('.card__footer a.btn:not([ng-click])')
    .wait('.stencil')
    .click('.card__footer a.btn:not([ng-click])')
    .wait('h3[translate="forms.wizard.finished.header"]')
    .click('.card--wizard .btn--primary')
    .wait('.sectionHeader')
}
