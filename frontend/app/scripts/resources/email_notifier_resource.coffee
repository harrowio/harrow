app = angular.module('harrowApp')

app.factory "EmailNotifier", ($injector, $http) ->
  EmailNotifier = (data) ->
    $.extend(true, @, data)
    @

  EmailNotifier

app.factory 'emailNotifierResource', (Resource, EmailNotifier) ->
  EmailNotifierResource = () ->
    Resource.call(@)
    @

  EmailNotifierResource:: = Object.create(Resource::)
  EmailNotifierResource::basepath = "/email-notifiers"
  EmailNotifierResource::model = EmailNotifier
  new EmailNotifierResource()
