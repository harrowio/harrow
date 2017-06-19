translate = angular.module("pascalprecht.translate")
translate.factory "harrowLoader", ($q, en_GB)->

  # chosen locale is given in options.key
  (options) ->
    deferred = $q.defer()
    deferred.resolve en_GB
    deferred.promise

app = angular.module("harrowApp")

app.config ($translateProvider) ->

  $translateProvider.useLoader "harrowLoader"

  $translateProvider.useSanitizeValueStrategy 'escapeParameters'

  $translateProvider.addInterpolation "$translateMessageFormatInterpolation"

  $translateProvider
    .preferredLanguage("en_GB")
    .fallbackLanguage("en_GB")
