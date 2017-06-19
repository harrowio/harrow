angular.module('harrowApp').config ($stateProvider) ->
  $stateProvider
    .state "enable_totp",
      parent: "layout_tight"
      url: "/a/enable-totp"
      data:
        requiresAuth: true
      views:
        main:
          templateUrl: 'views/enable-totp.html'
          controller: "enableTotpCtrl"
          controllerAs: "ctrl"
      resolve:
        totpSecret: (authentication) ->
          authentication.currentUser.requestTotpSecret()

    .state "disable_totp",
      parent: "layout_tight"
      url: "/a/disable-totp"
      data:
        requiresAuth: true
      views:
        main:
          templateUrl: 'views/disable-totp.html'
          controller: "disableTotpCtrl"
          controllerAs: "ctrl"
