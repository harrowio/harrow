app = angular.module("harrowApp")

app.factory "RepositoryCredential", ($injector) ->
  RepositoryCredential = (data) ->
    $.extend(true, @, data)
    @repositoryResource = $injector.get("repositoryResource")
    @

  RepositoryCredential::repository = ->
    @repositoryResource.fetch @_links.repository.href

  RepositoryCredential

app.factory "repositoryCredentialResource", (Resource, RepositoryCredential) ->
  RepositoryCredentialResource = () ->
    Resource.call(@)
    @

  RepositoryCredentialResource:: = Object.create(Resource::)
  RepositoryCredentialResource::basepath = "/repository-credentials" # not used
  RepositoryCredentialResource::model = RepositoryCredential

  new RepositoryCredentialResource()
