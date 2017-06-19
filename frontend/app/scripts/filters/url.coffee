angular.module('harrowApp').filter 'isSsh', ($filter) ->
  protocols = $filter('protocols')
  isSsh = (input) ->
    if Array.isArray(input)
      return input.indexOf('ssh') != -1 || input.indexOf('rsync') != -1
    if typeof input != 'string'
      return false

    if isSsh(protocols(input))
      return true

    input = input.substring(input.indexOf('://') + 3)
    input.indexOf('@') < input.indexOf(':')
  isSsh

angular.module('harrowApp').filter 'protocols', ->
  (input) ->
    index = input.indexOf('://')
    input.substring(0, index).split('+').filter(Boolean)

angular.module('harrowApp').filter 'url', ($filter) ->
  protocols = $filter('protocols')
  isSsh = $filter('isSsh')

  (url, format) ->
    return url unless url
    output = {
      protocol: ''
      port: ''
      hostname: ''
      host: ''
      user: ''
      username: ''
      password: ''
      search: ''
      pathname: ''
      isPrivate: ''
      isGitHub: false
      isBitbucket: false
    }
    output.isSsh = isSsh(url)
    output.href = url
    output.protocols = protocols(url)
    output.protocol = output.protocols[0] || (isSsh(url) && 'ssh' || 'file')

    protocolIndex = url.indexOf('://')
    if protocolIndex != -1
      url = url.substring(protocolIndex + 3)

    parts = url.split('/')
    output.hostname = parts.shift()

    # user@domain
    splits = output.hostname.split('@')
    if splits.length == 2
      output.user = splits[0]
      output.hostname = splits[1]

    # user:pass
    if output.user
      output.username = output.user
    splits = output.user.split(':')
    if splits.length == 2
      output.username = splits[0]
      output.password = splits[1]

    # domain.ext:port
    splits = output.hostname.split(':')
    if splits.length == 2
      output.hostname = splits[0]
      output.port = parseInt(splits[1])
      if isNaN(output.port)
        output.port = null
        parts.unshift(splits[1])

    # Remove empty parts
    parts = parts.filter(Boolean)

    # Stringify parthname
    output.pathname = "/#{parts.join('/')}"

    # #some-hash
    splits = output.pathname.split('#')
    if splits.length == 2
      output.pathname = splits[0]
      output.hash = splits[1]

    splits = output.pathname.split('?')
    if splits.length == 2
      output.pathname = splits[0]
      output.search = splits[1]

    if output.hostname == 'github.com'
      output.isGitHub = true

    if output.hostname == 'bitbucket.org'
      output.isBitbucket = true

    if output.isSsh || output.password
      output.isPrivate = true

    output.toString = (type) ->
      pathname = output.pathname.substring(1).replace(/\.git$/, '')
      auth = ''
      unless output.isSsh
        if output.password
          auth = "#{output.username}:#{output.password}@"
        else if output.username
          auth = "#{output.username}@"

      switch type
        when 'ssh'
          return "git@#{output.hostname}:#{pathname}.git"
        when 'git+ssh'
          return "git+ssh://git@#{output.hostname}/#{pathname}.git"
        when 'http', 'https'
          return "#{type}://#{auth}#{output.hostname}/#{pathname}"
        when 'humanFormat'
          return "#{output.hostname}/#{pathname}.git"
        when 'gitHubDeployKeys'
          return "https://#{output.toString('humanGitHubDeployKeys')}"
        when 'humanGitHubDeployKeys'
          return "github.com/#{pathname}/settings/keys"
        when 'bitbucketDeployKeys'
          return "https://#{output.toString('humanBitbucketDeployKeys')}/"
        when 'humanBitbucketDeployKeys'
          return "bitbucket.org/#{pathname}/admin/deploy-keys"
        when 'dir'
          paths = pathname.split('/')
          name = paths[paths.length - 1]
          return "$HOME/repositories/#{name}"
        when 'dirHTML'
          return output.toString('dir').replace('$HOME', '<span class="color__grey--400">$HOME</span>')
        else
          return output.href
    if format
      output.toString(format)
    else
      output
