app = angular.module("harrowApp")

app.factory "NotificationRule", ($injector, $http) ->
  NotificationRule = (data) ->
    $.extend(true, @, data)
    @projectResource = $injector.get("projectResource")
    @taskResource = $injector.get("taskResource")
    if @subject
      @subject.taskUuid = @subject.jobUuid
    @

  NotificationRule::project = () ->
    @projectResource.fetch @_links.project.href

  NotificationRule::task = () ->
    @taskResource.fetch @_links.task.href

  NotificationRule

app.factory "notificationRuleResource", (Resource, NotificationRule, $filter) ->
  NotificationRuleResource = () ->
    Resource.call(@)
    @

  NotificationRuleResource:: = Object.create(Resource::)
  NotificationRuleResource::basepath = "/notification-rules"
  NotificationRuleResource::model = NotificationRule

  NotificationRuleResource::_save = NotificationRuleResource::save

  NotificationRuleResource::save = (object) ->
    object.subject.jobUuid = object.subject.taskUuid
    object.subject.notifierType = $filter('pluralize')($filter('underscoreCase')(object.subject.notifierType))
    if object.subject.notifierType == 'task_notifiers'
      object.subject.notifierType = 'job_notifiers'
    @_save(object)

  new NotificationRuleResource()
