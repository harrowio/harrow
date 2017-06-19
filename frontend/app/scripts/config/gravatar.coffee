app = angular.module("harrowApp")

app.config (gravatarServiceProvider) ->
  gravatarServiceProvider.defaults =
   "default": 'mm'
