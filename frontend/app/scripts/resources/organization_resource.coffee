app = angular.module("harrowApp")
camelCase = angular.element.camelCase

app.factory "Organization", ($injector, $q) ->
  Organization = (data) ->
    $.extend(true, @, data)
    @projectResource = $injector.get("projectResource")
    @organizationMemberResource = $injector.get("organizationMemberResource")
    @$http = $injector.get("$http")
    if @_links
      Object.keys(@_links).forEach (linkKey) =>
        singularLinkKey = camelCase(linkKey).replace(/ies$/,'y').replace(/s$/,'')
        resourceName = "#{singularLinkKey}Resource"
        if $injector.has(resourceName)
          @[resourceName] = $injector.get(resourceName)
          @[camelCase(linkKey)] = ->
            @[resourceName].fetch(@_links[linkKey].href)
    if @subject
      @subject.planUuid = @subject.billingPlanUuid
    @

  # HACK: tasks is only gathered though each project, this can become rather wasteful with many projects and tasks
  Organization::tasks = () ->
    @projects().then (projects) ->
      promises = []
      projects.forEach (project) ->
        promises.push project.tasks()

      $q.all(promises).then (results) ->
        tasksArr = []
        results.forEach (tasks) ->
          tasks.forEach (task) ->
            tasksArr.push task
        tasksArr

  # HACK: operations is only gathered though each project, could lead a load of XHR requests.
  Organization::operations = () ->
    @projects().then (projects) ->
      promises = []
      projects.forEach (project) ->
        promises.push project.operations()

      $q.all(promises).then (results) ->
        arr = []
        results.forEach (items) ->
          items.forEach (item) ->
            arr.push item
        arr

  # HACK: repositories is only gathered though each project, could lead a load of XHR requests.
  Organization::repositories = () ->
    @projects().then (projects) ->
      promises = []
      projects.forEach (project) ->
        promises.push project.repositories()

      $q.all(promises).then (results) ->
        arr = []
        results.forEach (items) ->
          items.forEach (item) ->
            arr.push item
        arr

  # HACK: environments is only available though tasks.
  Organization::environments = () ->
    @tasks().then (tasks) ->
      promises = []
      tasks.forEach (task) ->
        promises.push task.environment()
      $q.all promises

  # HACK: scripts is only available though tasks.
  Organization::scripts = () ->
    @tasks().then (tasks) ->
      promises = []
      tasks.forEach (task) ->
        promises.push task.script()
      $q.all promises

  Organization::projects = () ->
    @projectResource.fetch(@_links.projects.href)

  Organization::projectCards = () ->
    @$http.get(@_links['project-cards'].href).then (response) ->
      response.data?.collection

  Organization::members = () ->
    @organizationMemberResource.fetch(@_links.members.href)

  Organization::addCreditCard = (nonce, organizationUuid) ->
    @$http.post @_links['add-credit-card'].href, {nonce, organizationUuid}
    .catch (response) ->
      if response.status == 422
        $q.reject(response.data)
      else
        $q.reject(response.status)

  Organization

app.factory "organizationResource", (Resource, Organization) ->
  OrganizationResource = () ->
    Resource.call(@)
    @

  OrganizationResource:: = Object.create(Resource::)
  OrganizationResource::basepath = "/organizations"
  OrganizationResource::model = Organization

  new OrganizationResource()
