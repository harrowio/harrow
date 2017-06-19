angular.module('harrowApp').filter 'conceal', ->
  (input, maxLength) ->
    output = []
    length = input.length
    length = maxLength if input.length >= maxLength
    for [1..length]
      output.push "●"
    output.join ''
