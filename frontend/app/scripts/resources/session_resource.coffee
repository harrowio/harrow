app = angular.module("harrowApp")

app.factory "Session", ($injector, $q) ->
  Session = (data) ->
    $.extend(true, @, data)
    @userResource = $injector.get("userResource")
    @$http = $injector.get("$http")
    @

  Session::confirm = (totp) ->
    @$http.patch(@_links.self.href, {totp: parseInt(totp, 10)})
    .then (response) =>
      @subject = response.data.subject
      @

  Session::user = () ->
    if @subject.valid
      @userResource.fetch(@_links.user.href)
    else
      $q.when()

  Session

app.factory "sessionResource", (Resource, Session) ->
  SessionResource = () ->
    Resource.call(@)
    @

  SessionResource:: = Object.create(Resource::)
  SessionResource::basepath = "/sessions"
  SessionResource::model = Session

  new SessionResource()
