xdescribe "Controller: taskShowCtrl", ->

  afterEach =>
    @scope.$broadcast("$destroy")

  beforeEach angular.mock.inject ($controller, $rootScope, $q) =>
    @scope = $rootScope.$new()
    @task =
      subject:
        uuid: "782d2eb2-2f80-40a9-9e2e-797276b7f14b"
    @taskResource =
      delete: (uuid) ->
    spyOn(@taskResource, "delete").and.returnValue($q.when())

    @state = jasmine.createSpyObj("$state", ["go"])

    @project =
      subject:
        uuid: "8b7b5278-8eae-496d-b4df-e22991765281"

    @controller = $controller "taskShowCtrl",
      organization: null
      project: @project
      environment: null
      secrets: null
      task: null
      task: @task
      tasks: null
      taskNotifiers: null
      taskResource: @taskResource
      operations: null
      subscriptions: jasmine.createSpyObj("subscriptions", ["isWatching"])
      scheduledExecutions: null
      $scope: @scope
      $state: @state
      notificationRules: null
      triggers: null
      notifiers: null

    @scope.taskShow = @controller

  it "should delete a task and redirect to the project", =>
    @controller.confirm = ->
      true
    @controller.delete()
    @scope.$apply()
    expect(@taskResource.delete).toHaveBeenCalledWith(@task.subject.uuid)
    expect(@state.go).toHaveBeenCalledWith("projects/edit", {uuid: @project.subject.uuid})
