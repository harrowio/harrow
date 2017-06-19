require '../bower_components/style-guide/source/css/style.scss'
require './styles/main.scss'

require './polyfills'
require './vendor'

global.Lom = require './scripts/lom_bundle'

templates = require.context('./views', true, /\.html$/);
templates.keys().forEach (key) ->
  templates(key)

require './scripts/app'



req = require.context('./scripts/controllers', true, /^(?!.*_test).*$/)
req.keys().forEach (key) ->
  req(key)
req = require.context('./scripts/config', true, /^(?!.*_test).*$/)
req.keys().forEach (key) ->
  req(key)
req = require.context('./scripts/filters', true, /^(?!.*_test).*$/)
req.keys().forEach (key) ->
  req(key)
req = require.context('./scripts/resources', true, /^(?!.*_test).*$/)
req.keys().forEach (key) ->
  req(key)
req = require.context('./scripts/services', true, /^(?!.*_test).*$/)
req.keys().forEach (key) ->
  req(key)
req = require.context('./scripts/controllers', true, /^(?!.*_test).*$/)
req.keys().forEach (key) ->
  req(key)
req = require.context('./scripts/directives', true, /^(?!.*_test).*$/)
req.keys().forEach (key) ->
  req(key)
req = require.context('./scripts/components', true, /^(?!.*_test).*$/)
req.keys().forEach (key) ->
  req(key)
