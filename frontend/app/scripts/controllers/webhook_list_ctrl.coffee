app = angular.module("harrowApp")

WebhookListCtrl = (
  @$controller
  @project
  @environments
  @webhookResource
  @webhooks
  @tasks
  @scripts
  @flash
  @$translate
) ->
  $.extend(true, @, $controller('baseCtrl'))

  @taskNames = {}
  @webhooks = @webhooks.filter (webhook) ->
    !webhook.subject.name.startsWith("urn:harrow:")

  angular.forEach @tasks, (task) =>
    @taskNames[task.subject.uuid] = task.subject.name
  @

WebhookListCtrl::delete = (webhook) ->
  if confirm(@$translate.instant("prompts.really?"))
    @webhookResource.delete(webhook.subject.uuid).then =>
      @flash.success = @$translate.instant("webhooks.flashes.delete.success", webhook)
      @webhooks = @webhooks.filter( (hook) -> hook.subject.uuid != webhook.subject.uuid )
      return
    .catch =>
      @flash.error = @$translate.instant("webhooks.flashes.delete.failure", webhook)
      return

WebhookListCtrl::regenerate = (webhook) ->
  return unless confirm(@$translate.instant("webhooks.prompts.regenerate"))
  webhook.regenerateSlug().then (response) ->
    webhook.subject = response.data.subject
    webhook._links = response.data._links
    return

WebhookListCtrl::environmentFor = (hook) ->
  @itemFor(@itemFor(hook.subject.taskUuid, 'tasks').subject.environmentUuid, 'environments')

WebhookListCtrl::scriptFor = (hook) ->
  @itemFor(@itemFor(hook.subject.taskUuid, 'tasks').subject.scriptUuid, 'scripts')

app.controller("webhookListCtrl", WebhookListCtrl)
