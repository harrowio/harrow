angular.module('harrowApp').config ($stateProvider) ->
  $stateProvider
    .state 'wizard',
      abstract: true
      parent: 'layout'
      url: '/a/wizard'
      data:
        requiresAuth: true
        showViews: []
        sectionsToComplete: 3
      views:
        "main@layout":
          controller: 'wizardCtrl'
          controllerAs: 'ctrl'
          templateUrl: 'views/wizard/main.html'
        'app-sidebar@layout':
          controller: "wizardCtrl"
          templateUrl: 'views/_app-sidebar.html'
          controllerAs: 'appSidebar'

    .state 'wizard.quick-start',
      url: '/quickStart'
      views:
        'main@wizard':
          controller: 'wizardQuickStartCtrl'
          controllerAs: 'ctrl'
          templateUrl: 'views/wizard/quick-start.html'

    .state 'wizard.create',
      url: '/create?{organizationUuid,quickStartFailed}'
      views:
        'main@wizard':
          controller: "wizardCreateCtrl"
          controllerAs: "wizardCreate"
          templateUrl: 'views/wizard/create.html'
      resolve:
        organization: (organizationResource, $stateParams) ->
          if $stateParams.organizationUuid
            organizationResource.find($stateParams.organizationUuid)
          else
            subject:
              public: false
              planUuid: "b99a21cc-b108-466e-aa4d-bde10ebbe1f3"
        project: (organization) ->
          subject:
            organizationUuid: organization.subject.uuid

    .state 'wizard.project',
      url: '/project/{projectUuid}'
      abstract: true
      template: '<div ui-view/>'
      data:
        completedSections: [
          'wizard.create'
        ]
      resolve:
        project: (projectResource, $stateParams) ->
          projectResource.find($stateParams.projectUuid)

    .state 'wizard.project.connect',
      url: '/connect'
      data:
        isNew: true
        nextState:
          accessible: 'wizard.project.stencils'
          ssh: 'wizard.project.connect.repo.ssh'
          https: 'wizard.project.connect.repo.private'
      views:
        'main@wizard':
          controller: 'repositoryCtrl',
          controllerAs: 'ctrl'
          templateUrl: 'views/repositories/edit.html'
      resolve:
        credential: () ->
          null
        repository: (project, repositoryResource, $stateParams) ->
          subject:
            projectUuid: project.subject.uuid

    .state 'wizard.project.connect.repo',
      abstract: true
      url: '/{repositoryUuid}'
      resolve:
        credential: (repository) ->
          repository.credential()
        repository: ($stateParams, repositoryResource) ->
          repositoryResource.find($stateParams.repositoryUuid)

    .state 'wizard.project.connect.repo.ssh',
      url: '/ssh-key'
      views:
        'main@wizard':
          controller: 'repositoryCtrl',
          controllerAs: 'ctrl'
          templateUrl: 'views/repositories/ssh.html'

    .state 'wizard.project.connect.repo.private',
      url: '/private'
      views:
        'main@wizard':
          controller: 'repositoryCtrl',
          controllerAs: 'ctrl'
          templateUrl: 'views/repositories/private.html'

    .state 'wizard.project.stencils',
      url: '/stencils'
      data:
        completedSections: [
          'wizard.create'
          'wizard.project.connect'
        ]
      views:
        'main@wizard':
          controller: 'wizardStencilsCtrl'
          controllerAs: 'ctrl'
          templateUrl: 'views/wizard/stencils.html'

    .state 'wizard.project.finished',
      url: '/finished'
      data:
        completedSections: [
          'wizard.create'
          'wizard.project.connect'
          'wizard.project.stencils'
        ]
      views:
        'main@wizard':
          controller: 'wizardCtrl'
          controllerAs: 'ctrl'
          templateUrl: 'views/wizard/finished.html'

angular.module('harrowApp').run ($transitions, $state) ->
  $transitions.onBefore {to:'wizard.create'}, ($transition$) ->
    params = $transition$.params('to')
    if params.organizationUuid
      $transition$.to().data.showViews.push 'sidebar'
    else
      $transition$.to().data.showViews = []
