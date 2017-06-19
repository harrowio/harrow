;(function (window) {
  var carousel = document.querySelector('.carousel')
  var seats = document.querySelectorAll('.carousel__seat')
  var dotsContainer = document.querySelector('.carousel__dots')
  var dots = []
  var nextButton = document.querySelector('.carousel__button--forwards')
  var prevButton = document.querySelector('.carousel__button--backwards')

  function next (el) {
    if (el.nextElementSibling) {
      return el.nextElementSibling
    } else {
      return seats[0]
    }
  }

  function prev (el) {
    if (el.previousElementSibling) {
      return el.previousElementSibling
    } else {
      return seats[seats.length - 1]
    }
  }

  function switchPlaces (event) {
    event.preventDefault()
    var nextSeat = null
    var direction = this.getAttribute('data-carousel-direction') || 'forward'
    var ref = carousel.querySelector('.carousel__seat--isRef')
    ref.classList.remove('carousel__seat--isRef')
    if (direction === 'backwards') {
      carousel.classList.add('carousel--isReversing')
      nextSeat = prev(ref) // backwards
    } else {
      carousel.classList.remove('carousel--isReversing')
      nextSeat = next(ref) // forward
    }

    nextSeat.classList.add('carousel__seat--isRef')
    nextSeat.style.order = 1

    var dotIndex = Array.prototype.indexOf.call(seats, next(nextSeat))
    for (var i = 0; i < dots.length; i++) {
      dots[i].classList.remove('active')
    }
    dots[dotIndex].classList.add('active')

    for (var j = 2; j < seats.length + 1; j++) {
      nextSeat = next(nextSeat)
      nextSeat.style.order = j
    }
    carousel.classList.remove('carousel--isSet')
    setTimeout(function () {
      carousel.classList.add('carousel--isSet')
    }, 50)
    return false
  }

  function dotClick (event) {
    event.preventDefault()
    var index = Array.prototype.indexOf.call(this.parentElement.children, this)
    var activeIndex = Array.prototype.indexOf.call(this.parentElement.children, document.querySelector('.carousel__dot.active'))
    if (index > activeIndex) {
      for (var i = 0; i < (index - activeIndex); i++) {
        nextButton.click()
      }
    } else if (index < activeIndex) {
      for (var i = 0; i < (activeIndex - index); i++) {
        prevButton.click()
      }
    }
    return false
  }

  var isAuto = false
  function startCarousel () {
    if (isAuto) {
      return
    }
    isAuto = true
    setInterval(function () {
      nextButton.click()
    }, 6000)
  }

  if (seats.length > 0) {
    seats[seats.length - 1].classList.add('carousel__seat--isRef')

    var dot = null
    for (var i = 0; i < seats.length; i++) {
      dot = document.createElement('span')
      dot.classList.add('carousel__dot')
      if (i === 0) {
        dot.classList.add('active')
      }
      dot.addEventListener('click', dotClick, false)
      dots.push(dot)

      dotsContainer.appendChild(dot)
    }

    nextButton.addEventListener('click', switchPlaces, false)
    prevButton.addEventListener('click', switchPlaces, false)

    var carouselTop = carousel.getBoundingClientRect().top
    window.addEventListener('scroll', function () {
      if ((document.body.scrollTop + window.innerHeight) >= carouselTop) {
        startCarousel()
      }
    })

  }
})(window)
