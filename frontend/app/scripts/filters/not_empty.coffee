angular.module('harrowApp').filter 'notEmpty', ->
  (object) ->
    if angular.isObject(object)
      Object.keys(object).length > 0
    else if angular.isString(object)
      object.length > 0
    else
      false

angular.module('harrowApp').filter 'empty', ($filter) ->
  (value) ->
    !$filter('notEmpty')(value)
