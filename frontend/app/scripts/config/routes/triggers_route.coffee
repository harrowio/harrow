triggerFactory = (
  $stateProvider
  baseState = 'triggers'
  nextState = 'triggers'
  parent = 'projects/edit'
) ->
  baseObject =
    parent: parent
    url: '/triggers?{scriptUuid}'
    data:
      isNew: false
      nextState: nextState
      showViews: [
        'sidebar'
        'header'
        'app-sidebar'
      ]
    views:
      "main@layout":
        controller: 'triggersCtrl'
        controllerAs: 'triggers'
        templateUrl: 'views/triggers/index.html'
    resolve:
      triggers: (project) ->
        project.triggers()
      environments: (project) ->
        project.environments()
      scripts: (project) ->
        project.scripts()
      tasks: (project) ->
        project.tasks()
      repositories: (project) ->
        project.repositories()

  if parent == 'task.edit'
    baseObject.resolve.triggers = (task) ->
      task.triggers()
  else
    baseObject.resolve.task = ($q) ->
      $q.when()

  $stateProvider
    .state baseState, baseObject

    .state "#{baseState}.gitTrigger",
      url: '/git'
      views:
        "main@layout":
          controller: 'triggerCtrl'
          controllerAs: 'trigger'
          templateUrl: 'views/triggers/git.html'
      data:
        isNew: true
      resolve:
        triggerType: ->
          'gitTrigger'
        triggerResource: (gitTriggerResource) ->
          gitTriggerResource
        trigger: (project, task, tasks) ->
          trigger = subject:
            projectUuid: project.subject.uuid
            changeType: 'change'
            matchRef: 'master'
            repositoryUuid: null
          if tasks && tasks.length > 0
            trigger.subject.taskUuid = tasks[0].subject.uuid
          if task
            trigger.subject.taskUuid = task.subject.uuid
          trigger

    .state "#{baseState}.gitTrigger.onBranch",
      data:
        autoSave: true
      views:
        "main@layout":
          controller: 'triggerCtrl'
          controllerAs: 'trigger'
          templateUrl: 'views/triggers/git.html'
      resolve:
        trigger: (project, task, tasks) ->
          trigger = subject:
            projectUuid: project.subject.uuid
            name: 'Run when branch is added'
            changeType: 'add'
            matchRef: '*'
            repositoryUuid: null
          if tasks && tasks.length > 0
            trigger.subject.taskUuid = tasks[0].subject.uuid
          if task
            trigger.subject.taskUuid = task.subject.uuid
          trigger

    .state "#{baseState}.gitTrigger.onMasterCommit",
      data:
        autoSave: true
      views:
        "main@layout":
          controller: 'triggerCtrl'
          controllerAs: 'trigger'
          templateUrl: 'views/triggers/git.html'
      resolve:
        trigger: (project, task, tasks) ->
          trigger = subject:
            projectUuid: project.subject.uuid
            name: 'Run when master changes'
            changeType: 'change'
            matchRef: 'master'
            repositoryUuid: null
          if tasks && tasks.length > 0
            trigger.subject.taskUuid = tasks[0].subject.uuid
          if task
            trigger.subject.taskUuid = task.subject.uuid
          trigger

    .state "#{baseState}.gitTrigger.edit",
      url: '/{uuid}'
      views:
        "main@layout":
          controller: 'triggerCtrl'
          controllerAs: 'trigger'
          templateUrl: 'views/triggers/git.html'
      data:
        isNew: false
      resolve:
        triggerType: ->
          'git'
        triggerResource: (gitTriggerResource) ->
          gitTriggerResource
        trigger: ($stateParams, triggerResource) ->
          triggerResource.find($stateParams.uuid)

    .state "#{baseState}.webhook",
      url: '/webhook'
      views:
        "main@layout":
          controller: 'triggerCtrl'
          controllerAs: 'trigger'
          templateUrl: 'views/triggers/webhook.html'
      data:
        isNew: true
      resolve:
        triggerType: ->
          'webhook'
        triggerResource: (webhookResource) ->
          webhookResource
        trigger: (project, task, tasks) ->
          trigger = subject:
            projectUuid: project.subject.uuid
          if tasks && tasks.length > 0
            trigger.subject.taskUuid = tasks[0].subject.uuid
          if task
            trigger.subject.taskUuid = task.subject.uuid
          trigger

    .state "#{baseState}.webhook.edit",
      url: '/{uuid}'
      views:
        "main@layout":
          controller: 'triggerCtrl'
          controllerAs: 'trigger'
          templateUrl: 'views/triggers/webhook.html'
      data:
        isNew: false
      resolve:
        triggerType: ->
          'webhook'
        triggerResource: (webhookResource) ->
          webhookResource
        trigger: ($stateParams, triggerResource) ->
          triggerResource.find($stateParams.uuid)

    .state "#{baseState}.schedule",
      url: '/schedule'
      views:
        "main@layout":
          controller: 'triggerCtrl'
          controllerAs: 'trigger'
          templateUrl: 'views/triggers/schedule.html'
      data:
        isNew: true
      resolve:
        triggerType: ->
          'schedule'
        triggerResource: (scheduleResource) ->
          scheduleResource
        trigger: (project, task, tasks) ->
          trigger = subject:
            projectUuid: project.subject.uuid
            description: "Ad-hoc"
          if tasks && tasks.length > 0
            trigger.subject.taskUuid = tasks[0].subject.uuid
          if task
            trigger.subject.taskUuid = task.subject.uuid
          trigger

    .state "#{baseState}.schedule.nighly",
      data:
        autoSave: true
      views:
        "main@layout":
          controller: 'triggerCtrl'
          controllerAs: 'trigger'
          templateUrl: 'views/triggers/schedule.html'
      resolve:
        trigger: (project, task, tasks) ->
          trigger = subject:
            projectUuid: project.subject.uuid
            description: "Every night at 8pm"
            scheduleType: 'cronspec'
            cronspec: '0 20 * * *'
          if tasks && tasks.length > 0
            trigger.subject.taskUuid = tasks[0].subject.uuid
          if task
            trigger.subject.taskUuid = task.subject.uuid
          trigger

    .state "#{baseState}.schedule.weekly",
      data:
        autoSave: true
      views:
        "main@layout":
          controller: 'triggerCtrl'
          controllerAs: 'trigger'
          templateUrl: 'views/triggers/schedule.html'
      resolve:
        trigger: (project, task, tasks) ->
          trigger = subject:
            projectUuid: project.subject.uuid
            description: "Every week"
            scheduleType: 'cronspec'
            cronspec: '0 0 * * 1'
          if tasks && tasks.length > 0
            trigger.subject.taskUuid = tasks[0].subject.uuid
          if task
            trigger.subject.taskUuid = task.subject.uuid
          trigger

    .state "#{baseState}.schedule.edit",
      url: '/{uuid}'
      views:
        "main@layout":
          controller: 'triggerCtrl'
          controllerAs: 'trigger'
          templateUrl: 'views/triggers/schedule.html'
      data:
        isNew: false
      resolve:
        triggerType: ->
          'schedule'
        triggerResource: (scheduleResource) ->
          scheduleResource
        trigger: ($stateParams, triggerResource) ->
          triggerResource.find($stateParams.uuid)

    .state "#{baseState}.taskNotifier",
      url: '/task'
      views:
        "main@layout":
          controller: 'triggerCtrl'
          controllerAs: 'trigger'
          templateUrl: 'views/triggers/task.html'
      data:
        isNew: true
      resolve:
        triggerType: ->
          'task'
        triggerResource: (taskNotifierResource) ->
          taskNotifierResource
        trigger: (project, task, tasks) ->
          trigger = subject:
            projectUuid: project.subject.uuid
          if tasks && tasks.length > 0
            trigger.subject.taskUuid = tasks[0].subject.uuid
          if task
            trigger.subject.taskUuid = task.subject.uuid
          trigger

    .state "#{baseState}.taskNotifier.edit",
      url: '/{uuid}'
      views:
        "main@layout":
          controller: 'triggerCtrl'
          controllerAs: 'trigger'
          templateUrl: 'views/triggers/task.html'
      data:
        isNew: false
      resolve:
        triggerType: ->
          'task'
        triggerResource: (taskNotifierResource) ->
          taskNotifierResource
        trigger: ($stateParams, triggerResource) ->
          triggerResource.find($stateParams.uuid)


angular.module('harrowApp').config ($stateProvider) ->
  triggerFactory($stateProvider)
  triggerFactory($stateProvider, 'task.edit.triggers', 'task.edit.triggers', 'task.edit')
