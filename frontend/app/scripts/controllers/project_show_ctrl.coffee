app = angular.module("harrowApp")

ProjectShowCtrl = (
  @$q
  @$state
  @$translate
  @flash
  @tasks
  @members
  @operations
  @organization
  @scripts
  @scriptCards
  @project
  @projectResource
  @repositories
  @modal
  @can
) ->
  if @can.can('update-repositories', @project) && @hasInaccessibleRepository()
    @modal.show(
      mode: 'warning'
      modal: "hasInaccessibleRepository"
      title: "There is a problem accessing your repositories"
      content: "One or more of your repositories has become inaccessible to harrow and may cause your tasks to fail"
      href: @$state.href('repositories', {projectUuid: @project.subject.uuid})
      name: "Go to repositories"
    )
  @

ProjectShowCtrl::archive = () ->
  return unless @confirm()
  @projectResource.delete(@project.subject.uuid).then (project) =>
    @flash.success = @$translate.instant("forms.projectForm.flashes.success", project.subject)
    @$state.go("organizations/show", {uuid: @project.subject.organizationUuid})
    return
  .catch (reason) =>
    @flash.error = @$translate.instant("forms.projectForm.flashes.fail", @project.subject)
    @$q.reject(reason)


ProjectShowCtrl::getTask = (uuid) ->
  j = null
  angular.forEach @tasks, (task) ->
    if uuid == task.subject.uuid
      j = task
  j

ProjectShowCtrl::confirm = () ->
  confirm(@$translate.instant("prompts.really?"))

ProjectShowCtrl::hasInaccessibleRepository = () ->
  @repositories.some (repo) ->
    repo.subject.accessible == false


app.controller("projectShowCtrl", ProjectShowCtrl)
