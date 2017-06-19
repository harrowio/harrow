angular.module('harrowApp').config ($stateProvider) ->
  $stateProvider
    .state "github/callback/authorize",
      parent: "layout_tight"
      url: "/a/github/callback/authorize"
      views:
        main:
          controller: "githubCallbackCtrl"
          controllerAs: "ctrl"
          templateUrl: 'views/github_callback.html'
      resolve:
        action: -> "authorize"

    .state "github/callback/signin",
      parent: "layout_tight"
      url: "/a/github/callback/signin"
      views:
        main:
          controller: "githubCallbackCtrl"
          controllerAs: "ctrl"
          templateUrl: 'views/github_callback.html'
      resolve:
        action: -> "signin"
