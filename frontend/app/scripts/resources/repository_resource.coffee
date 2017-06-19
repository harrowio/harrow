app = angular.module("harrowApp")

app.factory "Repository", ($injector, $http, $q, ws) ->
  Repository = (data) ->
    $.extend(true, @, data)
    @operationResource = $injector.get("operationResource")
    @projectResource = $injector.get("projectResource")
    @repositoryCredentialResource = $injector.get("repositoryCredentialResource")
    @

  Repository::check = (updateMetadata) ->
    $http
      method: "POST"
      url: @_links.checks.href
      params:
        updateMetadata:  if updateMetadata then "yes" else "no"
    .then (response) =>
      deferred = $q.defer()

      unless updateMetadata
        if response.data.accessible == true
          deferred.resolve(response.data)
        else
          deferred.reject(response.data)
      else
        operationId = response.data.subject.uuid
        cid = ws.subRow 'operations', operationId, () =>
          @operationResource.find(operationId).then (operation) =>
            return if operation.status() in ['running', 'pending']
            ws.unsubscribe cid
            $http
              method: "GET"
              url: @_links.self.href
            .then (response) ->
              @subject = response.data.subject
              if response.data.subject.accessible == true
                deferred.resolve(response.data)
              else
                deferred.reject(response.data)
      return deferred.promise

  Repository::class = ->
    switch @subject.accessible
      when true      then "repository-accessible"
      when false     then "repository-inaccessible"
      when undefined then "repository-checking"

  Repository::status = ->
    switch @subject.accessible
      when true      then "accessible"
      when false     then "inaccessible"
      when undefined then "being checked"

  Repository::operations = ->
    @operationResource.fetch @_links.operations.href

  Repository::credential = ->
    @repositoryCredentialResource.fetch @_links.credential.href

  Repository::project = ->
    @projectResource.fetch @_links.project.href

  Repository

app.factory "repositoryResource", (Resource, Repository, $q, $http, endpoint) ->
  RepositoryResource = () ->
    Resource.call(@)
    @

  RepositoryResource:: = Object.create(Resource::)
  RepositoryResource::basepath = "/repositories"
  RepositoryResource::model = Repository

  RepositoryResource::preFlight = (urlToCheck) ->
    $http
      method: 'GET'
      url: "#{endpoint}/repo_preflight"
      params:
        url: urlToCheck
    .then (response) ->
      response.data

  new RepositoryResource()
