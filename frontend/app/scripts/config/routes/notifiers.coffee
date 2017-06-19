notifierFactory = (
  $stateProvider
  baseState = 'notifiers'
  nextState = 'notifiers'
  parent = 'projects/edit'
) ->
  $stateProvider
    .state baseState,
      parent: parent
      url: '/notifiers'
      data:
        requiresAuth: true
      views:
        "main@layout":
          controller: 'notifiersCtrl',
          controllerAs: 'ctrl'
          templateUrl: 'views/notifiers/index.html'
      resolve:
        tasks: (project) ->
          project.tasks()
        scripts: (project) ->
          project.scripts()
        environments: (project) ->
          project.environments()
        notifiers: ($q, project, tasks) ->
          project.notifiers().then (response) ->
            response

    .state "#{baseState}.slackNotifier",
      url: '/slack'
      data:
        isNew: true
        nextState: nextState
      views:
        "main@layout":
          controller: 'notifierCtrl',
          controllerAs: 'ctrl'
          templateUrl: 'views/notifiers/slack.html'
      resolve:
        notifierType: ->
          'slackNotifier'
        notifierResource: (slackNotifierResource)->
          slackNotifierResource
        notifier: (project) ->
          subject:
            projectUuid: project.subject.uuid

    .state "#{baseState}.slackNotifier.edit",
      url: '/{uuid}'
      data:
        isNew: false
      views:
        "main@layout":
          controller: 'notifierCtrl',
          controllerAs: 'ctrl'
          templateUrl: 'views/notifiers/slack.html'
      resolve:
        notifierType: ->
          'slackNotifier'
        notifierResource: (slackNotifierResource)->
          slackNotifierResource
        notifier: (slackNotifierResource, $stateParams) ->
          slackNotifierResource.find($stateParams.uuid)

    .state "#{baseState}.emailNotifier",
      url: '/email'
      data:
        isNew: true
        nextState: nextState
      views:
        "main@layout":
          controller: 'notifierCtrl',
          controllerAs: 'ctrl'
          templateUrl: 'views/notifiers/email.html'
      resolve:
        notifierType: ->
          'emailNotifier'
        notifierResource: (emailNotifierResource)->
          emailNotifierResource
        notifier: (project) ->
          subject:
            projectUuid: project.subject.uuid

    .state "#{baseState}.emailNotifier.edit",
      url: '/{uuid}'
      data:
        isNew: false
      views:
        "main@layout":
          controller: 'notifierCtrl',
          controllerAs: 'ctrl'
          templateUrl: 'views/notifiers/email.html'
      resolve:
        notifier: (emailNotifierResource, $stateParams) ->
          emailNotifierResource.find($stateParams.uuid)

    .state "#{baseState}.taskNotifier",
      url: '/task'
      data:
        isNew: true
        nextState: nextState
      views:
        "main@layout":
          controller: 'notifierCtrl',
          controllerAs: 'ctrl'
          templateUrl: 'views/notifiers/task.html'
      resolve:
        notifierType: ->
          'taskNotifier'
        notifierResource: (taskNotifierResource)->
          taskNotifierResource
        notifier: (project) ->
          item = subject:
            projectUuid: project.subject.uuid
            triggerAction: 'operation.succeeded'
          item

    .state "#{baseState}.taskNotifier.edit",
      url: '/{uuid}'
      data:
        isNew: false
      views:
        "main@layout":
          controller: 'notifierCtrl',
          controllerAs: 'ctrl'
          templateUrl: 'views/notifiers/task.html'
      resolve:
        notifier: (taskNotifierResource, $stateParams) ->
          taskNotifierResource.find($stateParams.uuid)

angular.module('harrowApp').config ($stateProvider) ->

  notifierFactory($stateProvider)
  notifierFactory($stateProvider, 'task.edit.notifiers', 'task.edit.notification-rules', 'task.edit')
