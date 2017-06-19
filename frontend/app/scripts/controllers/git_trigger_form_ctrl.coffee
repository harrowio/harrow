app = angular.module("harrowApp")

GitTriggerFormCtrl = (
  @organization
  @project
  @repositories
  @gitTrigger
  @gitTriggerResource
  @tasks
  @flash
  @$state
  @$translate
  @$q
  @$scope
) ->
  @triggerReasonKinds = [
    {label: "is changed"; value: "change"},
    {label: "is added"; value: "add"},
    {label: "is removed"; value: "remove"},
  ]
  @

GitTriggerFormCtrl::save = () ->
  @gitTriggerResource.save(@gitTrigger).then (gitTrigger) =>
    @flash.success = @$translate.instant("forms.gitTriggerForm.flashes.success", gitTrigger.subject)
    @$state.go("projects/edit.triggers", {projectUuid: gitTrigger.subject.projectUuid})
    return
  .catch (reason) =>
    @flash.error = @$translate.instant("forms.gitTriggerForm.flashes.fail", @gitTrigger.subject)
    @$q.reject(reason)

app.controller("gitTriggerFormCtrl", GitTriggerFormCtrl)
