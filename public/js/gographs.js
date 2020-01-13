'use strict';

const defaultInput = 'github.com/linkerd/linkerd2';

function between(val, a, b) {
  if (a < b) {
    return Math.max(a, Math.min(val, b));
  } else {
    return Math.max(b, Math.min(val, a));
  }
}

function checkStatus(response) {
  if (response.status >= 200 && response.status < 300) {
    return response
  } else {
    var error = new Error(response.statusText)
    error.response = response
    throw error
  }
}

function loadSvg(svgHref, goRepo, blob) {
  // createObjectURL() must be coupled with revokeObjectURL(). ownership
  // of svgUrl passes from here to main-svg to thumb-svg.
  const svgUrl = URL.createObjectURL(blob)
  document.getElementById('main-svg').data = svgUrl;

  const externalSvg = document.getElementById('external-svg');
  const externalDot = document.getElementById('external-dot');
  const externalRepo = document.getElementById('external-repo');
  const externalGoDoc = document.getElementById('external-godoc');
  const checkCluster = document.getElementById('check-cluster-input');

  externalSvg.href = svgHref;
  externalSvg.style.display = 'block';

  if (goRepo) {
    externalDot.href = svgHref.replace('.svg', '.dot');
    externalRepo.href = "https://"+goRepo;
    externalGoDoc.href = "https://godoc.org/" + goRepo;

    externalDot.style.display = 'block';
    externalRepo.style.display = 'block';
    externalGoDoc.style.display = 'block';
    checkCluster.style.display = 'block';
  } else {
    externalDot.style.display = 'none';
    externalRepo.style.display = 'none';
    externalGoDoc.style.display = 'none';
    checkCluster.style.display = 'none';
  }
}

// based on:
// https://github.com/ariutta/svg-pan-zoom/blob/d107d73120460caae3ecee59192cd29a470e97b0/demo/thumbnailViewer.js

function updateThumbScope() {
  const thumbToMainZoomRatio = window.thumb.getSizes().realZoom / window.main.getSizes().realZoom;

  let scopeX = window.thumb.getPan().x - window.main.getPan().x * thumbToMainZoomRatio;
  let scopeY = window.thumb.getPan().y - window.main.getPan().y * thumbToMainZoomRatio;
  let scopeWidth = window.main.getSizes().width * thumbToMainZoomRatio;
  let scopeHeight = window.main.getSizes().height * thumbToMainZoomRatio;

  scopeX = Math.max(0, scopeX) + 1;
  scopeY = Math.max(0, scopeY) + 1;

  scopeWidth = Math.min(window.thumb.getSizes().width, scopeWidth);
  scopeHeight = Math.min(window.thumb.getSizes().height, scopeHeight);
  scopeWidth = Math.max(0.1, scopeWidth-2);
  scopeHeight = Math.max(0.1, scopeHeight-2);

  const scope = document.getElementById('scope');
  scope.setAttribute('x', scopeX);
  scope.setAttribute('y', scopeY);
  scope.setAttribute('width', scopeWidth);
  scope.setAttribute('height', scopeHeight);
};

function updateMainZoomPan(e){
  if (e.which === 0 && e.button === 0) {
    return false;
  }

  const dim = document.getElementById('thumb-svg').getBoundingClientRect();
  const scopeDim = document.getElementById('scope').getBoundingClientRect();

  const mainToThumbZoomRatio =  window.main.getSizes().realZoom / window.thumb.getSizes().realZoom;

  const thumbPanX = Math.min(Math.max(0, e.clientX - dim.left), dim.width) - scopeDim.width / 2;
  const thumbPanY = Math.min(Math.max(0, e.clientY - dim.top), dim.height) - scopeDim.height / 2;
  const mainPanX  = - thumbPanX * mainToThumbZoomRatio;
  const mainPanY  = - thumbPanY * mainToThumbZoomRatio;

  window.main.pan({x: mainPanX, y: mainPanY});
}

function resize(_) {
  const scopeContainer = document.getElementById('scope-container');
  const thumbSvg = document.getElementById('thumb-svg');

  scopeContainer.setAttribute('width', thumbSvg.clientWidth);

  window.main.resize();
  window.main.reset();
  window.thumb.resize();
  window.thumb.reset();

  updateThumbScope();
}

function scopeMouseDown(e) {
  captureMouseEvents(e);
  updateMainZoomPan(e);
}

function bindThumbnail(main, thumb){
  if (main) {
    if (window.main) {
      window.main.destroy();
    }
    window.main = main;
  }
  if (thumb) {
    if (window.thumb) {
      window.thumb.destroy();
    }
    window.thumb = thumb;
  }

  if (!window.main || !window.thumb) {
    return;
  }

  // all function below this expect window.main and window.thumb to be set

  window.addEventListener('resize', resize);

  const scopeContainer = document.getElementById('scope-container');
  const thumbSvg = document.getElementById('thumb-svg');

  // set scope-container to match size of thumbnail svg's 'width: auto'
  scopeContainer.setAttribute('width', thumbSvg.clientWidth);

  scopeContainer.addEventListener('mousedown', scopeMouseDown);

  window.main.setOnZoom(function(_){
    updateThumbScope();
  });

  window.main.setOnPan(function(_){
    updateThumbScope();
  });

  updateThumbScope();
}

document.getElementById('scope-container').addEventListener('load', function(){
  this.addEventListener(
    'wheel',
    function wheelZoom(e) {
      const mainSvg = document.getElementById('main-svg');
      const thumbSvg = document.getElementById('thumb-svg');

      Object.defineProperties(e, {
        clientX: { value: mainSvg.offsetWidth * e.offsetX / thumbSvg.offsetWidth },
        clientY: { value: mainSvg.offsetHeight * e.offsetY / thumbSvg.offsetHeight },
      });

      mainSvg.contentDocument.querySelector('svg').dispatchEvent(
        new WheelEvent(e.type, e)
      );

      e.preventDefault();
      return false;
    },
    { passive: false }
  );
});

document.getElementById('main-svg').addEventListener('load', function(){
  const mainSvg = this.contentDocument.querySelector('svg');
  if (!mainSvg) {
    console.debug('no svg loaded');
    return
  }

  // This passes ownership of the objectURL to thumb-svg, which will be
  // responsible for calling revokeObjectURL().
  document.getElementById('thumb-svg').data = this.data;

  mainSvg.addEventListener(
    'wheel',
    function wheelZoom(e) {e.preventDefault()},
    { passive: false }
  );

  const beforePan = function(_, newPan){
    let sizes = this.getSizes();

    const realWidth = sizes.viewBox.width * sizes.realZoom;
    const realHeight = sizes.viewBox.height * sizes.realZoom;

    return {
      x: between(newPan.x, 0, sizes.width - realWidth),
      y: between(newPan.y, 0, sizes.height - realHeight),
    };
  }

  const main = svgPanZoom(mainSvg, {
    viewportSelector: '#main-svg',
    panEnabled: true,
    controlIconsEnabled: false,
    zoomEnabled: true,
    dblClickZoomEnabled: true,
    mouseWheelZoomEnabled: true,
    preventMouseEventsDefault: false,
    zoomScaleSensitivity: 0.2,
    minZoom: 1,
    maxZoom: 20,
    fit: true,
    contain: false,
    center: true,
    refreshRate: 'auto',
    beforeZoom: null,
    onZoom: null,
    beforePan: beforePan,
    onPan: null,
    customEventsHandler: null,
    eventsListenerElement: null,
    onUpdatedCTM: null,
  });

  bindThumbnail(main, undefined);
})

document.getElementById('thumb-svg').addEventListener('load', function(){
  setTimeout(function() {
    // Delay revoking the temporary blob URL. This should not be necessary, but
    // is a workaround to ensure main-svg and thumb-svg successfully load the
    // URL. May be related to:
    // https://stackoverflow.com/a/30708820/7868488
    // https://bugs.chromium.org/p/chromium/issues/detail?id=827932
    URL.revokeObjectURL(this.data);
  }, 1000);

  const thumbSvg = this.contentDocument.querySelector('svg');
  if (!thumbSvg) {
    console.debug('no svg loaded');
    return
  }

  thumbSvg.addEventListener(
    'wheel',
    function wheelZoom(e) {e.preventDefault()},
    { passive: false }
  );

  const thumb = svgPanZoom(thumbSvg, {
    panEnabled: false,
    zoomEnabled: false,
    controlIconsEnabled: false,
    dblClickZoomEnabled: false,
    preventMouseEventsDefault: true,
  });

  bindThumbnail(undefined, thumb);
});

/*
 * window.onload
 */

window.addEventListener('load', (_) => {
  const input = document.getElementById('main-input');

  input.addEventListener('keyup', function(event) {
    if (event.keyCode !== 13) {
      return
    }

    const goRepo = !(this.value.startsWith('http://') || this.value.startsWith('https://')) ?
      this.value :
      null;

    let url;
    const cluster = document.getElementById('check-cluster').checked;
    if (goRepo) {
      url = new URL('/repo/' + this.value + '.svg?cluster=' + cluster, window.location.origin);
    } else {
      url = new URL(this.value);
      if (!url.pathname.endsWith('.svg')) {
        showError('Input URL not an SVG: ' + this.value);
        return
      }
    }

    hideError();

    const spinnerStart = startSpinner();

    fetch(url)
    .then(checkStatus)
    .then(resp => resp.blob())
    .then(blob => {
      let urlState = goRepo ? '/?repo='+this.value+'&cluster='+cluster : '/?url='+url;
      if (this.value === defaultInput && goRepo && !cluster) {
        // special root URL for default inputs
        urlState = '/';
      }

      history.pushState(
        { svgHref: url.href, goRepo: goRepo, blob: blob },
        this.value,
        urlState,
      );

      loadSvg(url.href, goRepo, blob);

      stopSpinner(spinnerStart);
    })
    .catch(error => {
      error.response.text().then(text => {
        showError(text);
        stopSpinner(spinnerStart);
      });
    });
  });

  document.getElementById('check-cluster').addEventListener('change', function(_) {
    input.dispatchEvent(new KeyboardEvent('keyup', { keyCode: 13}));
  });

  updateInputsFromUrl();
  input.dispatchEvent(new KeyboardEvent('keyup', { keyCode: 13}));

  initAutoComplete();
});

window.onpopstate = function(event) {
  updateInputsFromUrl();

  const input = document.getElementById('main-input');
  if (input.value !== '' && event.state && event.state.blob) {
    loadSvg(event.state.svgHref, event.state.goRepo, event.state.blob);
  } else {
    input.dispatchEvent(new KeyboardEvent('keyup', { keyCode: 13}));
  }
}

function updateInputsFromUrl() {
  const input = document.getElementById('main-input');

  const searchParams = new URLSearchParams(window.location.search);
  if (searchParams.has('repo')) {
    // /?repo=github.com/siggy/gographs&cluster=false
    input.value = searchParams.get('repo');
    document.getElementById('check-cluster').checked = searchParams.get('cluster') === 'true';
  } else if (searchParams.has('url')) {
    // /?url=http://gographs.io/repo/github.com/siggy/gographs.svg?cluster=false
    input.value = searchParams.get('url');
  } else {
    // unrecognized URL, reset everything to default
    input.value = defaultInput;
    document.getElementById('check-cluster').checked = false;
  }
}

/*
 * scope mouse capture
 *
 * based on:
 * http://code.fitness/post/2016/06/capture-mouse-events.html
 */

const EventListenerMode = {capture: true};

function preventGlobalMouseEvents() {
  document.body.style.pointerEvents = 'none';
}

function restoreGlobalMouseEvents() {
  document.body.style.pointerEvents = 'auto';
}

function mousemoveListener(e) {
  e.stopPropagation();
  updateMainZoomPan(e);
}

function mouseupListener(e) {
  restoreGlobalMouseEvents();
  document.removeEventListener('mouseup',   mouseupListener,   EventListenerMode);
  document.removeEventListener('mousemove', mousemoveListener, EventListenerMode);
  e.stopPropagation();
}

function captureMouseEvents(e) {
  preventGlobalMouseEvents();
  document.addEventListener('mouseup',   mouseupListener,   EventListenerMode);
  document.addEventListener('mousemove', mousemoveListener, EventListenerMode);
  e.preventDefault();
  e.stopPropagation();
}

/*
 * autocomplete
 */

function initAutoComplete() {
  const input = document.getElementById('main-input');

  fetch('/top-repos')
  .then(checkStatus)
  .then(resp => resp.json())
  .then(json => {

    new autoComplete({
      selector: '#main-input',
      minChars: 0,
      cache: false,
      source: function(term, suggest){
        term = term.toLowerCase();
        const choices = json.slice(0,10);
        const suggestions = [];
        for (let i=0; i<choices.length; i++) {
          if (~choices[i].toLowerCase().indexOf(term)) {
            suggestions.push(choices[i]);
          }
        }
        suggest(suggestions);
      },
      onSelect: function(e, term, item){
        if (e instanceof KeyboardEvent && e.keyCode === 13) {
          // the input element also handles keyboard enter, skip this one.
          return
        }
        input.dispatchEvent(new KeyboardEvent('keyup', { keyCode: 13}));
        e.preventDefault();
      },
    });
  })
  .catch(error => {
    console.error(error);
  });
}

/*
 * spinner
 */

function startSpinner() {
  return setTimeout(function() {
    document.getElementById('spinner').style.display = 'flex';
  }, 250);
}

function stopSpinner(timeout) {
  clearTimeout(timeout);
  document.getElementById('spinner').style.display = 'none';
}

/*
 * error messages
 */

function showError(message) {
  const elm = document.getElementById('input-error');
  elm.innerHTML = message;
  elm.classList.add("visible");
  setTimeout(hideError, 5000);
}

function hideError() {
  const elm = document.getElementById('input-error');
  elm.classList.remove("visible");
}
