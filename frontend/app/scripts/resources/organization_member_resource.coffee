app = angular.module("harrowApp")

app.factory "OrganizationMember", ($injector, $http) ->
  OrganizationMember = (data) ->
    $.extend(true, @, data)
    @

  OrganizationMember

app.factory "organizationMemberResource", (Resource, OrganizationMember) ->
  OrganizationMemberResource = () ->
    Resource.call(@)
    @

  OrganizationMemberResource:: = Object.create(Resource::)
  OrganizationMemberResource::basepath = "/organization-members"
  OrganizationMemberResource::model = OrganizationMember

  new OrganizationMemberResource()
