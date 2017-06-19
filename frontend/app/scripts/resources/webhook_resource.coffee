app = angular.module("harrowApp")

app.factory "Webhook", ($injector) ->
  Webhook = (data) ->
    $.extend(true, @, data)
    @projectResource = $injector.get("projectResource")
    @deliveryResource = $injector.get("deliveryResource")
    @$http = $injector.get("$http")
    if @subject
      @subject.taskUuid = @subject.jobUuid
    @

  Webhook::project = () ->
    @projectResource.fetch(@_links.project.href)

  Webhook::deliveries = () ->
    @deliveryResource.fetch(@_links.deliveries.href)

  Webhook::regenerateSlug = () ->
    @$http.patch(@_links.slug.href)

  Webhook

app.factory "webhookResource", (Resource, Webhook) ->
  WebhookResource = () ->
    Resource.call(@)
    @

  WebhookResource:: = Object.create(Resource::)
  WebhookResource::basepath = "/webhooks"
  WebhookResource::model = Webhook

  WebhookResource::_save = WebhookResource::save

  WebhookResource::save = (object) ->
    object.subject.jobUuid = object.subject.taskUuid
    @_save(object)

  new WebhookResource()
