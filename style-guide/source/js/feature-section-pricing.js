;(function (window) {
  var pricingToggles = document.querySelectorAll('.featureSection__pricing__switch .btn')
  var pricingOptions = document.querySelectorAll('.featureSection__pricing__card__header')

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
        if(saving) {
          output += '<br><small>(saving $'+ saving +')</small>'
        }
        pricingOptions[j].innerHTML = output
      }
    }
  }

  function onPricingToggle (event) {
    event.preventDefault()
    for (var i = 0; i < pricingToggles.length; i++) {
      pricingToggles[i].classList.remove('active')
    }
    this.classList.add('active')
    if (this.textContent.indexOf('Year') >= 0) {
      setPricing(12, 20)
    } else if (this.textContent.indexOf('Month') >= 0) {
      setPricing()
    }
  }
  for (var i = 0; i < pricingToggles.length; i++) {
    pricingToggles[i].addEventListener('click', onPricingToggle)
  }
  var matches = null
  for (var j = 0; j < pricingOptions.length; j++) {
    matches = pricingOptions[j].textContent.match(/(\d+)/)
    if (matches) {
      pricingOptions[j].setAttribute('data-core-price', parseInt(matches[1], 10))
    }
  }

})(window)
