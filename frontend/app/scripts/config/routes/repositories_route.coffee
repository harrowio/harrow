angular.module('harrowApp').config ($stateProvider) ->
  $stateProvider
    .state 'repositories',
      parent: 'projects/edit'
      url: '/repositories'
      views:
        'main@layout':
          controller: 'repositoriesCtrl'
          controllerAs: 'ctrl'
          templateUrl: 'views/repositories/index.html'
      resolve:
        repositories: (project) ->
          project.repositories()

    .state 'createRepository',
      parent: 'projects/edit'
      url: '/repositories/new'
      data:
        isNew: true
        nextState:
          accessible: 'repositories'
          ssh: 'repository.edit.ssh'
          https: 'repository.edit.private'
      views:
        'main@layout':
          controller: 'repositoryCtrl',
          controllerAs: 'ctrl'
          templateUrl: 'views/repositories/edit.html'
      resolve:
        repository: (project) ->
          subject:
            projectUuid: project.subject.uuid
        credential: () ->
          angular.noop()

    .state 'repository',
      parent: 'projects/edit'
      url: '/repositories/{repositoryUuid}'
      views:
        'main@layout':
          controller: 'repositoryCtrl',
          controllerAs: 'ctrl'
          templateUrl: 'views/repositories/show.html'
      resolve:
        repository: ($stateParams, repositoryResource) ->
          repositoryResource.find($stateParams.repositoryUuid)
        credential: () ->
          angular.noop()

    .state 'repository.edit',
      url: '/edit'
      data:
        nextState:
          accessible: 'repositories'
          ssh: 'repository.edit.ssh'
          https: 'repository.edit.private'
      views:
        'main@layout':
          controller: 'repositoryCtrl',
          controllerAs: 'ctrl'
          templateUrl: 'views/repositories/edit.html'

    .state 'repository.edit.ssh',
      url: '/ssh-key'
      views:
        'main@layout':
          controller: 'repositoryCtrl',
          controllerAs: 'ctrl'
          templateUrl: 'views/repositories/ssh.html'
      resolve:
        credential: (repository) ->
          repository.credential()

    .state 'repository.edit.private',
      url: '/private'
      views:
        'main@layout':
          controller: 'repositoryCtrl',
          controllerAs: 'ctrl'
          templateUrl: 'views/repositories/private.html'
