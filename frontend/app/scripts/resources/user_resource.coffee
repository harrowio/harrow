app = angular.module("harrowApp")

app.factory "User", ($injector, $http) ->
  User = (data) ->
    $.extend(true, @, data)
    @sessionResource = $injector.get("sessionResource")
    @organizationResource = $injector.get("organizationResource")
    @projectResource = $injector.get("projectResource")
    @activityResource = $injector.get("activityResource")
    @taskResource = $injector.get("taskResource")
    @_links.tasks = @_links.jobs
    @

  User::sessions = () ->
    @sessionResource.fetch(@_links.sessions.href)

  User::organizations = () ->
    @organizationResource.fetch(@_links.organizations.href)

  User::activities = () ->
    @activityResource.fetch(@_links.activities.href)

  User::projects = () ->
    @projectResource.fetch(@_links.projects.href)

  User::projects_with_memberships = () ->
    @projectResource.fetch(@_links.projects.href, membershipOnly: "yes")

  User::tasks = () ->
    @taskResource.fetch(@_links.tasks.href)

  User::mfaEnabled = () ->
    @totpEnabled()

  User::totpEnabled = () ->
    @subject.totpEnabledAt?

  User::requestTotpSecret = () ->
    $http.patch(@_links.mfa.href, {totpGenerateSecret: true}).then (response) =>
      @subject.totpSecret = response.data.subject.totpSecret

  User::enableTotp = (totp) ->
    $http.patch(@_links.mfa.href, {twoFactorAuthEnabled: true, totpToken: parseInt(totp, 10)}).then (response) =>
      @subject.totpEnabledAt = response.data.subject.totpEnabledAt

  User::disableTotp = (totp) ->
    $http.patch(@_links.mfa.href, {twoFactorAuthEnabled: false, totpToken: parseInt(totp, 10)}).then (response) =>
      @subject.totpEnabledAt = null

  User::isGuest = () ->
    /@localhost$/.test @subject.email

  User

app.factory "userResource", (Resource, User) ->
  UserResource = () ->
    Resource.call(@)
    @

  UserResource:: = Object.create(Resource::)
  UserResource::basepath = "/users"
  UserResource::model = User

  new UserResource()
