/*eslint-disable new-cap*/
describe('Resources: userResource', function () {
  var api, model, data, $httpBackend
  beforeEach(angular.mock.inject(function (userResource, _$httpBackend_) {
    $httpBackend = _$httpBackend_
    data = jasmine.getJSONFixture('GET_api_user.json')
    api = userResource
    model = new api.model(data)
  }))

  describe('.model() - linked resources', function () {
    it('defines "tasks()"', function () {
      var taskData = jasmine.getJSONFixture('empty_collection.json')
      $httpBackend.expect('GET', /\/api\/users\/[0-9a-f-]*\/jobs/).respond(200, taskData)
      expect(model.tasks).toBeDefined('has tasks')
      model.tasks().then(function (response) {
        expect(response).toEqual(taskData.collection)
      })
      $httpBackend.flush()
    })
  })
})
