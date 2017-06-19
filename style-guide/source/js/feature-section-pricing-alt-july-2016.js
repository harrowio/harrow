;(function (window) {
  var priceToggler = document.querySelector('.featureSection__pricing--alt__toggle__button')
  var labels = document.querySelectorAll('.featureSection__pricing--alt__toggle__item')
  var pricingOptions = document.querySelectorAll('.featureSection__pricing--alt__block__item__price')
  var callToAction = document.querySelector('.featureSection__pricing--alt__footer .btn--primary')

  function setPricing (multi, discountPercent) {
    if (!multi) {
      multi = 1
    }
    if (!discountPercent) {
      discountPercent = 0
    }
    var price = 0
    var newPrice = 0
    var saving = 0
    var output = ''
    for (var j = 0; j < pricingOptions.length; j++) {
      price = parseInt(pricingOptions[j].getAttribute('data-core-price'), 10)
      if (price) {
        price = price * multi
        newPrice = Math.round((price - ((price / 100) * discountPercent)))
        saving = Math.round(price - newPrice)
        output = '$' + newPrice
        if (saving) {
          // output += '<br><small>(saving $' + saving + ')</small>'
        }
        if (discountPercent > 0) {
          pricingOptions[j].classList.add('discounted')
        } else {
          pricingOptions[j].classList.remove('discounted')
        }
        pricingOptions[j].innerHTML = output
      }
    }
  }

  function toggle () {
    var isOn = priceToggler.classList.contains('on')
    if (isOn) {
      labels[1].classList.remove('active')
      labels[0].classList.add('active')
      priceToggler.classList.remove('on')
      setPricing()
    } else {
      labels[0].classList.remove('active')
      labels[1].classList.add('active')
      priceToggler.classList.add('on')
      setPricing(12, 30)
    }
    jello()
  }

  function jello () {
    callToAction.classList.add('jello')
    setTimeout(function () {
      callToAction.classList.remove('jello')
    }, 750)
  }

  if (priceToggler) {
    priceToggler.addEventListener('click', toggle)
  }

  var matches = null
  for (var j = 0; j < pricingOptions.length; j++) {
    matches = pricingOptions[j].textContent.match(/(\d+)/)
    if (matches) {
      pricingOptions[j].setAttribute('data-core-price', parseInt(matches[1], 10))
    }
  }

  if (callToAction) {
    setInterval(jello, 10000)
  }
})(window)
