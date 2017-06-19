app = angular.module("harrowApp")

app.factory "Invitation", ($injector, $http) ->
  Invitation = (data) ->
    $.extend(true, @, data)
    @projectResource = $injector.get("projectResource")
    @organizationResource = $injector.get("organizationResource")
    @userResource = $injector.get("userResource")
    @

  Invitation::isAccepted = () ->
    !!@subject.acceptedAt

  Invitation::isRefused = () ->
    !!@subject.refusedAt

  Invitation::accept = () ->
    $http.patch(@_links.self.href, { accept: "accept" })

  Invitation::refuse = () ->
    $http.patch(@_links.self.href, { accept: "refuse" })

  Invitation::creator = () ->
    @userResource.fetch(@_links.creator.href)

  Invitation::project = () ->
    @projectResource.fetch(@_links.project.href)

  Invitation::organization = () ->
    @organizationResource.fetch(@_links.organization.href)

  Invitation

app.factory "invitationResource", (Resource, Invitation) ->
  InvitationResource = () ->
    Resource.call(@)
    @

  InvitationResource:: = Object.create(Resource::)
  InvitationResource::basepath = "/invitations"
  InvitationResource::model = Invitation

  new InvitationResource()
