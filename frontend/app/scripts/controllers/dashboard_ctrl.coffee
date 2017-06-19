app = angular.module("harrowApp")

Controller = (
  @organizations
  @projects
  @tasks
  @$filter
  @cardsByOrganization
) ->
  @_generateEmbeddedContent()
  @

Controller::_generateEmbeddedContent = ->
  @projects?.forEach (project) =>
    @_addMissingOrg(project)
    project._embedded = {} unless project._embedded
    project._embedded.tasks = @tasks.filter (task) ->
      task.subject.projectUuid == project.subject.uuid

Controller::_addMissingOrg = (project) ->
  projectOrg = project._embedded.organizations[0]
  hasOrg = @organizations.some (org) ->
    org.subject.uuid == projectOrg.subject.uuid
  unless hasOrg
    @organizations.push project._embedded.organizations[0]

Controller::projectsFor = (org) ->
  @projects.filter (project) ->
    project.subject.organizationUuid == org.subject.uuid

Controller::mostRecentOperationFor = (project) ->
  r = @$filter('orderBy')(project._embedded.operations, 'subject.createdAt', true)
  r[0]

Controller::taskNameFor = (operation) ->
  task = operation._embedded.tasks[0]
  envName = task._embedded.environments[0].subject.name
  scriptName = task._embedded.scripts[0].subject.name
  "#{envName} - #{scriptName}"

app.controller("dashboardCtrl", Controller)
