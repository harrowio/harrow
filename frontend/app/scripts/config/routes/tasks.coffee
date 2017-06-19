angular.module('harrowApp').config ($stateProvider) ->
  $stateProvider
    .state 'task',
      parent: 'project'
      url: '/tasks/{taskUuid}'

      data:
        showViews: [
          'sidebar'
          'header'
        ]
        breadcrumbs: ['organization', 'project', 'task']
      views:
        "main@layout":
          controller: "taskShowCtrl"
          controllerAs: "ctrl"
          templateUrl: 'views/tasks/show.html'
        "header@layout":
          controller: 'breadcrumbsCtrl'
          controllerAs: 'ctrl'
          templateUrl: 'views/_header.html'
        'operationList@task':
          templateUrl: 'views/projects/_operation_list.html'
          controller: 'operationListCtrl',
          controllerAs: 'operationList'
      resolve:
        taskName: (task, environment, script) ->
          if script and environment
            name = "#{environment.subject.name} - #{script.subject.name}"
            task.subject.name = name
            name
          else
            task.subject.name
        task: (taskResource, $stateParams) ->
          taskResource.find($stateParams.taskUuid)
        environment: (task) ->
          task.environment()
        script: (task) ->
          task.script()
        project: (script) ->
          script.project()
        operations: (task) ->
          task.operations()
        environments: (project) ->
          project.environments()
        repositories: (project) ->
          project.repositories()
        tasks: (project) ->
          project.tasks()
        scripts: (project) ->
          project.scripts()
        triggers: (task) ->
          task.triggers()
        notifiers: (task, project) ->
          project.notifiers().then (notifiers) ->
            task.notificationRules().then (rules) ->
              task._filterNotifiers(notifiers, rules)

    .state 'task.edit',
      url: '/edit'
      controller: 'taskEditCtrl'
      controllerAs: 'taskEdit'
      data:
        requiresAuth: true
        showViews: [
          'sidebar'
          'header'
          'app-sidebar'
        ]
        breadcrumbs: ['organization', 'project', 'task']
      views:
        "header@layout":
          controller: 'breadcrumbsCtrl'
          controllerAs: 'ctrl'
          templateUrl: 'views/_header.html'
        "app-sidebar@layout":
          controller: (@menuItems, @taskName) ->
            @menu = @menuItems.taskEdit
            @
          controllerAs: 'appSidebar'
          templateUrl: 'views/_app-sidebar.html'

    .state 'task.edit.notification-rules',
      url: '/notification-rules'
      views:
        "main@layout":
          controller: 'taskEditNotificationRulesCtrl'
          controllerAs: 'ctrl'
          templateUrl: 'views/tasks/edit/notification-rules.html'
      resolve:
        notifiers: (project) ->
          project.notifiers()
        rules: (task) ->
          task.notificationRules()

    .state "task.edit.archive",
      url: "/archive"
      views:
        "main@layout":
          controller: 'taskEditCtrl'
          controllerAs: 'ctrl'
          templateUrl: 'views/tasks/archive.html'
