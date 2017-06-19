app = angular.module("harrowApp")

SecretShowCtrl = (
  @organization
  @project
  @environment
  @secret
  secretResource
  ws
  $scope
  $timeout
) ->
  cid = ws.subRow "secrets", @secret.subject.uuid, () =>
    secretResource.find(@secret.subject.uuid).then (secret) =>
      @secret = secret
      $timeout =>
        @$scope.$apply()

  $scope.$on "$destroy", =>
    ws.unsubscribe(cid)

  @

app.controller("secretShowCtrl", SecretShowCtrl)
