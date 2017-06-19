angular.module('harrowApp').service 'later', ($window) ->
  if angular.isDefined($window.later)
    $window.later
    
