app = angular.module("harrowApp")

app.factory "Project", ($injector, $http, $q, $filter) ->
  camelCase = $filter('camelCase')
  Project = (data) ->
    $.extend(true, @, data)

    if @_links
      links = angular.copy(@_links)
      @_links['task-notifiers'] = links['job-notifiers']
      @_links.tasks = links.jobs
      @_links.scripts = links.tasks
      @_links.scriptCards = links.scripts
      Object.keys(@_links).forEach (linkKey) =>
        singularLinkKey = camelCase(linkKey).replace(/ies$/,'y').replace(/s$/,'')
        resourceName = "#{singularLinkKey}Resource"
        if $injector.has(resourceName)
          @[resourceName] = $injector.get(resourceName)
          @[camelCase(linkKey)] = ->
            @[resourceName].fetch(@_links[linkKey].href)
    @

  Project::members = ->
    return @projectMembers()

  Project::notifiers = ->
    promises = {}
    Object.keys(@_links).forEach (linkKey) =>
      if /-notifiers$/.test(linkKey)
        objKey = camelCase(linkKey)
        if @.hasOwnProperty(objKey)
          promises[objKey] = @[objKey]()
    $q.all promises

  Project::triggers = ->
    promises = {}
    promises.gitTriggers = @gitTriggers()
    promises.schedules = @futureSchedules()
    promises.webhooks = @webhooks()
    promises.taskNotifiers = @taskNotifiers()
    $q.all promises

  Project::futureSchedules = ->
    promises = []
    @schedules().then (schedules) =>
      schedules.forEach (schedule) =>
        if schedule.subject.disabled == null
          promises.push @scheduleResource.fetch(schedule._links.self.href)
      $q.all promises

  Project::leave = ->
    $http.delete(@_links.leave.href)

  Project

app.factory "projectResource", (Resource, Project) ->
  ProjectResource = () ->
    Resource.call(@)
    @

  ProjectResource:: = Object.create(Resource::)
  ProjectResource::basepath = "/projects"
  ProjectResource::model = Project

  new ProjectResource()
