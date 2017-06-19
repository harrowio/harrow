app = angular.module("harrowApp")

GithubCallbackCtrl = (
  action
  authentication
  oauth
  @$window
  $state
  $translate
  flash
) ->
  oauth.callbackGithub
    state: @param("state")
    code: @param("code")
    action: action
  .then () =>
    if action == "authorize"
      # HACK: get rid of the url params from the github redirect
      # $state can only modify the hash
      @$window.location = "/#/a/settings/oauth"
    else
      if !authentication.currentSession.subject.valid
        @$window.location = "/#/a/session-confirmation"
      else
        @$window.location = "/#/a/dashboard"
    return
  .catch (resp) =>
    if resp.data.reason == "oauth.github.existing_unlinked_user"
      @$window.location = "/#/a/errors/github_existing_unlinked_user"
    else
      flash.error = $translate.instant("oauth.githubCallback.error")
    return

  @

# HACK: stolen from http://snipplr.com/view/26662/get-url-parameters-with-jquery--improved/
#
#
# $location.search().state won't work, because in non-html5-mode the location service does not consider anything before
# the hash to be part of its state.
# For example, in "/?foo=bar#/a/route", $location.search().foo would be undefined,
# but in "/#/a/route?foo=bar" it would be defined
GithubCallbackCtrl::param = (name) ->
  results = new RegExp("[\\?&]" + name + "=([^&#]*)").exec(@$window.location.search)
  return if results?.length < 0
  @$window.decodeURIComponent(results[1])

app.controller("githubCallbackCtrl", GithubCallbackCtrl)
