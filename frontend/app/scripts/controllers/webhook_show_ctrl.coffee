app = angular.module("harrowApp")

WebhookShowCtrl = (
  @organization
  @project
  @webhook
  @deliveries
  @flash
  @$translate
) ->
  @

app.controller("webhookShowCtrl", WebhookShowCtrl)
