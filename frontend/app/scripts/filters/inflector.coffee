app = angular.module('harrowApp')

app.filter 'dashCase', (inflector) ->
  (input) ->
    return inflector.parameterize(inflector.dasherize(input, '-'), '-')
app.filter 'dasherize', (inflector) ->
  inflector.dasherize
app.filter 'singularize', (inflector) ->
  inflector.singularize
app.filter 'parameterize', (inflector) ->
  inflector.parameterize
app.filter 'pluralize', (inflector) ->
  inflector.pluralize
app.filter 'camelize', (inflector) ->
  inflector.camelize
app.filter 'camelCase', (inflector) ->
  inflector.camelize
app.filter 'underscoreCase', (inflector) ->
  (input) ->
    return inflector.parameterize(inflector.dasherize(input, '_'), '_')
app.filter 'ordinal', ->
  (input) ->
    return input unless angular.isNumber(input)
    suffix = if Math.floor(input / 10) == 1
      'th'
    else if input % 10 == 1
      'st'
    else if input % 10 == 2
      'nd'
    else if input % 10 == 3
      'rd'
    else
      'th'
    "#{input}#{suffix}"
