app = angular.module("harrowApp")

TaskListEditCtrl = (
  @tasks
  @scheduleResource
  @project
  @taskResource
  @$translate
  @$state
  @$scope
  @flash
) ->
  @$scope.$on "reloadTasks", =>
    @project.tasks().then (@tasks) =>
  @


TaskListEditCtrl::deleteTask = (task) ->
  return unless @confirm()
  @taskResource.delete(task.subject.uuid).then =>
    @flash.success = @$translate.instant("tasks.flashes.delete.success")
    @tasks = @tasks.filter (existingTask) -> existingTask.subject.uuid != task.subject.uuid
    return
  .catch () =>
    @flash.error = @$translate.instant("tasks.flashes.delete.fail")
    return
  .finally =>
    @$scope.$emit("tasksChanged")
    return

TaskListEditCtrl::confirm = () ->
  confirm(@$translate.instant("prompts.really?"))

TaskListEditCtrl::createSchedule = (task) ->
  @scheduleResource.save
    subject:
      taskUuid: task.subject.uuid
      description: "Ad-hoc"
      timespec: "now"

app.controller("taskListEditCtrl", TaskListEditCtrl)
