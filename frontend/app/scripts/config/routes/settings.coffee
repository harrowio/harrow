angular.module('harrowApp').config ($stateProvider) ->
  $stateProvider
    .state "settings",
      parent: "layout"
      url: "/a/settings"
      data:
        requiresAuth: true
        showViews: [
          'sidebar'
          'header'
          'app-sidebar'
        ]
        breadcrumbs: ['settings']

      views:
        "main@layout":
          controller: "settingsCtrl"
          controllerAs: "ctrl"
          templateUrl: 'views/settings.html'
        "header@layout":
          controller: 'breadcrumbsCtrl'
          controllerAs: 'ctrl'
          templateUrl: 'views/_header.html'
        "app-sidebar@layout":
          controller: (@menuItems) ->
            @menu = @menuItems.settings
            @
          controllerAs: 'appSidebar'
          templateUrl: 'views/_app-sidebar.html'
        "@settings":
          controller: "settingsCtrl"
          controllerAs: "ctrl"
          templateUrl: 'views/settings/personal.html'
      resolve:
        projects: (authentication) ->
          authentication.currentUser.projects_with_memberships()

    .state "settings.mfa",
      url: "/mfa"
      templateUrl: 'views/settings/mfa.html'

    .state "settings.prompts",
      url: "/prompts"
      templateUrl: 'views/settings/prompts.html'

    .state "settings.project-memberships",
      url: "/project-memberships"
      templateUrl: 'views/settings/project-memberships.html'

    .state "settings.oauth",
      url: "/oauth"
      templateUrl: 'views/settings/oauth.html'
