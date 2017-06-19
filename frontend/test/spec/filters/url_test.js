describe('Filter:', function () {
  var filter, isSsh, url, protocols
  beforeEach(angular.mock.inject(function ($filter) {
    url = $filter('url')
    isSsh = $filter('isSsh')
    protocols = $filter('protocols')
  }))

  describe('protocols', function () {
    it('returns ', function () {
      expect(protocols('ssh://git@github.com/harrow/harrow.git')).toContain('ssh')
      expect(protocols('git+ssh://git@github.com/harrow/harrow.git')).toContain('git')
      expect(protocols('git+ssh://git@github.com/harrow/harrow.git')).toContain('ssh')
      expect(protocols('https://github.com/harrow/harrow')).toContain('https')
    })
  })

  describe('isSsh', function () {
    it('retruns true', function () {
      expect(isSsh('ssh://git@github.com/harrow/harrow.git')).toBeTruthy()
      expect(isSsh('git+ssh://git@github.com/harrow/harrow.git')).toBeTruthy()
      expect(isSsh('https://github.com/harrow/harrow')).toBeFalsy()
    })
  })

  describe('url', function () {
    it('handles null input', function () {
      expect(url()).toEqual(undefined)
    })
    it('parses standrd url', function () {
      var actual = url('https://github.com/harrow/harrow')
      expect(actual.protocol).toEqual('https')
      expect(actual.hostname).toEqual('github.com')
      expect(actual.pathname).toEqual('/harrow/harrow')
      expect(actual.toString('ssh')).toEqual('git@github.com:harrow/harrow.git')
      expect(actual.toString('https')).toEqual('https://github.com/harrow/harrow')
      expect(actual.toString('git+ssh')).toEqual('git+ssh://git@github.com/harrow/harrow.git')
      expect(actual.toString()).toEqual('https://github.com/harrow/harrow')
    })

    it('parses ssh', function () {
      var actual = url('git@github.com:harrow/harrow.git')
      expect(actual.protocol).toEqual('ssh')
      expect(actual.hostname).toEqual('github.com')
      expect(actual.pathname).toEqual('/harrow/harrow.git')
      expect(actual.toString('ssh')).toEqual('git@github.com:harrow/harrow.git')
      expect(actual.toString('https')).toEqual('https://github.com/harrow/harrow')
      expect(actual.toString('git+ssh')).toEqual('git+ssh://git@github.com/harrow/harrow.git')
      expect(actual.toString()).toEqual('git@github.com:harrow/harrow.git')
    })

    it('parses git+ssh', function () {
      var actual = url('git+ssh://git@github.com/harrow/harrow.git')
      expect(actual.protocol).toEqual('git')
      expect(actual.hostname).toEqual('github.com')
      expect(actual.pathname).toEqual('/harrow/harrow.git')
      expect(actual.toString('ssh')).toEqual('git@github.com:harrow/harrow.git')
      expect(actual.toString('https')).toEqual('https://github.com/harrow/harrow')
      expect(actual.toString('git+ssh')).toEqual('git+ssh://git@github.com/harrow/harrow.git')
      expect(actual.toString()).toEqual('git+ssh://git@github.com/harrow/harrow.git')
    })

    it('parses git+ssh', function () {
      var actual = url('ssh://git@github.com/harrow/harrow.git')
      expect(actual.protocol).toEqual('ssh')
      expect(actual.hostname).toEqual('github.com')
      expect(actual.pathname).toEqual('/harrow/harrow.git')
      expect(actual.toString('ssh')).toEqual('git@github.com:harrow/harrow.git')
      expect(actual.toString('https')).toEqual('https://github.com/harrow/harrow')
      expect(actual.toString('git+ssh')).toEqual('git+ssh://git@github.com/harrow/harrow.git')
      expect(actual.toString()).toEqual('ssh://git@github.com/harrow/harrow.git')
    })
  })
})
