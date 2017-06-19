app = angular.module("harrowApp")

app.factory "oauth", (
  $http
  $window
  $state
  endpoint
  authentication
  Session
) ->
  Oauth = () ->
    @

  Oauth::callbackGithub = (params) ->
    if params.action == "signin"
      @callbackGithubSignin(params)
    else
      @callbackGithubAuthorize(params)

  Oauth::authorizeGithub = () ->
    $http.get(endpoint + "/oauth/github/authorize").success (data, status, headers, config) ->
      $window.location = headers().location

  Oauth::callbackGithubAuthorize = (params) ->
    $http(
      method: "POST"
      url: endpoint + "/oauth/github/callback/authorize"
      params: params
    ).success (data, status, headers, config) ->
      data

  Oauth::callbackGithubSignin = (params) ->
    $http(
      method: "POST"
      url: endpoint + "/oauth/github/callback/signin"
      params: params
    ).then (response) ->
      authentication.setSession(new Session(response.data))

  Oauth::signinGithub = () ->
    $http.get(endpoint + "/oauth/github/signin").success (data, status, headers, config) ->
      $window.location = headers().location

  Oauth::deauthorizeGithub = () ->
    $http.get(endpoint + "/oauth/github/deauthorize")

  Oauth::pingGithub = () ->
    $http(
      method: "GET"
      url: endpoint + "/oauth/github/ping"
    ).success (data, status, headers, config) ->
      data

  Oauth::getGithubUser = () ->
    $http(
      method: "GET"
      url: endpoint + "/oauth/github/user"
    ).success (data, status, headers, config) ->
      data


  Oauth::getGithubOrganizations = () ->
    $http(
      method: "GET"
      url: endpoint + "/oauth/github/organizations"
    ).success (data, status, headers, config) ->
      data

  Oauth::getGithubUserRepositories = (login) ->
    $http(
      method: "GET"
      url: endpoint + "/oauth/github/repositories/users/" + login
    ).success (data, status, headers, config) ->
      data

  Oauth::getGithubOrgRepositories = (login) ->
    $http(
      method: "GET"
      url: endpoint + "/oauth/github/repositories/orgs/" + login
    ).success (data, status, headers, config) ->
      data

  Oauth::createGithubDeployKey = (repoUuid) ->
    $http(
      method: "POST"
      url: endpoint + "/oauth/github/repositories/#{repoUuid}/keys"
    ).success (data, status, headers, config) ->
      data

   new Oauth()
