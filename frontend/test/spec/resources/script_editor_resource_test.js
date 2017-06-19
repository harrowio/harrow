describe('Resource: scriptEditorResource', function () {
  var $httpBackend, api

  beforeEach(angular.mock.inject(function (_$httpBackend_, scriptEditorResource) {
    $httpBackend = _$httpBackend_
    api = scriptEditorResource
  }))

  describe('.apply', function () {
    it('POST to "/api/script-editor/apply"', function () {
      $httpBackend.expect(
        'POST',
        /\/api\/script-editor\/apply/,
        jasmine.validateHttpParams({projectUuid: 'proj123', jobUuid: 'task123'})
      ).respond(201)

      api.apply({
        projectUuid: 'proj123',
        taskUuid: 'task123'
      })
      $httpBackend.flush()
    })
  })

  describe('.diff', function () {
    it('POST to "/api/script-editor/diff"', function () {
      $httpBackend.expect(
        'POST',
        /\/api\/script-editor\/diff/,
        jasmine.validateHttpParams({projectUuid: 'proj123', jobUuid: 'task123'})
      ).respond(201)

      api.diff({
        projectUuid: 'proj123',
        taskUuid: 'task123'
      })
      $httpBackend.flush()
    })
  })

  describe('.save', function () {
    it('POST to "/api/script-editor/save"', function () {
      $httpBackend.expect(
        'POST',
        /\/api\/script-editor\/save/,
        jasmine.validateHttpParams({projectUuid: 'proj123', jobUuid: 'task123'})
      ).respond(201)

      api.save({
        projectUuid: 'proj123',
        taskUuid: 'task123'
      })
      $httpBackend.flush()
    })
  })
})
