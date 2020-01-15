'use strict';

const defaultInput = 'github.com/linkerd/linkerd2';

const DOM = {
  // https://magnushoff.com/blog/dependency-free-javascript/
  checkCluster:      document.getElementById('check-cluster'),
  checkClusterInput: document.getElementById('check-cluster-input'),
  externalDot:       document.getElementById('external-dot'),
  externalGoDoc:     document.getElementById('external-godoc'),
  externalRepo:      document.getElementById('external-repo'),
  externalSvg:       document.getElementById('external-svg'),
  inputError:        document.getElementById('input-error'),
  mainInput:         document.getElementById('main-input'),
  mainSvg:           document.getElementById('main-svg'),
  scope:             document.getElementById('scope'),
  scopeContainer:    document.getElementById('scope-container'),
  spinner:           document.getElementById('spinner'),
  thumbSvg:          document.getElementById('thumb-svg'),
};

/*
 * window.onload
 */

window.addEventListener('load', (_) => {
  // TODO: AUTO?
  // document.getElementById('control-toggle').checked = true;



  DOM.checkCluster.addEventListener('change', function(_) {
    handleQuery();
  });

  updateInputsFromUrl();
  handleQuery();

  initAutoComplete();
});

window.onpopstate = function(event) {
  updateInputsFromUrl();

  if (DOM.mainInput.value !== '' && event.state && event.state.blob) {
    loadSvg(event.state.svgHref, event.state.goRepo, event.state.blob);
  } else {
    handleQuery();
  }
}

function updateInputsFromUrl() {
  const searchParams = new URLSearchParams(window.location.search);
  if (searchParams.has('repo')) {
    // /?repo=github.com/siggy/gographs&cluster=false
    DOM.mainInput.value = searchParams.get('repo');
    DOM.checkCluster.checked = searchParams.get('cluster') === 'true';
  } else if (searchParams.has('url')) {
    // /?url=https://gographs.io/repo/github.com/siggy/gographs.svg?cluster=false
    DOM.mainInput.value = searchParams.get('url');
  } else {
    // unrecognized URL, reset everything to default
    DOM.mainInput.value = defaultInput;
    DOM.checkCluster.checked = false;
  }
}

/*
 * DOM element EventListeners
 */

DOM.mainInput.addEventListener('keyup', function(event) {
  if (event.keyCode !== 13) {
    return
  }

  handleQuery();
});

DOM.mainSvg.addEventListener('load', function(){
  // This passes ownership of the objectURL to thumb-svg, which will be
  // responsible for calling revokeObjectURL().
  DOM.thumbSvg.data = this.data;

  const svg = this.contentDocument.querySelector('svg');
  if (!svg) {
    console.debug('no svg loaded');
    return
  }

  svg.addEventListener(
    'wheel',
    function wheelZoom(e) {e.preventDefault()},
    { passive: false }
  );

  const mainSvgPanZoom = svgPanZoom(svg, {
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

  bindThumbnail(mainSvgPanZoom, undefined);
})

DOM.thumbSvg.addEventListener('load', function(){
  setTimeout(function() {
    // Delay revoking the temporary blob URL. This should not be necessary, but
    // is a workaround to ensure main-svg and thumb-svg successfully load the
    // URL. May be related to:
    // https://stackoverflow.com/a/30708820/7868488
    // https://bugs.chromium.org/p/chromium/issues/detail?id=827932
    URL.revokeObjectURL(this.data);
  }, 1000);

  const svg = this.contentDocument.querySelector('svg');
  if (!svg) {
    console.debug('no svg loaded');
    return
  }

  svg.addEventListener(
    'wheel',
    function wheelZoom(e) {e.preventDefault()},
    { passive: false }
  );

  const thumbSvgPanZoom = svgPanZoom(svg, {
    panEnabled: false,
    zoomEnabled: false,
    dblClickZoomEnabled: false,
  });

  bindThumbnail(undefined, thumbSvgPanZoom);
});

DOM.scopeContainer.addEventListener(
  'wheel',
  function wheelZoom(e) {
    const svg = DOM.mainSvg.contentDocument.querySelector('svg')
    if (!svg) {
      console.debug('no svg loaded');
      return
    }

    Object.defineProperties(e, {
      clientX: { value: DOM.mainSvg.offsetWidth * e.offsetX / DOM.thumbSvg.offsetWidth },
      clientY: { value: DOM.mainSvg.offsetHeight * e.offsetY / DOM.thumbSvg.offsetHeight },
    });

    // forward wheel zooming from scopeContainer to svg
    svg.dispatchEvent(
      new WheelEvent(e.type, e)
    );

    e.preventDefault();
    return false;
  },
  { passive: false }
);

/*
 * Misc functions for updating state
 */

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

  // set scope-container to match size of thumbnail svg's 'width: auto'
  DOM.scopeContainer.setAttribute('width', DOM.thumbSvg.clientWidth);

  DOM.scopeContainer.addEventListener('mousedown', scopeMouseDown);

  window.main.setOnZoom(function(_){
    updateThumbScope();
  });

  window.main.setOnPan(function(_){
    updateThumbScope();
  });

  updateThumbScope();
}

function between(val, a, b) {
  if (a < b) {
    return Math.max(a, Math.min(val, b));
  } else {
    return Math.max(b, Math.min(val, a));
  }
}

function beforePan(_, newPan){
  let sizes = this.getSizes();

  const realWidth = sizes.viewBox.width * sizes.realZoom;
  const realHeight = sizes.viewBox.height * sizes.realZoom;

  return {
    x: between(newPan.x, 0, sizes.width - realWidth),
    y: between(newPan.y, 0, sizes.height - realHeight),
  };
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

function handleQuery() {
  if (DOM.mainInput.value === "") {
    DOM.mainInput.value = defaultInput;
  }

  const goRepo = !(DOM.mainInput.value.startsWith('http://') || DOM.mainInput.value.startsWith('https://')) ?
  DOM.mainInput.value :
    null;

  let url;
  const cluster = DOM.checkCluster.checked;
  if (goRepo) {
    url = new URL('/repo/' + DOM.mainInput.value + '.svg?cluster=' + cluster, window.location.origin);
  } else {
    url = new URL(DOM.mainInput.value);
    if (!url.pathname.endsWith('.svg')) {
      showError('Input URL not an SVG: ' + DOM.mainInput.value);
      return
    }
  }

  hideError();

  const spinner = startSpinner();

  fetch(url)
  .then(checkStatus)
  .then(resp => resp.blob())
  .then(blob => {
    let urlState = goRepo ? '/?repo='+DOM.mainInput.value+'&cluster='+cluster : '/?url='+url;
    if (DOM.mainInput.value === defaultInput && !cluster) {
      // special root URL for default inputs
      urlState = '/';
    }

    history.pushState(
      { svgHref: url.href, goRepo: goRepo, blob: blob },
      DOM.mainInput.value,
      urlState,
    );

    loadSvg(url.href, goRepo, blob);

    stopSpinner(spinner);
  })
  .catch(error => {
    error.response.text().then(text => {
      showError(text);
      stopSpinner(spinner);
    });
  });
}

function loadSvg(svgHref, goRepo, blob) {
  // createObjectURL() must be coupled with revokeObjectURL(). ownership
  // of svgUrl passes from here to main-svg to thumb-svg.
  const svgUrl = URL.createObjectURL(blob)
  DOM.mainSvg.data = svgUrl;

  DOM.externalSvg.href = svgHref;
  DOM.externalSvg.style.display = 'block';

  if (goRepo) {
    DOM.externalDot.href = svgHref.replace('.svg', '.dot');
    DOM.externalRepo.href = "https://"+goRepo;
    DOM.externalGoDoc.href = "https://godoc.org/" + goRepo;

    DOM.externalDot.style.display = 'block';
    DOM.externalRepo.style.display = 'block';
    DOM.externalGoDoc.style.display = 'block';
    DOM.checkClusterInput.style.display = 'block';
  } else {
    DOM.externalDot.style.display = 'none';
    DOM.externalRepo.style.display = 'none';
    DOM.externalGoDoc.style.display = 'none';
    DOM.checkClusterInput.style.display = 'none';
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

  DOM.scope.setAttribute('x', scopeX);
  DOM.scope.setAttribute('y', scopeY);
  DOM.scope.setAttribute('width', scopeWidth);
  DOM.scope.setAttribute('height', scopeHeight);
};

function updateMainZoomPan(e){
  if (e.which === 0 && e.button === 0) {
    return false;
  }

  const dim = DOM.thumbSvg.getBoundingClientRect();
  const scopeDim = DOM.scope.getBoundingClientRect();

  const mainToThumbZoomRatio =  window.main.getSizes().realZoom / window.thumb.getSizes().realZoom;

  const thumbPanX = Math.min(Math.max(0, e.clientX - dim.left), dim.width) - scopeDim.width / 2;
  const thumbPanY = Math.min(Math.max(0, e.clientY - dim.top), dim.height) - scopeDim.height / 2;
  const mainPanX  = - thumbPanX * mainToThumbZoomRatio;
  const mainPanY  = - thumbPanY * mainToThumbZoomRatio;

  window.main.pan({x: mainPanX, y: mainPanY});
}

function resize(_) {
  DOM.scopeContainer.setAttribute('width', DOM.thumbSvg.clientWidth);

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
        handleQuery();
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
    DOM.spinner.style.display = 'flex';
  }, 250);
}

function stopSpinner(timeout) {
  clearTimeout(timeout);
  DOM.spinner.style.display = 'none';
}

/*
 * error messages
 */

function showError(message) {
  DOM.inputError.innerHTML = message;
  DOM.inputError.classList.add("visible");
  setTimeout(hideError, 5000);
}

function hideError() {
  DOM.inputError.classList.remove("visible");
}
