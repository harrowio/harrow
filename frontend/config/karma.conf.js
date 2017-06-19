var webpackConfig = require('./webpack.test')
var customLaunchers = {
  sl_chrome: {
    base: 'SauceLabs',
    browserName: 'Chrome',
    timeZone: 'London'
  },
  sl_firefox: {
    base: 'SauceLabs',
    browserName: 'Firefox',
    timeZone: 'London'
  },
  sl_ie: {
    base: 'SauceLabs',
    browserName: 'Internet Explorer',
    platform: 'Windows 10',
    timeZone: 'London'
  },
  sl_ie_edge: {
    base: 'SauceLabs',
    browserName: 'MicrosoftEdge',
    platform: 'Windows 10',
    timeZone: 'London'
  }
}

module.exports = function (config) {
  var _config = {
    basePath: '',
    frameworks: ['jasmine'],
    files: [
      { pattern: './config/test-bundle.js', watched: false }
    ],
    port: 9876,
    logLevel: config.LOG_INFO,
    browserNoActivityTimeout: 60000,
    // customLaunchers: customLaunchers,
    sauceLabs: {
      testName: 'Web App Karma Tests',
      username: 'harrowio',
      accessKey: '•••••••••••••••••'
    },
    autoWatch: false,
    singleRun: true,
    colors: true,
    preprocessors: {
      './config/test-bundle.js': ['webpack', 'sourcemap']
    },
    webpack: webpackConfig,
    webpackMiddleware: {
      stats: 'errors-only'
    },
    webpackServer: {
      noInfo: true
    },
    browsers: [
      'Electron',
      'Chrome'
    ], // .concat(Object.keys(customLaunchers)),
    electronOpts: {
      show: false
    },
    reporters: [
      'spec',
      'coverage' // ,
    // 'saucelabs'
    ],
    coverageReporter: {
      reporters: [
        {
          type: 'html',
          dir: 'coverage/'
        }, {
          type: 'cobertura',
          dir: 'coverage/'
        }
      ]
    },
    specReporter: {
      maxLogLines: 5,
      suppressErrorSummary: false,
      suppressPassed: false,
      suppressFailed: false,
      suppressSkipped: true, // config.singleRun == false,
      showSpecTiming: true
    }
  }
  config.set(_config)
}
