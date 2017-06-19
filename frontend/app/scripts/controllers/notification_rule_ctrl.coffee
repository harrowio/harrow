Controller = (
  @notificationRuleResource
  @emailNotifierResource
  @$scope
  @authentication
  @$q
  @flash
  @$filter
  @$translate
  @$state
  Stateful
) ->
  @project = @$scope.$resolve.project
  @rules = @$scope.$resolve.rules
  @ruleChangeStateful =
    attrs: {}
  @stateful = new Stateful()
  @stateful.on 'busy', =>
    @ruleChangeStateful.attrs.disabled = true
    @ruleChangeStateful.attrs.ngDisabled = true
  @stateful.on 'completed', =>
    @ruleChangeStateful.attrs.disabled = false
    @ruleChangeStateful.attrs.ngDisabled = false

  @_getCurrentUserEmailNotifier()
  @_generateCheckedRules()
  @operationActions = [
    { name:'Success', actionKey: 'operation.succeeded' }
    { name:'Failed', actionKey: 'operation.failed' }
  ]
  @

Controller::_getCurrentUserEmailNotifier = () ->
  @currentUserEmailNotifier = null
  notifier = @rules.filter (rule) =>
    rule.subject.notifierType == 'email_notifiers' && rule._embedded.notifier[0].subject.recipient.toLowerCase() == @authentication.currentUser.subject.email.toLowerCase()
  return null if notifier.length == 0
  @currentUserEmailNotifier = notifier[0]._embedded.notifier[0]

Controller::_generateCheckedRules = () ->
  @checkedRules = {}
  @rules.forEach (rule) =>
    @checkedRules[rule.subject.notifierUuid] = {} unless @checkedRules[rule.subject.notifierUuid]
    @checkedRules[rule.subject.notifierUuid][rule.subject.matchActivity] = true

Controller::onRuleChange = (notifier, project, task, activity, notifierType) ->
  return unless @stateful.state != 'busy'
  @stateful.transitionTo('busy')
  promises = []
  notifierPromise = @$q.when(notifier)
  if notifier == null && notifierType == 'email_notifiers'
    item =
      subject:
        projectUuid: @project.subject.uuid
        recipient: @authentication.currentUser.subject.email.toLowerCase()

    notifierPromise = @emailNotifierResource.save(item).then (notifier) =>
      @currentUserEmailNotifier = notifier
      notifier

  notifierPromise.then (notifier) =>
    if @hasRule(notifier, activity)
      @rulesForNotifierActivity(notifier, activity).forEach (rule) =>
        promises.push @notificationRuleResource.delete(rule.subject.uuid).then (res) =>
          @rules.splice(@rules.indexOf(rule), 1)
          res
    else
      item =
        subject:
          projectUuid: project.subject.uuid
          taskUuid: task.subject.uuid
          notifierUuid: notifier.subject.uuid
          notifierType: @$filter('underscoreCase')(notifierType)
          matchActivity: activity
      promises.push @notificationRuleResource.save(item).then (res) =>
        @rules.push(res)
        res

    @$q.all(promises).then (res) =>
      @flash.success = @$translate.instant('forms.tasks.notifierRules.flashes.success')
      @$state.reload()
    .catch (reason) =>
      @flash.error = @$translate.instant('forms.tasks.notifierRules.flashes.fail')
      @$q.reject(reason)

Controller::rulesForNotifierActivity = (notifier, activity) ->
  @rules.filter (rule) ->
    rule.subject.matchActivity == activity && rule.subject.notifierUuid == notifier.subject.uuid

Controller::hasRule = (notifier, activity) ->
  @rulesForNotifierActivity(notifier, activity).length > 0

angular.module('harrowApp').controller 'notificationRuleCtrl', Controller
