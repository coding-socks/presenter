(() => {
  const slidesWrapper = document.querySelector('.slides')
  const slides = slidesWrapper.querySelectorAll('article')
  const helpSnackbar = document.querySelector('#help')

  function goTo(slideIndex) {
    slides.forEach(slide => slide.className = '')
    if (slideIndex >= slides.length) {
      return
    }
    slides[slideIndex].className = 'current'
    const nextSlide = slides[slideIndex + 1]
    if (nextSlide) {
      nextSlide.className = 'next'
      const afterNextSlide = slides[slideIndex + 2]
      if (afterNextSlide) {
        afterNextSlide.className = 'after-next'
      }
    }
    const prevSlide = slides[slideIndex - 1]
    if (prevSlide) {
      prevSlide.className = 'prev'
      const beforePrevSlide = slides[slideIndex - 2]
      if (beforePrevSlide) {
        beforePrevSlide.className = 'before-prev'
      }
    }
  }

  function getSlideIndex() {
    return Number(window.location.hash.substr(1));
  }

  window.addEventListener('hashchange', () => {
    const slide = getSlideIndex()
    goTo(slide)
  }, false)

  const slide = getSlideIndex()
  if (isNaN(slide) || slide < 1) {
    history.replaceState(undefined, undefined, `#0`)
    window.dispatchEvent(new HashChangeEvent('hashchange'))
  } else {
    goTo(slide)
  }

  function scale() {
    const widthUnit = window.innerWidth / 16
    const heightUnit = window.innerHeight / 9
    let ratio
    if (widthUnit < heightUnit) {
      ratio = 1 / (1220 / 16 / widthUnit)
    } else {
      ratio = 1 / (760 / 9 / heightUnit)
    }
    if (ratio > 1) {
      ratio = 1
    }
    slidesWrapper.style.transform = `scale(${ratio})`
  }
  scale()

  let timeOut = null;
  window.addEventListener('resize', () => {
    clearTimeout(timeOut)
    timeOut = setTimeout(() => scale(), 200)
  })

  function hideHelpSnackbar() {
    helpSnackbar.style.display = 'none'
  }

  function moveBackward() {
    const slide = getSlideIndex()
    if (slide > 0) {
      history.replaceState(undefined, undefined, `#${slide - 1}`)
      window.dispatchEvent(new HashChangeEvent('hashchange'))
    }
    hideHelpSnackbar()
  }

  function moveForward() {
    const slide = getSlideIndex()
    if (slide < slides.length - 1) {
      history.replaceState(undefined, undefined, `#${slide + 1}`)
      window.dispatchEvent(new HashChangeEvent('hashchange'))
    }
    hideHelpSnackbar()
  }

  window.addEventListener('keydown', event => {
    if (event.key === "ArrowLeft") {
      moveBackward()
    } else if (event.key === "ArrowRight") {
      moveForward();
    }
  });

  window.addEventListener('keypress', event => {
    if (event.code === 'Space' && event.shiftKey) {
      moveBackward();
    } else if (event.code === 'Space') {
      moveForward();
    } else if (event.code === 'KeyH') {
      hideHelpSnackbar()
    }
  })
})()
