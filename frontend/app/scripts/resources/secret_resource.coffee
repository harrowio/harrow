app = angular.module("harrowApp")

app.factory "Secret", ($injector) ->
  Secret = (data) ->
    $.extend(true, @, data)
    @environmentResource = $injector.get("environmentResource")
    @

  Secret::environment = ->
    @environmentResource.fetch @_links.environment.href

  Secret

app.factory "secretResource", (Resource, Secret) ->
  SecretResource = () ->
    Resource.call(@)
    @

  SecretResource:: = Object.create(Resource::)
  SecretResource::basepath = "/secrets"
  SecretResource::model = Secret

  new SecretResource()
