angular.module('harrowApp').config ($stateProvider) ->
  $stateProvider
    .state "organizations/create",
      parent: "layout"
      url: "/a/organizations/create"
      data:
        requiresAuth: true
      views:
        main:
          templateUrl: 'views/organizations/create.html'
          controller: "organizationFormCtrl"
          controllerAs: "organizationForm"
      resolve:
        organization: () ->
          subject:
            public: false


    .state 'organization',
      parent: 'layout'
      url: "/a/organizations/{uuid}"
      data:
        showViews: [
          'sidebar'
          'header'
        ]
        breadcrumbs: ['organization']
      views:
        main:
          controller: "organizationShowCtrl"
          controllerAs: "ctrl"
          templateUrl: 'views/organizations/show.html'
        "header@layout":
          controller: 'breadcrumbsCtrl'
          controllerAs: 'ctrl'
          templateUrl: 'views/_header.html'
      resolve:
        organization: (organizationResource, $stateParams) ->
          organizationResource.find($stateParams.uuid)
        cards: (organization) -> organization.projectCards()

    .state 'organization.edit',
      url: '/edit'
      data:
        requiresAuth: true
        showViews: [
          'sidebar'
          'app-sidebar'
          'header'
        ]

      views:
        "main@layout":
          controller: "organizationShowCtrl"
          controllerAs: "ctrl"
          templateUrl: 'views/organizations/edit.html'
        "app-sidebar@layout":
          controller: (@menuItems) ->
            @menu = @menuItems.organizationEdit
            @
          controllerAs: 'appSidebar'
          templateUrl: 'views/_app-sidebar.html'
      resolve:
        limits: (organization) ->
          organization.limits()
    .state 'organization.edit.details',
      url: '/details'
      views:
        "main@layout":
          controller: "organizationShowCtrl"
          controllerAs: "ctrl"
          templateUrl: 'views/organizations/edit.html'
    .state 'organization.edit.archive',
      url: '/archive'
      views:
        "main@layout":
          controller: "organizationShowCtrl"
          controllerAs: "ctrl"
          templateUrl: 'views/organizations/archive.html'
