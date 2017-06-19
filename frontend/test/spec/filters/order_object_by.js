describe('Filter: orderObjectBy', function () {
  var api
  beforeEach(angular.mock.inject(function ($filter) {
    api = $filter('orderObjectBy')
  }))

  it('orders nested object', function () {
    var actual = api({
      abc: {
        subject: {
          name: 'Zoolu'
        }
      },
      bcd: {
        subject: {
          name: 'Alpha'
        }
      }
    }, 'subject.name')

    expect(actual).toEqual([
      {
        subject: {
          name: 'Alpha'
        }
      },
      {
        subject: {
          name: 'Zoolu'
        }
      }
    ])
  })
})
