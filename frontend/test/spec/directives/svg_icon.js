describe('Directive: svgIcon', function () {
  var $compile, $rootScope, $httpBackend
  beforeEach(angular.mock.inject(function (_$rootScope_, _$compile_, _$httpBackend_) {
    $rootScope = _$rootScope_
    $compile = _$compile_
    $httpBackend = _$httpBackend_
    $httpBackend.expect('GET', /icons\.svg/).respond(200, `<?xml version="1.0" encoding="UTF-8"?>
  <!DOCTYPE svg PUBLIC "-//W3C//DTD SVG 1.1//EN" "http://www.w3.org/Graphics/SVG/1.1/DTD/svg11.dtd">
  <svg xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink">
    <symbol id="icon-complete" viewBox="0 0 36 36">
      <title>complete</title>
      <path d="M18 36c9.94 0 18-8.06 18-18S27.94 0 18 0 0 8.06 0 18s8.06 18 18 18zm7.378-24.478l2.758 2.314-10.8 12.87-2.757-2.314 10.798-12.87zm-14 5.484l5.515 4.628-2.314 2.758-5.516-4.628 2.314-2.758z" fill-rule="evenodd"/>
    </symbol>
  </svg>`)
  }))

  it('expands attribute to html content', function () {
    var el = $compile('<span svg-icon="icon-complete"/>')($rootScope)
    $httpBackend.flush()
    expect(el[0].tagName).toEqual('svg')
    expect(el[0].getAttribute('width')).toEqual('36')
    expect(el[0].getAttribute('height')).toEqual('36')
    expect(el.find('title').text()).toEqual('complete')
    expect(el.find('path').attr('d')).not.toEqual('')
  })

  it('expands attribute to html content with class', function () {
    var el = $compile('<span svg-icon="example" class="iconColor"/>')($rootScope)
    $httpBackend.flush()
    expect(el[0].tagName).toEqual('svg')
    expect(el[0].classList).toContain('iconColor')
    expect(el[0].getAttribute('width')).toEqual('36')
    expect(el[0].getAttribute('height')).toEqual('36')
  })

  it('expands attribute with custom size', function () {
    var el = $compile('<span svg-icon="example" svg-icon-size="20"/>')($rootScope)
    $httpBackend.flush()
    expect(el[0].tagName).toEqual('svg')
    expect(el[0].getAttribute('width')).toEqual('20')
    expect(el[0].getAttribute('height')).toEqual('20')
  })
})
