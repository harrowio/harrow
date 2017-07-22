app = angular.module("harrowApp")

app.factory "RepositoryCheckout", () ->
  RepositoryCheckout = (data) ->
    @Ref = data.Ref
    @Hash = data.Hash
    @

  RepositoryCheckout::shortHash = () ->
    @Hash.substring(0,7)

  RepositoryCheckout::refName = () ->
    parts = @Ref.split('/')
    return parts[parts.length - 1]

  RepositoryCheckout

app.factory "Operation", ($filter, $injector, RepositoryCheckout) ->

  Operation = (data) ->
    $.extend(true, @, data)
    @taskResource = $injector.get("taskResource")
    @deliveryResource = $injector.get("deliveryResource")
    @formatMostRecentRepositoryCheckouts()
    @_calculateDurations()
    if @subject
      @subject.taskUuid = @subject.jobUuid
    @

  Operation::reason = ->
    @subject?.parameters?.reason

  Operation::task = ->
    @taskResource.fetch @_links.job.href

  Operation::delivery = ->
    if @subject.parameters?.triggeredByDelivery
      @deliveryResource.fetch @_links.task.href

  Operation::newestStatusLogEntry = ->
    n = @subject?.statusLogs?.entries.length
    return @subject?.statusLogs?.entries[n-1] if n > 0

  Operation::_calculateDurations = ->
    return unless @subject.statusLogs and @subject.statusLogs.entries
    @subject.statusLogs.entries.forEach (entry, index) =>
      if index == 0
        entry.duration = moment(entry.occurredOn).diff(@subject.createdAt)
      else
        entry.duration = moment(entry.occurredOn).diff(@subject.statusLogs.entries[index - 1].occurredOn)
      entry.formattedDuration = $filter('momentDurationFormat')(entry.duration)

  Operation::formatMostRecentRepositoryCheckouts = ->
    result = {}
    for repositoryUuid, checkouts of @subject.repositoryCheckouts?.refs
      result[repositoryUuid] = new RepositoryCheckout(checkouts[checkouts.length - 1])
    @mostRecentRepositoryCheckouts = result

  Operation::elapsed = ->
    if @subject.endedTime()
      @subject.endedTime() - @subject?.startedAt

  Operation::cancelable = ->
    status = @status();
    return status == 'running' || status == 'pending'

  Operation::status = ->
    if @subject.canceledAt
      "canceled"
    else if @subject.fatalError
      "fatalerr"
    else if @subject.timedOutAt
      "timedout"
    else if 256 == @subject.exitStatus
      if @newestStatusLogEntry()
        "running"
      else
        "pending"
    else if 0 == @subject.exitStatus
      "success"
    else
      "failed"

  Operation::isError = ->
    @subject.fatalError || @subject.timedOutAt || @subject.failedAt

  Operation::isSuccess = ->
    @subject.exitStatus == 0

  Operation::isBusy = ->
    @subject.exitStatus == 256

  # returns the time the operation has ended or undefined
  Operation::endedTime = () ->
    if @subject.finishedAt
      return @subject.finishedAt
    if @subject.failedAt
      return @subject.failedAt
    if @subject.timedOutAt
      return @subject.timedOutAt
    return undefined

  Operation::taskName = () ->
    @_embedded?.task?[0]?.subject?.name

  Operation

app.factory "operationResource", (Resource, Operation) ->
  OperationResource = () ->
    Resource.call(@)
    @

  OperationResource:: = Object.create(Resource::)
  OperationResource::basepath = "/operations"
  OperationResource::model = Operation

  new OperationResource()
