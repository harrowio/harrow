angular.module('harrowApp').config ($stateProvider) ->
  $stateProvider
    .state "forgot_password",
      parent: "layout_tight"
      url: "/a/forgot-password"
      views:
        main:
          templateUrl: 'views/forgot-password.html'
          controller: "forgotPasswordCtrl"
          controllerAs: "forgotPassword"
      data:
        requiresAuth: false

    .state "reset_password",
      parent: "layout_tight"
      url: "/a/reset-password?email&mac"
      views:
        main:
          templateUrl: 'views/reset-password.html'
          controller: "resetPasswordCtrl"
          controllerAs: "resetPassword"
      data:
        requiresAuth: false
      resolve:
        email: ($stateParams) ->
          $stateParams.email
        mac: ($stateParams) ->
          $stateParams.mac

    .state "login",
      parent: "layout_tight"
      url: "/a/login?invitation&origin"
      views:
        main:
          templateUrl: 'views/login.html'
          controller: "loginCtrl"
          controllerAs: "ctrl"
      data:
        requiresAuth: false
        container: 'small'


    .state "session_confirmation",
      parent: "layout_tight"
      url: "/a/session-confirmation?invitation&origin"
      views:
        main:
          templateUrl: 'views/session-confirmation.html'
          controller: "sessionConfirmationCtrl"
          controllerAs: "ctrl"
      data:
        requiresAuth: false

    .state "verify_email",
      parent: "layout_tight"
      url: "/a/verify-email?user&token"
      views:
        main:
          templateUrl: 'views/verify-email.html'
          controller: "verifyEmailCtrl"
          controllerAs: "verifyEmail"
      resolve:
        userUuid: ($stateParams) ->
          $stateParams.user
        token: ($stateParams) ->
          $stateParams.token
      data:
        requiresAuth: false

    .state "signup",
      parent: "layout_tight"
      url: "/a/signup?invitation&origin&utm_campaign&utm_term&utm_content&utm_source&utm_medium&cta&gclid"
      views:
        main:
          templateUrl: 'views/signup.html'
          controller: "signupCtrl"
          controllerAs: "ctrl"
      data:
        requiresAuth: false
        container: 'small'
