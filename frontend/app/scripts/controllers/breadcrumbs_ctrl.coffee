Controller = (
  @$scope
  @$state
  @$filter
  @$log
  @ws
  @operationResource
) ->
  @menu = []
  @editItem = null
  @_generateMenuItems()
  @

Controller::_generateMenuItems = () ->
  @menu = []
  @menu.push title: 'Harrow', name: 'Home', stateName: 'dashboard'
  if @$state.current.data.breadcrumbs
    @$state.current.data.breadcrumbs.forEach (crumb) =>
      if crumb == 'organization'
        @_generateCrumb('organization', 'organization', 'uuid')
      else if crumb == 'project'
        @_generateCrumb('project', 'projects/show', 'projectUuid', 'projects/edit')
      else if crumb == 'billing'
        @menu.push title: 'Billing', name: 'Billing', stateName: 'billing'
        @editItem = stateName: 'billing'
      else if crumb == 'settings'
        @menu.push title: 'Settings', name: @$scope.$resolve.currentUser.subject.name, stateName: 'settings'
        @editItem = null
      else if crumb == 'schedule'
        @_generateCrumb('schedule', 'schedules/show', 'uuid', 'triggers.schedule.edit')
        @editItem.name = 'Edit schedule'
      else if crumb == 'script'
        if @$state.current.name == "createScript"
          @menu.push title: '&nbsp;', name: 'New Script'
          @editItem = null
        else
          @_generateCrumb(crumb, crumb, "#{crumb}Uuid")
          @editItem.name = "Edit Script"
      else if crumb == 'operation'
        @$scope.$resolve.operation.subject.name = @$filter('amDateFormat')(@$scope.$resolve.operation.subject.createdAt, 'LLLL')
        @_generateCrumb(crumb, 'operations/show', "uuid")
      else
        @_generateCrumb(crumb, crumb, "#{crumb}Uuid")

Controller::_generateCrumb = (title, state, uuidKey, editState) ->
  obj = @$scope.$resolve[title]
  project = @$scope.$resolve.project
  if title == 'operation'
    cid = @ws.subRow "operations", @$scope.$resolve.operation.subject.uuid, () =>
      @operationResource.fetch(@$scope.$resolve.operation._links.self.href).then (operation) =>
        obj.subject = operation.subject
    @$scope.$on '$destroy', => @ws.unsubscribe cid

  if title == 'schedule'
    if obj.subject.timespec
      obj.subject.name = obj.subject.timespec
    else
      obj.subject.name = obj.subject.cronspec

  if obj && !angular.isArray(obj) && obj.subject.name
    stateParams = {}
    stateParams.projectUuid = project.subject.uuid if project
    stateParams[uuidKey] = obj.subject.uuid
    item =
      crumb: title
      title: @$filter('titlecase')(title)
      name: obj.subject.name
      resource: obj
      stateName: state
      stateParams: stateParams

    @editItem =
      stateName: editState || "#{state}.edit"
      stateParams: item.stateParams
      name: "#{item.title} Settings"
      resource: item.resource

    unless @$state.get(@editItem.stateName)
      @editItem = null

    @menu.push item
  else
    @$log.debug "Breadcrumb Unhandled Crumb", title, @$scope.$resolve

Controller::hasEditing = () ->
  @$state.get(@editItem.stateName)

Controller::isEditing = () ->
  @$state.includes(@editItem.stateName)
angular.module('harrowApp').controller 'breadcrumbsCtrl', Controller
