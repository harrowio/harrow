angular.module('harrowApp').filter 'orderObjectBy', ($filter) ->
  (items, field, reverse) ->
    collection = []
    Object.keys(items).forEach (key, index) =>
      collection[index] = items[key]
    $filter('orderBy')(collection, field)
