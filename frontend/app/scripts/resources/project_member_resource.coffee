app = angular.module("harrowApp")

app.factory "ProjectMember", ($injector, $http) ->
  ProjectMember = (data) ->
    $.extend(true, @, data)
    @

  ProjectMember::remove = (projectUuid) ->
    $http.delete(@_links.self.href, {params: {projectUuid: projectUuid}})

  ProjectMember

app.factory "projectMemberResource", (Resource, ProjectMember) ->
  ProjectMemberResource = () ->
    Resource.call(@)
    @

  ProjectMemberResource:: = Object.create(Resource::)
  ProjectMemberResource::basepath = "/project-members"
  ProjectMemberResource::model = ProjectMember

  new ProjectMemberResource()
