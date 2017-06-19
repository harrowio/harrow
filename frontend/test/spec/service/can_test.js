describe('Service: can', function () {
  var api
  beforeEach(angular.mock.inject(function (can) {
    api = can
  }))
  describe('.can()', function () {
    it('returns true when the resource includes a method', function () {
      var project = {
        '_links': {
          repositories: {
            update: 'PUT'
          }
        }
      }
      expect(api.can('update-repositories', project)).toBeTruthy()
    })

    it('returns false when resource excludes method', function () {
      var project = {
        '_links': {
          repositories: {}
        }
      }
      expect(api.can('update-repositories', project)).toBeFalsy()
    })

    it('returns false otherwise', function () {
      expect(api.can('update-repositories', {})).toBeFalsy()
    })
  })
})
