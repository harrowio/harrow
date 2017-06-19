app = angular.module("harrowApp")

app.factory "Task", ($injector, $http, $q, $log, $filter) ->
  camelCase = $filter('camelCase')
  Task = (data) ->
    $.extend(true, @, data)
    @scheduledExecutionResource = $injector.get("scheduledExecutionResource")
    @triggersScheduleResource = $injector.get('scheduleResource')
    @triggersWebhookResource = $injector.get('webhookResource')
    @triggersGitResource = $injector.get('gitTriggerResource')
    @triggersTaskResource = $injector.get('taskTriggerResource')

    if @_links
      @_links.script = @_links.task
      @_links['triggers-task'] = @_links['triggers-jobs']
      Object.keys(@_links).forEach (linkKey) =>
        singularLinkKey = camelCase(linkKey).replace(/ies$/,'y').replace(/s$/,'')
        resourceName = "#{singularLinkKey}Resource"
        if $injector.has(resourceName) || @.hasOwnProperty(resourceName)
          if !@.hasOwnProperty(resourceName)
            @[resourceName] = $injector.get(resourceName)
          @[camelCase(linkKey)] = ->
            try
              @[resourceName].fetch(@_links[linkKey].href)
            catch e
              $log.warn 'failed to create resource', resourceName, linkKey

    if @subject
      @subject.scriptUuid = @subject.taskUuid
      @subject.lastRunStatus = null
      if @subject.runs && @subject.runs.length > 0
        @subject.lastRunStatus = @subject.runs[0]

    @

  Task::scheduledExecutions = () ->
    @scheduledExecutionResource.fetch @_links['scheduled-executions'].href

  # HACK: We have to get the project via environment or script.
  Task::project = () ->
    @environment().then (environment) ->
      environment.project()


  Task::futureSchedules = ->
    promises = []
    @triggersSchedules().then (schedules) =>
      schedules.forEach (schedule) =>
        if schedule.subject.disabled == null
          promises.push @triggersScheduleResource.fetch(schedule._links.self.href)
      $q.all promises

  Task::triggers = () ->
    promises = {}
    promises.gitTriggers = @triggersGit()
    promises.schedules = @futureSchedules()
    promises.webhooks = @triggersWebhooks()
    promises.taskNotifiers = @triggersTask()
    $q.all promises

  Task::_filterNotifiers = (notifiers, rules) ->
    taskRules = rules.filter (rule) =>
      rule.subject.taskUuid == @subject.uuid
    @_embedded.notifiers = {}
    Object.keys(notifiers).forEach (key) =>
      @_embedded.notifiers[key] = []
    taskRules.forEach (rule) =>
      Object.keys(notifiers).forEach (key) =>
        notifiers[key].filter (notifier) ->
          notifier.subject.uuid == rule.subject.notifierUuid
        .forEach (notifier) =>
          unless notifier.subject.rules
            notifier.subject.rules = []

          hasRule = notifier.subject.rules.some (item) ->
            item.subject.uuid == rule.subject.uuid

          unless hasRule
            notifier.subject.rules.push rule
          cleanRules = []
          notifier.subject.rules.forEach (rule) ->
            hasRule = cleanRules.some (cleanRule) ->
              cleanRule.subject.matchActivity == rule.subject.matchActivity
            unless hasRule
              cleanRules.push rule
          notifier.subject.rules = cleanRules

          hasNotifier = @_embedded.notifiers[key].some (item) ->
            item.subject.uuid == notifier.subject.uuid
          unless hasNotifier
            @_embedded.notifiers[key].push notifier
    @_embedded.notifiers

  Task::notifiers = () ->
    # HACK: Ideally would be nice to get the project owner from `_links.project`
    @script().then (script) =>
      script.project().then (project) =>
        promises = {}
        promises.notifiers = project.notifiers()
        promises.rules = @notificationRules()

        $q.all(promises).then (results) =>
          @_filterNotifiers(results.notifiers, results.rules)


  Task::watch = () ->
    $http.put(@_links.subscriptions.href, { watch: true })

  Task::unwatch = () ->
    $http.put(@_links.subscriptions.href, { watch: false })

  Task::_generateEmbedded = (scripts, environments, projects) ->
    script = scripts.find (script) =>
      script.subject.uuid == @subject.scriptUuid
    env = environments.find (env) =>
      env.subject.uuid == @subject.environmentUuid
    project = projects.find (project) =>
      project.subject.uuid == @subject.projectUuid

    if project
      unless project._embedded.tasks
        project._embedded.tasks = []
      if project._embedded.tasks.indexOf(@) == -1
        project._embedded.tasks.push @

    @_embedded.project = [project]
    @_embedded.scripts = [script]
    @_embedded.environments = [env]
    @subject.taskName = 'Unknown'
    if script && env
      @subject.taskName = "#{env.subject.name} - #{script.subject.name}"
  Task

app.factory "taskResource", (Resource, Task) ->
  TaskResource = () ->
    Resource.call(@)
    @

  TaskResource:: = Object.create(Resource::)
  TaskResource::basepath = "/jobs"
  TaskResource::model = Task

  TaskResource::_save = TaskResource::save

  TaskResource::save = (object) ->
    obj = angular.copy(object)
    object.subject.taskUuid = obj.subject.scriptUuid
    object.subject.jobUuid = obj.subject.taskUuid
    @_save(object)

  new TaskResource()
