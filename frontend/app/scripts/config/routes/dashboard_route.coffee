angular.module('harrowApp').config ($stateProvider) ->
  $stateProvider
    .state "dashboard",
      parent: "layout"
      url: "/a/dashboard"
      data:
        requiresAuth: true
        showViews: [
          'sidebar'
          'header'
        ]
      views:
        "main@layout":
          templateUrl: 'views/dashboard/index.html'
          controller: "dashboardCtrl"
          controllerAs: "ctrl"
        "header@layout":
          controller: 'breadcrumbsCtrl'
          controllerAs: 'ctrl'
          templateUrl: 'views/_header.html'
        "taskList@dashboard":
          templateUrl: 'views/projects/task_list_show.html'
          controller: "taskListShowCtrl"
          controllerAs: "taskList"
      resolve:
        project: () ->
          angular.noop()
        organizations: (authentication) ->
          authentication.currentUser?.organizations()
        cardsByOrganization: ($q, organizations) ->
          byOrganizationId = {}
          promises = []
          organizations?.forEach (organization) ->
            promises.push organization.projectCards().then (cards) ->
              byOrganizationId[organization.subject.uuid] = cards

          $q.all(promises).then () -> byOrganizationId
