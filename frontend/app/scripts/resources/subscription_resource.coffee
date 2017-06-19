app = angular.module("harrowApp")

app.factory "Subscription", ($injector, $http) ->
  Subscription = (data) ->
    $.extend(true, @, data)
    @userResource = $injector.get("userResource")
    @

  Subscription::watcher = () ->
    @userResource.fetch(@_links.watcher.href)

  Subscription::isWatching = () ->
    subscribed = (subscribed for _, subscribed of @subject.subscribed)
    subscribed.reduce (acc, item) ->
      acc && item

  Subscription

app.factory "subscriptionResource", (Resource, Subscription) ->
  SubscriptionResource = () ->
    Resource.call(@)
    @

  SubscriptionResource:: = Object.create(Resource::)
  SubscriptionResource::basepath = "/subscriptions"
  SubscriptionResource::model = Subscription

  new SubscriptionResource()
