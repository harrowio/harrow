angular.module('harrowApp').filter 'toTrusted', ($sce) ->
  (text) ->
    $sce.trustAsHtml(text)
