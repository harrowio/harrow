SidebarCtrl = (
  @$scope
  @$rootScope
  @$window
  @projects
  @organizations
  @authentication
  @$state
) ->
  @organizationList = {}
  @formatOrganizations(@organizations)
  @formatProjects(@projects)

  @$scope.$on 'projectChanged', (event, uuid, response) =>
    @refreshData()

  @$scope.$on 'organizationChanged', (event, uuid, response) =>
    @refreshData()

  @websocketError = false
  @$scope.$root.$on 'wsError', =>
    @websocketError = true

  @

SidebarCtrl::refreshData = () ->
  @authentication.currentUser.organizations().then (items) =>
    @organizations = items
    @authentication.currentUser.projects().then (items) =>
      @projects = items
  .then () =>
    @organizationList = {}
    @formatOrganizations(@organizations)
    @formatProjects(@projects)
    return

SidebarCtrl::formatProjects = (projectsList) ->
  angular.forEach projectsList, (project) =>
    unless @organizationList[project.subject.organizationUuid]
      @organizationList[project.subject.organizationUuid] = project._embedded.organizations[0]
      @organizationList[project.subject.organizationUuid].projects = {}
    @organizationList[project.subject.organizationUuid].projects[project.subject.uuid] = project
    @organizationsLength = Object.keys(@organizationList).length

SidebarCtrl::formatOrganizations = (organizationsList) ->
  angular.forEach organizationsList, (org) =>
    unless @organizations[org.subject.uuid]
      @organizationList[org.subject.uuid] = org
      @organizationList[org.subject.uuid].projects = {}
  @organizationsLength = Object.keys(@organizationList).length

SidebarCtrl::reconnectApp = ->
  # TODO: Make the application gracefully reconnect
  @$window.location.reload()

SidebarCtrl::isActive = (project) ->
  @$state.params.projectUuid == project.subject.uuid

angular.module('harrowApp').controller 'sidebarCtrl', SidebarCtrl
