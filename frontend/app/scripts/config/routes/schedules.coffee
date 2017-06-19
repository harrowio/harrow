angular.module('harrowApp').config ($stateProvider) ->
  $stateProvider
    .state "schedules/show",
      parent: "layout"
      data:
        showViews: [
          'sidebar'
          'header'
        ]
        breadcrumbs: ['organization', 'project', 'task', 'schedule']
      url: "/a/schedules/{uuid}"

      views:
        "main@layout":
          controller: "scheduleShowCtrl"
          controllerAs: "scheduleShow"
          templateUrl: 'views/schedules/show.html'
        "header@layout":
          controller: 'breadcrumbsCtrl'
          controllerAs: 'ctrl'
          templateUrl: 'views/_header.html'
        'operationList@schedules/show':
          templateUrl: 'views/projects/_operation_list.html'
          controller: 'operationListCtrl',
          controllerAs: 'operationList'
      resolve:
        repositories: (project) ->
          project.repositories()
        organization: (project) ->
          project.organization()
        project: (environment) ->
          environment.project()
        environment: (task) ->
          task.environment()
        environments: (project) ->
          project.environments()
        scripts: (project) ->
          project.scripts()
        taskName: (task, environment, script) ->
          if script and environment
            name = "#{environment.subject.name} - #{script.subject.name}"
            task.subject.name = name
            name
          else
            task.subject.name
        task: (schedule) ->
          schedule.task()
        script: (task) ->
          task.script()
        schedule: (scheduleResource, $stateParams) ->
          scheduleResource.find($stateParams.uuid)
        scheduledExecutions: (schedule) -> schedule.scheduledExecutions()
        operations: (schedule) ->
          schedule.operations().then (operations) ->
            operations[0..2]
