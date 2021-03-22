'use strict';

// TODO: needed?
const defaultInput = 'github.com/siggy/gographs';

const DOM = {
  // https://magnushoff.com/blog/dependency-free-javascript/
  checkCluster:      document.getElementById('check-cluster'),
  externalDot:       document.getElementById('external-dot'),
  externalGoDoc:     document.getElementById('external-godoc'),
  externalRepo:      document.getElementById('external-repo'),
  externalSvg:       document.getElementById('external-svg'),
  inputError:        document.getElementById('input-error'),
  mainInput:         document.getElementById('main-input'),
  mainSvg:           document.getElementById('main-svg'),
  badge:             document.getElementById('badge'),
  badgeMarkdown:     document.getElementById('badge-markdown'),
  badgeText:         document.getElementById('badge-text'),
  refreshButton:     document.getElementById('refresh'),
  scope:             document.getElementById('scope'),
  scopeContainer:    document.getElementById('scope-container'),
  spinner:           document.getElementById('spinner'),
  thumbSvg:          document.getElementById('thumb-svg'),
  viewport:          null, // fill in on load with "".svg-pan-zoom_viewport"
  autocomplete:      null, // fill in on handleInput
};

/*
 * window.onload
 */

window.addEventListener('load', (_) => {
  DOM.checkCluster.addEventListener('change', function(_) {
    handleInput(false);
  });

  updateInputsFromUrl();
  handleInput(false);

  initAutoComplete();
});

window.onpopstate = function(event) {
  updateInputsFromUrl();

  if (DOM.mainInput.value !== '' && event.state && event.state.blob) {
    loadSvg(event.state.svgHref, event.state.goRepo, event.state.blob);
  } else {
    handleInput(false);
  }
}

function updateInputsFromUrl() {
  const searchParams = new URLSearchParams(window.location.search);

  if (window.location.pathname.startsWith("/repo/")) {
    // /repo/github.com/siggy/gographs?cluster=false
    DOM.mainInput.value = window.location.pathname.slice("/repo/".length)
    DOM.checkCluster.checked = searchParams.get('cluster') === 'true';
  } else if (window.location.pathname.startsWith("/svg")) {
    // /svg?url=https://gographs.io/repo/github.com/siggy/gographs.svg?cluster=false
    DOM.mainInput.value = searchParams.get('url');
  } else {
    // unrecognized URL, reset everything to default
    DOM.mainInput.value = defaultInput;
    DOM.checkCluster.checked = true;
  }

  return;
}

/*
 * DOM element EventListeners
 */

DOM.mainInput.addEventListener('keyup', function(event) {
  if (event.keyCode !== 13) {
    return
  }

  handleInput(false);
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
    controlIconsEnabled: true,
    zoomEnabled: true,
    dblClickZoomEnabled: true,
    mouseWheelZoomEnabled: true,
    preventMouseEventsDefault: false,
    zoomScaleSensitivity: 0.2,
    minZoom: 1,
    maxZoom: 100,
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

  DOM.viewport = DOM.mainSvg.contentWindow.document.
    getElementsByClassName("svg-pan-zoom_viewport")[0];
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
    const svg = DOM.mainSvg.contentDocument.querySelector('svg');
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

DOM.refreshButton.addEventListener(
  'click',
  function() {
    handleInput(true);
    return false;
  }
);

DOM.badge.addEventListener(
  'click',
  function() {
    DOM.badgeMarkdown.classList.toggle("visible");
    DOM.badgeText.focus();
    DOM.badgeText.select();
    return false;
  }
);

DOM.badgeMarkdown.addEventListener(
  'click',
  function() {
    DOM.badgeText.focus();
    DOM.badgeText.select();
    return false;
  }
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

function beforePan(oldPan, newPan){
  if (oldPan.x == newPan.x && oldPan.y == newPan.y) {
    // zoom, not a pan
    return newPan;
  }

  const sizes = this.getSizes();
  const viewportRect = DOM.viewport.getBoundingClientRect();

  return {
    x: between(newPan.x, 0, sizes.width - viewportRect.width),
    y: between(newPan.y, 0, sizes.height - viewportRect.height),
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

function handleInput(refresh) {
  const input = (DOM.mainInput.value !== "") ?
    DOM.mainInput.value.trim() :
    defaultInput;
  const cluster = DOM.checkCluster.checked;

  DOM.badgeMarkdown.classList.remove("visible");

  if (DOM.autocomplete === null) {
    const ac = document.getElementsByClassName("autocomplete-suggestions");
    if (ac.length > 0) {
      DOM.autocomplete = ac[0];
    }
  }
  if (DOM.autocomplete !== null) {
    DOM.autocomplete.style.display = "none";
  }

  let url;
  let goRepo;
  if (input.endsWith('.svg')) {
    url = new URL(input);
  } else {
    goRepo = input;
    if (input.startsWith('https://')) {
      goRepo = input.slice('https://'.length);
    } else if (input.startsWith('http://')) {
      goRepo = input.slice('http://'.length);
    }
    DOM.mainInput.value = goRepo;

    url = new URL('/graph/' + goRepo + '.svg', window.location.origin);
    if (cluster === true) {
      url.searchParams.append("cluster", "true");
    }
  }

  hideError();

  const spinner = startSpinner();

  fetch(url, {method: refresh ? 'POST' : 'GET'})
    .then(checkStatus)
    .then(resp => resp.blob())
    .then(blob => {
      const u = goRepo ?
        '/repo/'+goRepo :
        '/svg?url='+url;

      const urlState = new URL(u, window.location.origin);
      if (cluster === true) {
        urlState.searchParams.append("cluster", "true");
      }

      history.pushState(
        { svgHref: url.href, goRepo: goRepo, blob: blob },
        input,
        urlState,
      );

      loadSvg(url.href, goRepo, blob);

      stopSpinner(spinner);
    })
    .catch(error => {
      if (error.response !== undefined) {
        error.response.text().then(text => {
          showError(text);
          stopSpinner(spinner);
        });
      } else {
        showError(error);
        stopSpinner(spinner);
      }
    });
}

function loadSvg(svgHref, goRepo, blob) {
  // createObjectURL() must be coupled with revokeObjectURL(). ownership
  // of svgUrl passes from here to main-svg to thumb-svg.
  const svgUrl = URL.createObjectURL(blob)
  DOM.mainSvg.data = svgUrl;

  DOM.externalSvg.href = svgHref;
  DOM.externalSvg.classList.add("visible");

  if (goRepo) {
    DOM.externalDot.href = svgHref.replace('.svg', '.dot');
    DOM.externalRepo.href = "https://" + goRepo;
    DOM.externalGoDoc.href = "https://pkg.go.dev/" + goRepo;

    DOM.checkCluster.parentElement.classList.add("visible");
    DOM.externalDot.classList.add("visible");
    DOM.externalRepo.classList.add("visible");
    DOM.externalGoDoc.classList.add("visible");
    DOM.refreshButton.classList.add("visible");
    DOM.badge.classList.add("visible");

    DOM.badgeText.value =
      "[![gographs](https://gographs.io/badge.svg)](https://gographs.io/repo/" + goRepo + ")";
  } else {
    DOM.checkCluster.parentElement.remove("visible");
    DOM.externalDot.classList.remove("visible");
    DOM.externalRepo.classList.remove("visible");
    DOM.externalGoDoc.classList.remove("visible");
    DOM.refreshButton.classList.remove("visible");
    DOM.badge.classList.remove("visible");
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
          return;
        }
        handleInput(false);
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
