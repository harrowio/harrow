angular.module('harrowApp').config ($stateProvider) ->
  $stateProvider
    .state "invitations/show",
      parent: "layout_tight"
      url: "/a/invitations/{uuid}"
      views:
        main:
          templateUrl: 'views/invitations/show.html'
          controller: "invitationShowCtrl"
          controllerAs: "invitationShow"
      resolve:
        invitation: (invitationResource, $stateParams) ->
          invitationResource.find($stateParams.uuid)

    .state "invitations/create",
      parent: "projects/edit"
      url: "/a/invitations/create/project={projectUuid}"
      views:
        "main@layout":
          templateUrl: 'views/invitations/create.html'
          controller: "invitationFormCtrl"
          controllerAs: "invitationForm"
      resolve:
        project: (projectResource, $stateParams) ->
          projectResource.find($stateParams.projectUuid)
        organization: (project) ->
          project.organization()
        invitation: ($stateParams) ->
          subject:
            projectUuid: $stateParams.projectUuid
            membershipType: "member"
