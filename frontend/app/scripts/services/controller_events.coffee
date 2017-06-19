app = angular.module("harrowApp")

app.value "initEvents", (ctrl, scope) ->
  for event in ctrl.events
    scope.$on event, angular.bind(ctrl, ctrl["#{event}Handler"])
