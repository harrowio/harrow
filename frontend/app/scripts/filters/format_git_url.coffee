angular.module('harrowApp').filter 'formatGitUrl', (
  $filter
) ->
  (input) ->
    url = $filter('url')(input)
    if url
      url.toString('humanFormat')
