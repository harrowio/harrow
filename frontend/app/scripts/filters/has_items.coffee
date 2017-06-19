angular.module('harrowApp').filter 'hasItems', ->
  (obj) ->
    return false unless angular.isObject(obj)
    arr = []
    Object.keys(obj).forEach (key) ->
      arr.push obj[key].length
    arr.some (item) ->
      item > 0

angular.module('harrowApp').filter 'hasNoItems', ($filter) ->
  (obj) ->
    !$filter('hasItems')(obj)
