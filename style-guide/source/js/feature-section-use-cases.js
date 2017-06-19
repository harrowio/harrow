;(function (window) {
  var el, elements
  el = document.querySelector('.use-cases-diagram')
  elements = document.querySelectorAll('.featureSection__useCases__case')

  function clearConnectionLines () {
    highlightGroup('Environments')
    highlightGroup('Repos')
    highlightGroup('Tasks')
    var connections = el.querySelectorAll('g#Connecting-lines > g')
    for (var i = 0; i < connections.length; i++) {
      connections[i].style.opacity = 0
    }
  }

  function highlightGroup (key, id, arr) {
    var items, item, dataAttr
    items = el.querySelectorAll('g#' + key + ' > g')
    item = null
    dataAttr = null
    for (var i = 0; i < items.length; i++) {
      items[i].style.opacity = 0.2
    }

    for (var j = 0; j < elements.length; j++) {
      elements[j].classList.remove('active')
    }

    if (!id) {
      return
    }
    for (var g = 0; g < elements.length; g++) {
      dataAttr = elements[g].getAttribute('data-use-case-example')
      if (dataAttr === id) {
        elements[g].classList.add('active')
      }
    }

    arr.forEach(function (t) {
      if (id.toLowerCase().indexOf(t.toLowerCase()) > 0) {
        item = el.querySelector('g#' + key + ' > g#' + t)
        if (item) {
          item.style.opacity = 1
        }
      }
    })
  }

  function onUseCaseClick (event) {
    clearConnectionLines()
    var useCaseExample = event.target.getAttribute('data-use-case-example')
    highlightGroup('Environments', useCaseExample, ['production', 'staging'])
    highlightGroup('Repos', useCaseExample, ['frontend', 'backend'])
    highlightGroup('Tasks', useCaseExample, ['lint', 'unitTests', 'acceptenceTests', 'deploy'])
    var useCase = el.querySelector(useCaseExample)
    if (useCase) {
      useCase.style.opacity = 1
    }
  }
  var isAuto = false
  function startCycle () {
    if (isAuto) {
      return
    }
    isAuto = true
    setInterval(function () {
      var activeEl = document.querySelector('.featureSection__useCases__case.active')
      if (activeEl.nextElementSibling) {
        activeEl.nextElementSibling.click()
      } else {
        elements[0].click()
      }
    }, 6000)
  }

  if (elements.length) {
    for (var i = 0; i < elements.length; i++) {
      elements[i].addEventListener('click', onUseCaseClick)
    }
    elements[0].click()

    var carouselTop = el.getBoundingClientRect().top
    window.addEventListener('scroll', function () {
      if ((document.body.scrollTop + window.innerHeight) >= carouselTop) {
        startCycle()
      }
    })
  }
})(window)
