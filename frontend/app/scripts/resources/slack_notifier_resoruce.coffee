app = angular.module('harrowApp')

app.factory "SlackNotifier", ($injector, $http) ->
  SlackNotifier = (data) ->
    $.extend(true, @, data)
    @

  SlackNotifier

app.factory 'slackNotifierResource', (Resource, SlackNotifier) ->
  SlackNotifierResource = () ->
    Resource.call(@)
    @

  SlackNotifierResource:: = Object.create(Resource::)
  SlackNotifierResource::basepath = "/slack-notifiers"
  SlackNotifierResource::model = SlackNotifier
  new SlackNotifierResource()
