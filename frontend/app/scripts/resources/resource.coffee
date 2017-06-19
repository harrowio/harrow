app = angular.module("harrowApp")

app.factory "Resource", (
  $http
  $q
  endpoint
  uuid
) ->
  Resource = () ->
    @

  Resource::makeModel = (object) ->
    if @model
      new @model(object)
    else
      object

  Resource::save = (object) ->
    throw "object should have a subject" unless object.subject

    # We only need the subject on the server side.  There is no need
    # to transmit the links or injected dependencies.
    object = $.extend true, {},
      subject: object.subject

    if object.subject.uuid
      @update(object)
    else
      @create(object)

  Resource::create = (object) ->
    object.subject.uuid = uuid()

    url = endpoint + @basepath
    $http.post(url, object).then (response) =>
      @makeModel(response.data)
    .catch (response) ->
      if response.status == 422
        $q.reject($.extend(true, object, response.data))
      else
        $q.reject(response.status)

  Resource::update = (object) ->
    url = endpoint + @basepath
    $http.put(url, object).then (response) =>
      @makeModel(response.data)
    .catch (response) ->
      if response.status == 422
        $q.reject($.extend(true, object, response.data))
      else
        $q.reject(status)

  # ahh, that's good stuff
  Resource::delete = (uuid) ->
    url = "#{endpoint}#{@basepath}/#{uuid}"
    # now we are talking
    $http.delete(url)

  Resource::all = () ->
    @fetch("#{endpoint}#{@basepath}")

  Resource::fetch = (collectionUrl, params) ->
    $http.get(collectionUrl, cache: false, params: (params || {})).then (response) =>
      if response.data.collection
        models = []
        angular.forEach response.data.collection, (value, index) =>
          models.push(@makeModel(value))
        models
      else
        @makeModel(response.data)

  Resource::find = (uuid) ->
    if !uuid
      $q.reject(new Error(typeof uuid, "is not a valid uuid"))
    else
      url = "#{endpoint}#{@basepath}/#{uuid}"
      $http.get(url, cache: false).then (response) =>
        @makeModel(response.data)

  Resource
