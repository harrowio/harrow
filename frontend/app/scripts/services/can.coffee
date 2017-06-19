Can = (
  @$filter
) ->
  @

Can::can = (can, subject) ->
  [action, relname] = can.split '-'
  relname = "self" unless relname
  relname = @$filter('dashCase')(relname)
  return subject._links?[relname]?.hasOwnProperty(action) || false

angular.module('harrowApp').service 'can', Can
