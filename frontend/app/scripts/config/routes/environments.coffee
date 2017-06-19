angular.module('harrowApp').config ($stateProvider) ->
  $stateProvider
    .state "environments",
      parent: "projects/edit"
      url: "/environments"
      views:
        "main@layout":
          controller: "environmentListCtrl"
          controllerAs: "environmentList"
          templateUrl: 'views/environments/index.html'
      resolve:
        environments: (project) ->
          project.environments()
        secrets: ($q, environments) ->
          promises = []
          environments.forEach (environment) ->
            promises.push environment.secrets()
          $q.all promises
    .state "createEnvironment",
      parent: "projects/edit"
      url: "/environments/new"
      views:
        "main@layout":
          templateUrl: 'views/environments/create.html'
          controller: "environmentFormCtrl"
          controllerAs: "environmentForm"
      resolve:
        environment: (project) ->
          subject:
            projectUuid: project.subject.uuid
        secrets: () ->
          []

    .state 'environment',
      parent: "projects/edit"
      url: "/environments/{uuid}"
      resolve:
        environment: (environmentResource, $stateParams) ->
          environmentResource.find($stateParams.uuid)

    .state "environment.edit",
      url: "/edit"
      views:
        "main@layout":
          controller: "environmentFormCtrl"
          controllerAs: "environmentForm"
          templateUrl: 'views/environments/edit.html'
      resolve:
        secrets: (environment) ->
          environment.secrets()

    .state "secrets/show",
      parent: "projects/edit"
      url: "/secrets/{uuid}"
      views:
        "main@layout":
          templateUrl: 'views/secrets/show.html'
          controller: "secretShowCtrl"
          controllerAs: "secretShow"
      resolve:
        organization: (project) ->
          project.organization()
        project: (environment) ->
          environment.project()
        environment: (secret) ->
          secret.environment()
        secret: ($stateParams, secretResource) ->
          secretResource.find($stateParams.uuid)
