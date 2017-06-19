app = angular.module("harrowApp")

WebhookFormCtrl = (
  @organization
  @project
  @webhook
  @tasks
  @flash
  @$state
  @$translate
  @webhookResource
  @$q
  @$scope
) ->

  @

WebhookFormCtrl::save = () ->
  @webhookResource.save(@webhook).then (webhook) =>
    @flash.success = @$translate.instant("forms.webhookForm.flashes.success", webhook.subject)
    @$state.go("projects/edit.triggers", {projectUuid: webhook.subject.projectUuid})
    return
  .catch (reason) =>
    @flash.error = @$translate.instant("forms.webhookForm.flashes.fail", @webhook.subject)
    @$q.reject(reason)

app.controller("webhookFormCtrl", WebhookFormCtrl)
