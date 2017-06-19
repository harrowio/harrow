app = angular.module("harrowApp")

OperationShowCtrl = (
  @$scope
  @$state
  $timeout
  @$translate
  @environment
  @flash
  @task
  @operation
  @operationResource
  @organization
  @project
  @repositories
  @script
  @ws
  initEvents
) ->
  @_commitLimit = 5
  @newestStatusLogEntry = operation.newestStatusLogEntry()
  cid = @ws.subRow "operations", @operation.subject.uuid, () =>
    @operationResource.find(@operation.subject.uuid).then (operation) =>
      @operation = operation
      @newestStatusLogEntry = operation.newestStatusLogEntry()
      $timeout =>
        @$scope.$apply()

  $scope.$on '$destroy', =>
    @ws.unsubscribe cid

  initEvents(@, $scope)
  @

OperationShowCtrl::events = [
  "taskControlsRun"
]

OperationShowCtrl::commitLimit = () ->
  if @operation.subject.gitLogs
    repos = Object.keys(@operation.subject.gitLogs.repositories).length
    Math.ceil(@_commitLimit / repos)
  else
    @_commitLimit

OperationShowCtrl::moreCommits = () ->
  moreCommitsCount = 0;
  if @operation.subject.gitLogs
    Object.keys(@operation.subject.gitLogs.repositories).forEach (key) =>
      moreCommitsCount = moreCommitsCount + @operation.subject.gitLogs.repositories[key].length
    moreCommitsCount = moreCommitsCount - @_commitLimit;
  0 if moreCommitsCount < 0

OperationShowCtrl::isDoubleRepoError = () ->
  repos = {}
  @repositories.forEach (repo) ->
    unless repos[repo.subject.url]
      repos[repo.subject.url] = 1
    repos[repo.subject.url] += 1
  console.log repos

OperationShowCtrl::usesGithub = () ->
  githubRepositories = @repositories.filter (repository) ->
    return repository.subject.url => /github\.com/
  return githubRepositories.length > 0

OperationShowCtrl::controlsDisabled = () ->
  if @operation.status() == "pending" || @operation.status() == "running"
    return true

OperationShowCtrl::repoFromUuid = (repoUuid) ->
  for repo in @repositories
    if repo.subject.uuid == repoUuid
      return repo

OperationShowCtrl::statusView = () ->
  if @operation.status() == "timedout" then "timed out" else @operation.status()

app.controller("operationShowCtrl", OperationShowCtrl)
