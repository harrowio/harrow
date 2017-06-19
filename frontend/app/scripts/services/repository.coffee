Service = (
  @organizationResource
  @projectResource
  @repositoryResource
  @$filter
  @$q
) ->
  @

Service::import = (url) ->
  deferred = @$q.defer()
  url = @$filter('url')(url, null)
  parts = url.pathname.replace(/^\//,'').split('/')
  isSsh = @$filter('isSsh')(url)
  if parts.length == 2
    organizationName = @$filter('titlecase')(parts[0])
    projectName = @$filter('titlecase')(parts[1].replace(/\.git$/,''))
  else
    return deferred.reject()
  @organizationResource.save(
    subject:
      public: false
      planUuid: "b99a21cc-b108-466e-aa4d-bde10ebbe1f3"
      name: organizationName
  )
  .then (organization) =>
    @organization = organization
    @projectResource.save(
      subject:
        organizationUuid: organization.subject.uuid
        name: projectName
    )
  .then (project) =>
    repository =
      subject:
        projectUuid: project.subject.uuid
    @project = project
    if isSsh
      repository.subject.url = @url.toString('ssh')
    else
      repository.subject.url = @url.toString('https')

    @repositoryResource.save(repository)
  .then (repository) =>
    @repository = repository
    repository.check().catch () =>
      return false
  .then (check) =>
    return resolve(
      project: @project
      organization: @organization
      repository: @repository
      repositoryIsSsh: isSsh
      repositoryAccessible: check
    )

angular.module('harrowApp').service 'repositoryImport', Service
