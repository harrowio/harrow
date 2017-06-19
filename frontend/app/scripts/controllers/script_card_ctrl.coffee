Controller = (
  @project
  @scriptCards
  @tasks
) ->
  @environmentLimit = 4
  @historyLimit = 5
  @scriptCards.forEach (script) =>
    task = @tasks.find (task) ->
      script.subject.lastOperation?.subject.taskUuid == task.subject.uuid
    if task
      script.subject.lastOperationEnvironmentUuid = task.subject.environmentUuid
  @

Controller::emptyDots = (count = 0) ->
  arr = []
  [0...(@historyLimit - count)].forEach (i) ->
    arr.push i
  arr




angular.module('harrowApp').controller 'scriptCardCtrl', Controller
