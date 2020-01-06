function between(val, a, b) {
  if (a < b) {
    return Math.max(a, Math.min(val, b));
  } else {
    return Math.max(b, Math.min(val, a));
  }
}

// based on:
// https://github.com/ariutta/svg-pan-zoom/blob/d107d73120460caae3ecee59192cd29a470e97b0/demo/thumbnailViewer.js

function updateThumbScope() {
  const thumbToMainZoomRatio =  window.thumb.getSizes().realZoom / window.main.getSizes().realZoom;

  let scopeX = window.thumb.getPan().x - window.main.getPan().x * thumbToMainZoomRatio;
  let scopeY = window.thumb.getPan().y - window.main.getPan().y * thumbToMainZoomRatio;
  let scopeWidth = window.main.getSizes().width * thumbToMainZoomRatio;
  let scopeHeight = window.main.getSizes().height * thumbToMainZoomRatio;

  scopeX = Math.max(0, scopeX);
  scopeY = Math.max(0, scopeY);
  scopeWidth = Math.min(window.thumb.getSizes().width, scopeWidth);
  scopeHeight = Math.min(window.thumb.getSizes().height, scopeHeight);

  const scope = document.getElementById('scope');
  scope.setAttribute('x', scopeX + 1);
  scope.setAttribute('y', scopeY + 1);
  scope.setAttribute('width', scopeWidth - 2);
  scope.setAttribute('height', scopeHeight - 2);
};

function updateMainZoomPan(evt){
  if (evt.which == 0 && evt.button == 0) {
    return false;
  }

  const dim = document.getElementById('thumb-svg').getBoundingClientRect();
  const scopeDim = document.getElementById('scope').getBoundingClientRect();

  const mainToThumbZoomRatio =  window.main.getSizes().realZoom / window.thumb.getSizes().realZoom;

  const thumbPanX = Math.min(Math.max(0, evt.clientX - dim.left), dim.width) - scopeDim.width / 2;
  const thumbPanY = Math.min(Math.max(0, evt.clientY - dim.top), dim.height) - scopeDim.height / 2;
  const mainPanX  = - thumbPanX * mainToThumbZoomRatio;
  const mainPanY  = - thumbPanY * mainToThumbZoomRatio;

  window.main.pan({x: mainPanX, y: mainPanY});
}

function bindThumbnail(main, thumb){
  if (!window.main && main) {
    window.main = main;
  }
  if (!window.thumb && thumb) {
    window.thumb = thumb;
  }
  if (!window.main || !window.thumb) {
    return;
  }

  // all function below this expect window.main and window.thumb to be set

  const scopeContainer = document.getElementById('scope-container');
  const thumbSvg = document.getElementById('thumb-svg');

  window.addEventListener('resize', function(event){
    scopeContainer.setAttribute('width', thumbSvg.getBoundingClientRect().width);

    window.main.resize();
    window.thumb.resize();
    window.main.reset();
    window.thumb.reset();
    updateThumbScope();
  });

  // set scope-container to match size of thumbnail svg's 'width: auto'
  scopeContainer.setAttribute('width', thumbSvg.getBoundingClientRect().width);

  // TODO: use document.getElementById('thumb-svg').contentDocument.querySelector('svg') ?
  scopeContainer.addEventListener('click', function(evt){
    updateMainZoomPan(evt);
  });

  scopeContainer.addEventListener('mousemove', function(evt){
    updateMainZoomPan(evt);
  });

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
    function wheelZoom(e) {e.preventDefault()},
    { passive: false }
  );
});

document.getElementById('main-svg').addEventListener('load', function(){
  const mainSvg = this.contentDocument.querySelector('svg');
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

  // Will get called after embed element was loaded
  const main = svgPanZoom(mainSvg, {
    viewportSelector: '#main-svg',
    panEnabled: true,
    controlIconsEnabled: false,
    zoomEnabled: true,
    dblClickZoomEnabled: true,
    mouseWheelZoomEnabled: true,
    preventMouseEventsDefault: false,
    zoomScaleSensitivity: 0.1,
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
  const thumbSvg = this.contentDocument.querySelector('svg');
  thumbSvg.addEventListener(
    'wheel',
    function wheelZoom(e) {e.preventDefault()},
    { passive: false }
  );

  const thumb = svgPanZoom(thumbSvg, {
    zoomEnabled: false,
    panEnabled: false,
    controlIconsEnabled: false,
    dblClickZoomEnabled: false,
    preventMouseEventsDefault: true,
  });

  bindThumbnail(undefined, thumb);
});
