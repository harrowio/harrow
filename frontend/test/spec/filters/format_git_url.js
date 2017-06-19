describe('Filter: formatGitUrl', function () {
  var filter
  beforeEach(angular.mock.inject(function ($filter) {
    filter = $filter('formatGitUrl')
  }))

  it('handles null input', function () {
    expect(filter()).toEqual(undefined)
  })

  it('it handles https urls', function () {
    expect(filter('https://github.com/sshkit/sshkit.git')).toEqual('github.com/sshkit/sshkit.git')
  })

  it('it handles git urls', function () {
    expect(filter('git://user@server/project.git')).toEqual('server/project.git')
  })

  it('it handles ssh urls', function () {
    expect(filter('ssh://user@server/project.git')).toEqual('server/project.git')
  })

  // Apparently we don't handle URLs with : in them with this library
  // which is unfortunate.
  it('it handles scp urls', function () {
    expect(filter('user@server:project.git')).toEqual('server/project.git')
  })

  it('it handles URLs with an undefined scheme', function () {
    expect(filter('github.com/sshkit/sshkit.git')).toEqual('github.com/sshkit/sshkit.git')
  })
})
