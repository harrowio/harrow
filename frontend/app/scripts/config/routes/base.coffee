angular.module('harrowApp').config ($stateProvider) ->
  $stateProvider
    .state "layout",
      abstract: true
      params:
        returnState: ''
        returnParams: {}
        returnTo: false
        returnToHere: false
      views:
        "@":
          templateUrl: 'views/layout_app.html'
        "sidebar@layout":
          templateUrl: 'views/sidebar.html'
          controller: "sidebarCtrl"
          controllerAs: "sidebar"
      data:
        showViews: [
          'sidebar'
        ]
      resolve:
        currentUser: (authentication) ->
          authentication.currentUser
        organizations: (authentication) ->
          authentication.currentUser?.organizations()
        projects: (authentication) ->
          authentication.currentUser?.projects()
        tasks: (authentication) ->
          authentication.currentUser?.tasks()
    .state 'layout_tight',
      abstract: true
      params:
        returnState: ''
        returnParams: {}
        returnTo: false
        returnToHere: false
      views:
        "@":
          templateUrl: 'views/layout_tight.html'
    .state "errors/403",
      parent: "layout_tight",
      views:
        main:
          templateUrl: 'views/errors/403.html'
          controller: "errorCtrl"
          controllerAs: "error"

    .state "errors/404",
      parent: "layout_tight",
      views:
        main:
          templateUrl: 'views/errors/404.html'
          controller: "errorCtrl"
          controllerAs: "error"

    .state "errors/418",
      parent: "layout_tight",
      views:
        main:
          templateUrl: 'views/errors/418.html'
          controller: "errorCtrl"
          controllerAs: "error"

    .state "errors/502",
      parent: "layout_tight",
      views:
        main:
          templateUrl: 'views/errors/502.html'
          controller: "errorCtrl"
          controllerAs: "error"

    .state "errors/500",
      parent: "layout_tight",
      views:
        main:
          templateUrl: 'views/errors/500.html'
          controller: "errorCtrl"
          controllerAs: "error"

    .state "errors/blocked",
      parent: "layout_tight",
      url: "/a/errors/blocked",
      views:
        main:
          templateUrl: 'views/errors/blocked.html'
          controller: "errorCtrl"
          controllerAs: "error"

    .state "errors/session_invalidated",
      parent: "layout_tight"
      views:
        main:
          templateUrl: 'views/errors/session_invalidated.html'
          controller: 'errorCtrl'
          controllerAs: 'error'

    .state "errors/verification_email_sent",
      parent: "layout_tight",
      url: "/a/errors/verification_email_sent",
      views:
        main:
          templateUrl: 'views/errors/verification_email_sent.html'
          controller: "errorCtrl"
          controllerAs: "error"

    .state "errors/github_existing_unlinked_user",
      parent: "layout_tight",
      url: "/a/errors/github_existing_unlinked_user",
      views:
        main:
          templateUrl: 'views/errors/github_existing_unlinked_user.html'

    .state "activity",
      parent: "layout"
      url: "/a/activity"
      views:
        main:
          templateUrl: 'views/activity.html'
          controller: "activityCtrl"
          controllerAs: "activityList"
        "@settings":
          templateUrl: 'views/activity.html'
      resolve:
        activities: (authentication) ->
          authentication.currentUser.activities()
