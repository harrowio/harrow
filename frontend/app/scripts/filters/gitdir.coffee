app = angular.module("harrowApp")

app.filter 'gitdir', ->
  (input) ->
    input.split("/").pop().replace(/\.git$/, '')
