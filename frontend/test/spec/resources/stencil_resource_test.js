/*eslint-disable new-cap*/
describe('Resource: stencilResource', function () {
  var api, model, data, $httpBackend
  beforeEach(angular.mock.inject(function (stencilResource, _$httpBackend_) {
    $httpBackend = _$httpBackend_
    data = {
      subject: {}
    }
    api = stencilResource
    model = new api.model(data)
  }))

  describe('.save()', function () {
    it('POST to "/api/stencils"', function () {
      $httpBackend.expect('POST', /\/api\/stencils/).respond(201)
      api.save(data)
      $httpBackend.flush()
    })
  })
})
