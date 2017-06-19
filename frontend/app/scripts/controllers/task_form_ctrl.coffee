app = angular.module("harrowApp")

TaskFormCtrl = (
  @organization
  @project
  @environments
  @scripts
  @job
  @flash
  @$state
  $filter
  $scope
  @$translate
  @jobResource
  @$q
) ->
  @filter = $filter("filter")
  $scope.$watch =>
    @jobName()
  , (jobName) =>
    @job ||= {}
    @job.subject ||= {}
    @job.subject.name = jobName
  @

TaskFormCtrl::save = () ->
  @jobResource.save(@job).then (job) =>
    @flash.success = @$translate.instant("forms.jobForm.flashes.success", job.subject)
    @$state.go("projects/edit.jobs", {uuid: @project.subject.uuid})
    return
  .catch (reason) =>
    @flash.error = @$translate.instant("forms.jobForm.flashes.fail", @job.subject)
    @$q.reject(reason)

TaskFormCtrl::jobName = () ->
  envName = @getName(@environments, @job?.subject?.environmentUuid)
  scriptName = @getName(@scripts, @job?.subject?.scriptUuid)
  if envName && scriptName
    "#{envName} - #{scriptName}"
  else
    ""

TaskFormCtrl::getName = (collection, uuid) ->
  obj = @filter collection,
    subject:
      uuid: uuid
  if obj.length == 0
    undefined
  else
    obj[0].subject.name

app.controller("jobFormCtrl", TaskFormCtrl)
