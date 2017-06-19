angular.module('harrowApp').config ($stateProvider) ->
  $stateProvider
    .state "operations/show",
      parent: "layout"
      data:
        showViews: [
          'sidebar'
          'header'
        ]
        breadcrumbs: ['organization','project','task','operation']
      url: "/a/operations/{uuid}"

      views:
        "main@layout":
          controller: "operationShowCtrl"
          controllerAs: "operationShow"
          templateUrl: 'views/operations/show.html'
        "header@layout":
          controller: 'breadcrumbsCtrl'
          controllerAs: 'ctrl'
          templateUrl: 'views/_header.html'
        "summary@operations/show":
          controller: "operationShowCtrl"
          controllerAs: "operationShow"
          templateUrl: 'views/operations/summary.html'
        "details@operations/show":
          controller: "operationShowCtrl"
          controllerAs: "operationShow"
          templateUrl: 'views/operations/details.html'
      resolve:
        repositories: (project) ->
          project.repositories()
        organization: (project) ->
          project.organization()
        project: (environment) ->
          environment.project()
        environment: (task) ->
          task.environment()
        taskName: (task, environment, script) ->
          if script and environment
            name = "#{environment.subject.name} - #{script.subject.name}"
            task.subject.name = name
            name
          else
            task.subject.name
        task: (operation) ->
          operation.task()
        script: (task) ->
          task.script()
        operation: (operationResource, $stateParams) ->
          operationResource.find($stateParams.uuid)
