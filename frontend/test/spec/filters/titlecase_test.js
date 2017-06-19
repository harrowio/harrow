describe('Filter: titlecase', function () {
  var filter
  beforeEach(angular.mock.inject(function ($filter) {
    filter = $filter('titlecase')
  }))

  it('handles null input ', function () {
    expect(filter()).toBeUndefined()
  })

  it('uppercases first letter', function () {
    expect(filter('hello world')).toEqual('Hello World')
  })

  it('excludes little words', function () {
    expect(filter('this and that')).toEqual('This and That')
  })

  it('excludes little words', function () {
    expect(filter('testing is cool')).toEqual('Testing is Cool')
  })

  it('excludes little words', function () {
    expect(filter('testing if needed')).toEqual('Testing if Needed')
  })
})
