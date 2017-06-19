app = angular.module("harrowApp")

app.factory "Activity", ($injector, $http) ->
  Activity = (data) ->
    $.extend(true, @, data)
    @

  Activity::markAsRead = () ->
    $http.put(@_links['read-status'].href)

  Activity


app.factory "activityResource", (Resource, Activity) ->
  ActivityResource = () ->
    Resource.call(@)
    @

  ActivityResource:: = Object.create(Resource::)
  ActivityResource::basepath = "/activities"
  ActivityResource::model = Activity

  new ActivityResource()
