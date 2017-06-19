Error.stackTraceLimit = Infinity

require('core-js/es6')

require('../app/main')

require('angular-mocks/angular-mocks')

require('../test/lib/test_helper')

var appContext = require.context('../app', true, /^.*_(test|spec)\.(js|ts|coffee)$/)
appContext.keys().forEach(appContext)

var legacyAppContext = require.context('../test/spec', true)
legacyAppContext.keys().forEach(legacyAppContext)
