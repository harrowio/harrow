BaseCtrl = () ->
  @
BaseCtrl::itemFor = (uuid, resourceName) ->
  items = @[resourceName].filter (item) ->
    item.subject.uuid == uuid
  items[0] if items.length

angular.module('harrowApp').controller 'baseCtrl', BaseCtrl
