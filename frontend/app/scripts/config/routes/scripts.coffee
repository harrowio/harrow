angular.module('harrowApp').config ($stateProvider) ->
  $stateProvider
    .state 'scripts',
      parent: 'projects/edit'
      url: "/scripts"
      views:
        "main@layout":
          controller: 'scriptsCtrl'
          controllerAs: 'ctrl'
          templateUrl: 'views/scripts/index.html'
      resolve:
        webhooks: (project) ->
          project.webhooks()
        repositories: (project) ->
          project.repositories()
        scripts: (project) ->
          project.scripts()
        tasks: (project) ->
          project.tasks()
        environments: (project) ->
          project.environments()

    .state 'createScript',
      parent: "projects/show"
      url: "/scripts/new"
      data:
        isNew: true
        breadcrumbs: ['organization', 'project', 'script']
      views:
        "main@layout":
          controller: 'scriptEditCtrl'
          controllerAs: 'ctrl'
          templateUrl: 'views/scripts/edit.html'
        "header@layout":
          controller: 'breadcrumbsCtrl'
          controllerAs: 'ctrl'
          templateUrl: 'views/_header.html'
      resolve:
        environments: (project) ->
          project.environments()
        tasks: ->
          []
        script: (project, environments, uuid, currentUser) ->
          data = subject:
            uuid: uuid()
            projectUuid: project.subject.uuid
            body: "#!/bin/bash -e\n\ndate"
            name: "#{currentUser.subject.name}'s script"
          data
        testScript: (environment, project, script) ->
          script: script.subject
          secrets: []
          projectUuid: project.subject.uuid
          environment: environment.subject
        currentUser: (authentication) ->
          if authentication.currentUser
            return authentication.currentUser
          authentication.loadSession().then ->
            authentication.currentUser
        environment: (environments, script) ->
          env = environments.find (env) ->
            env.subject.name == 'Default'
          unless env
            env = environments[0]
          env
        triggers: (project) ->
          []
        notifiers: (project) ->
          []
        scriptTasks: ->
          []
        secrets: ($q, environments) ->
          promises = []
          environments.forEach (environment) ->
            promises.push environment.secrets()

          $q.all(promises).then (secretsCollection) ->
            array = []
            secretsCollection.forEach (secrets) ->
              secrets.forEach (secret) ->
                array.push secret
            array


    .state 'script',
      parent: "projects/show"
      url: '/scripts/{scriptUuid}'
      data:
        breadcrumbs: ['organization', 'project', 'script']
        isNew: false
      resolve:
        environments: (project) ->
          project.environments()
        tasks: (project) ->
          project.tasks()
        scriptTasks: ($q, script, project, tasks, environments, notifiers, triggers) ->
          scriptTasks = tasks.filter (task) ->
            task.subject.scriptUuid == script.subject.uuid
          promises = []
          scriptTasks.forEach (task) ->
            task._embedded.environments = environments.filter (env) ->
              env.subject.uuid == task.subject.environmentUuid
            task._embedded.environments.forEach (env) ->
              env._embedded.tasks = [task]
            task._embedded.triggers = {}
            Object.keys(triggers).forEach (key) ->
              task._embedded.triggers[key] = triggers[key].filter (trigger) ->
                trigger.subject.taskUuid == task.subject.uuid
            promises.push task.notificationRules().then (rules) ->
              task._embedded.notifiers = task._filterNotifiers(notifiers, rules)
          $q.all(promises).then () ->
            scriptTasks

        script: (scriptResource, $stateParams) ->
          scriptResource.find($stateParams.scriptUuid)
        notifiers: (project) ->
          project.notifiers()
        triggers: (project) ->
          project.triggers()

      views:
        "main@layout":
          templateUrl: 'views/scripts/show.html'
          controller: 'scriptCtrl'
          controllerAs: 'ctrl'
        "header@layout":
          controller: 'breadcrumbsCtrl'
          controllerAs: 'ctrl'
          templateUrl: 'views/_header.html'

    .state 'script.edit',
      url: '/edit'
      views:
        "main@layout":
          controller: 'scriptEditCtrl'
          controllerAs: 'ctrl'
          templateUrl: 'views/scripts/edit.html'
      resolve:
        repositories: (project) ->
          project.repositories()
        environment: (environments, script) ->
          env = environments.find (env) ->
            env.subject.name == 'Default'
          unless env
            env = environments[0]
          env
        testScript: (environment, project, script) ->
          script: script.subject
          secrets: []
          projectUuid: project.subject.uuid
          environment: environment.subject
        secrets: ($q, environments) ->
          promises = []
          environments.forEach (environment) ->
            promises.push environment.secrets()

          $q.all(promises).then (secretsCollection) ->
            array = []
            secretsCollection.forEach (secrets) ->
              secrets.forEach (secret) ->
                array.push secret
            array
