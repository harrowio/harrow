Controller = (
  @task
  @project
  @notifiers
  @rules
  @notificationRuleResource
  @tasks
  @scripts
  @environments
  @$q
  @flash
  @$translate
  @$stateParams
  @$filter
) ->
  @actions = [
    { name:'Success', actionKey: 'operation.succeeded' }
    { name:'Failed', actionKey: 'operation.failed' }
  ]

  @generateCheckedRules()
  @generateStatefulOptions()
  @_generateNamesForTasks()
  @

Controller::_generateNamesForTasks = () ->
  if @notifiers.taskNotifiers && @tasks && @scripts && @environments
    @notifiers.taskNotifiers.forEach (notifier) =>
      task = @tasks.find (task) ->
        notifier.subject.taskUuid == task.subject.uuid
      if task
        script = @scripts.find (script) ->
          script.subject.uuid == task.subject.scriptUuid
        env = @environments.find (env) ->
          env.subject.uuid == task.subject.environmentUuid
        if env && script
          notifier.subject.name = "#{env.subject.name} - #{script.subject.name}"
      else
        notifier.subject.name = "Unknown Task"



Controller::generateStatefulOptions = () ->
  @statefulOptions = {}
  @ruleChangeBusy = {}
  Object.keys(@notifiers).forEach (notifierType) =>
    @notifiers[notifierType].forEach (notifier) =>
      @ruleChangeBusy[notifierType] = {} unless @ruleChangeBusy[notifierType]
      @statefulOptions[notifierType] = {} unless @statefulOptions[notifierType]
      @ruleChangeBusy[notifierType][notifier.subject.uuid] = false
      @statefulOptions[notifierType][notifier.subject.uuid] =
        pending:
          watch: () =>
            @ruleChangeBusy[notifierType][notifier.subject.uuid] == true
          attrs: {
            ngShow: true
          }
        completed:
          watch: () =>
            @ruleChangeBusy[notifierType][notifier.subject.uuid] == false
          attrs: {
            ngShow: false
          }
Controller::generateCheckedRules = () ->
  @checkedRules = {}
  @rules.forEach (rule) =>
    @checkedRules[rule.subject.notifierUuid] = {} unless @checkedRules[rule.subject.notifierUuid]
    @checkedRules[rule.subject.notifierUuid][rule.subject.matchActivity] = true

Controller::onRuleChange = (notifier, activity, notifierType) ->
  promises = []
  @ruleChangeBusy[notifierType][notifier.subject.uuid] = true
  if @hasRule(notifier, activity)
    @rulesForNotifierActivity(notifier, activity).forEach (rule) =>
      promises.push @notificationRuleResource.delete(rule.subject.uuid).then (res) =>
        @rules.splice(@rules.indexOf(rule), 1)
        res
  else
    item =
      subject:
        projectUuid: @project.subject.uuid
        taskUuid: @task.subject.uuid
        notifierUuid: notifier.subject.uuid
        notifierType: @$filter('underscoreCase')(notifierType)
        matchActivity: activity
    promises.push @notificationRuleResource.save(item).then (res) =>
      @rules.push(res)
      res

  @$q.all(promises).then (res) =>
    @ruleChangeBusy[notifierType][notifier.subject.uuid] = false
    @flash.success = @$translate.instant('forms.tasks.notifierRules.flashes.success')
    res
  .catch (reason) =>
    @flash.error = @$translate.instant('forms.tasks.notifierRules.flashes.fail')
    @$q.reject(reason)

Controller::rulesForNotifierActivity = (notifier, activity) ->
  @rules.filter (rule) ->
    rule.subject.matchActivity == activity && rule.subject.notifierUuid == notifier.subject.uuid

Controller::hasRule = (notifier, activity) ->
  @rulesForNotifierActivity(notifier, activity).length > 0

Controller::editSrefFor = (triggerType) ->
  type = @$filter('singularize')(triggerType)
  "task.edit.notifiers.#{type}.edit"

Controller::createSrefFor = (triggerType) ->
  type = @$filter('singularize')(triggerType)
  "task.edit.notifiers.#{type}"

angular.module('harrowApp').controller 'taskEditNotificationRulesCtrl', Controller
