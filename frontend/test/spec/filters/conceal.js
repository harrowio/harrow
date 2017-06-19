describe('Filter: conceal', function () {
  var api
  beforeEach(angular.mock.inject(function ($filter) {
    api = $filter('conceal')
  }))

  it('converts the input text to "BLACK CIRCLE"', function () {
    expect(api('hello world')).toEqual('●●●●●●●●●●●')
  })

  it('converts the input text to "BLACK CIRCLE" and limits length to 4', function () {
    expect(api('hello world', 4)).toEqual('●●●●')
  })
})
