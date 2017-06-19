SockJS = require('sockjs-client')

app = angular.module("harrowApp")

Ws = (
  $q
  @$rootScope
  @authentication
  @uuid
  @$log
  $timeout
) ->
  @subs = {}
  @socket = new SockJS("/ws")
  deferred = $q.defer()
  @socket.onopen = ->
    deferred.resolve()
  @socket.onmessage = (e) =>
    @receive(e)
  @socket.onclose = (e) =>
    @close(e)
    $timeout =>
      @$rootScope.$digest()
  @opened = deferred.promise
  @

Ws::subRow = (table, subscriptionUuid, handler) ->
  cid = @uuid()
  @send {command: "subRow", uuid: subscriptionUuid, table, cid}
  @subs[cid] = handler
  return cid

Ws::subLogevents = (operationUuid, handler) ->
  cid = @uuid()
  @send {command: "subLogevents", operationUuid, cid}
  @subs[cid] = handler
  return cid

Ws::unsubscribe = (cid) ->
  @send {cid, stop: true}
  delete @subs[cid]

Ws::send = (obj) ->
  @opened.then =>
    cmd = $.extend true, {}, obj, sessionUuid: @authentication.sessionUuid()
    @socket.send(angular.toJson(cmd))

Ws::receive = (e) ->
  data = e.data
  message = angular.fromJson(data)
  handler = @subs[message.cid]
  handler?(message)

Ws::close = (e) ->
  @$log.info("WebSocket closed, reason: #{e.reason}")
  @subs = {}
  @$rootScope.$emit "wsError", e

app.service "ws", Ws
