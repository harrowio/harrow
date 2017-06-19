app = angular.module("harrowApp")

app.factory "Delivery", ($injector) ->
  Delivery = (data) ->
    $.extend(true, @, data)
    @webhookResource = $injector.get("webhookResource")
    @

  Delivery::webhook = () ->
    @webhookResource.fetch(@_links.webhook.href)

  Delivery

app.factory "deliveryResource", (Resource, Delivery) ->
  DeliveryResource = () ->
    Resource.call(@)
    @

  DeliveryResource:: = Object.create(Resource::)
  DeliveryResource::basepath = "/deliveries"
  DeliveryResource::model = Delivery

  new DeliveryResource()
