app = angular.module("harrowApp")

app.filter 'empty', ->
  (input) ->
    $.isEmptyObject(input)
