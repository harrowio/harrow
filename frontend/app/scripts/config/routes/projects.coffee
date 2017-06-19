angular.module('harrowApp').config ($stateProvider) ->
  $stateProvider
    .state 'project',
      parent: 'layout'
      abstract: true # TODO project show?
      url: '/a/projects/{projectUuid}'
      data:
        showViews: [
          'sidebar'
          'header'
        ]
        breadcrumbs: ['organization', 'project']
      views:
        main:
          template: '<div ui-view="main"/>'
        "header@layout":
          controller: 'breadcrumbsCtrl'
          controllerAs: 'ctrl'
          templateUrl: 'views/_header.html'
      resolve:
        project: (projectResource, $stateParams) ->
          projectResource.find($stateParams.projectUuid)
        organization: (project) ->
          project.organization()
        environments: (project) ->
          project.environments()
        rules: (project) ->
          project.notificationRules()

    .state "projects/show",
      parent: "layout"
      url: "/a/projects/{projectUuid}"
      data:
        showViews: [
          'sidebar'
          'header'
        ],
        breadcrumbs: ['organization', 'project']
      views:
        main:
          controller: "projectShowCtrl"
          controllerAs: "ctrl"
          templateUrl: 'views/projects/show.html'
        "header@layout":
          controller: 'breadcrumbsCtrl'
          controllerAs: 'ctrl'
          templateUrl: 'views/_header.html'
        "taskList@projects/show":
          templateUrl: 'views/projects/task_list_show.html'
          controller: "taskListShowCtrl"
          controllerAs: "taskList"
        'scriptList@projects/show':
          templateUrl: 'views/scripts/_cards.html'
          controllerAs: 'ctrl'
          controller: 'scriptCardCtrl'

        'operationList@projects/show':
          templateUrl: 'views/projects/_operation_list.html'
          controller: 'operationListCtrl',
          controllerAs: 'operationList'
      resolve:
        autoCheck: ($stateParams) ->
          $stateParams.autoCheck
        organization: (project) ->
          project.organization()
        project: (projectResource, $stateParams) ->
          projectResource.find($stateParams.projectUuid)
        tasks: (project) ->
          project.tasks()
        scripts: (project) ->
          project.scripts()
        scriptCards: (project) ->
          project.scriptCards()
        repositories: (project) ->
          project.repositories()
        operations: (project, repositories) ->
          project.operations().then (opsCollection) ->
            opsCollection.forEach (op) ->
              Object.keys(op.mostRecentRepositoryCheckouts).forEach (repoUuid) ->
                op.mostRecentRepositoryCheckouts[repoUuid].repository = repositories.find (repo) ->
                  repo.subject.uuid == repoUuid
            opsCollection
        environments: (project) ->
          project.environments()
        members: (project) ->
          project.members()
        rules: (project) ->
          project.notificationRules()

    .state "projects/edit",
      parent: "layout"
      url: "/a/projects/{projectUuid}/edit?{autoCheck?}"
      data:
        requiresAuth: true
        showViews: [
          'sidebar'
          'app-sidebar'
          'header'
        ]
        breadcrumbs: ['organization', 'project']
      views:
        "main@layout":
          controller: "projectCtrl"
          controllerAs: "ctrl"
          templateUrl: 'views/projects/edit.html'
        "header@layout":
          controller: 'breadcrumbsCtrl'
          controllerAs: 'ctrl'
          templateUrl: 'views/_header.html'
        "app-sidebar@layout":
          controller: "projectCtrl"
          controllerAs: "appSidebar"
          templateUrl: 'views/_app-sidebar.html'
      resolve:
        autoCheck: ($stateParams) ->
          $stateParams.autoCheck
        organization: (project) ->
          project.organization()
        project: (projectResource, $stateParams) ->
          projectResource.find($stateParams.projectUuid)

    .state "projects/edit.details",
      url: "/details"
      views:
        "main@layout":
          templateUrl: 'views/projects/edit.html'
          controller: "projectCtrl"
          controllerAs: "ctrl"
    .state "projects/edit.people",
      url: "/people"
      views:
        "main@layout":
          templateUrl: 'views/projects/people.html'
          controller: "projectMemberListCtrl"
          controllerAs: "ctrl"
        "projectMemberList@projects/edit.people":
          templateUrl: 'views/projects/project_member_list.html'
          controller: "projectMemberListCtrl"
          controllerAs: "projectMemberList"
      resolve:
        members: (project) ->
          project.members()

    .state "projects/edit.archive",
      url: "/archive"
      views:
        "main@layout":
          controller: 'projectCtrl'
          controllerAs: 'ctrl'
          templateUrl: 'views/projects/archive.html'
      resolve:
        members: (project) ->
          project.members()
