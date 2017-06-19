angular.module('harrowApp').filter 'formattedOperationTrigger', ($filter, $sce) ->
  (operation) ->
    reason = operation.subject.parameters.reason
    if reason == 'user'
      out = "Last run manually by <u>#{operation.subject.parameters.username}</u>"
    if reason == 'schedule'
      out = "Last scheduled"
    if reason == 'webhook'
      out = "Last triggered by Webhook"
    if reason == 'git-trigger'
      out = "Last triggered by change in repository"
    out = "#{out} #{$filter('amTimeAgo')(operation.subject.createdAt)}"
    $sce.trustAsHtml(out)
