function between(val, a, b) {
  if (a < b) {
    return Math.max(a, Math.min(val, b));
  } else {
    return Math.max(b, Math.min(val, a));
  }
}

// based on:
// https://github.com/ariutta/svg-pan-zoom/blob/d107d73120460caae3ecee59192cd29a470e97b0/demo/thumbnailViewer.js

function updateThumbScope(main, thumb) {
  const scope = document.getElementById('scope');

  const mainPanX   = main.getPan().x;
  const mainPanY   = main.getPan().y;
  const mainWidth  = main.getSizes().width;
  const mainHeight = main.getSizes().height;
  const mainZoom   = main.getSizes().realZoom;
  const thumbPanX  = thumb.getPan().x;
  const thumbPanY  = thumb.getPan().y;
  const thumbZoom  = thumb.getSizes().realZoom;

  const thumByMainZoomRatio =  thumbZoom / mainZoom;

  let scopeX = thumbPanX - mainPanX * thumByMainZoomRatio;
  let scopeY = thumbPanY - mainPanY * thumByMainZoomRatio;
  let scopeWidth = mainWidth * thumByMainZoomRatio;
  let scopeHeight = mainHeight * thumByMainZoomRatio;

  scopeX = Math.max(0, scopeX);
  scopeY = Math.max(0, scopeY);
  scopeWidth = Math.min(thumb.getSizes().width, scopeWidth);
  scopeHeight = Math.min(thumb.getSizes().height, scopeHeight);

  scope.setAttribute("x", scopeX + 1);
  scope.setAttribute("y", scopeY + 1);
  scope.setAttribute("width", scopeWidth - 2);
  scope.setAttribute("height", scopeHeight - 2);
};

function updateMainViewPan(evt){
  if (evt.which == 0 && evt.button == 0) {
    return false;
  }
  // const scope = document.getElementById('thumb-svg');

  const dim = document.getElementById('thumb-svg').getBoundingClientRect();
  const mainWidth   = window.main.getSizes().width;
  const mainHeight  = window.main.getSizes().height;
  const mainZoom    = window.main.getSizes().realZoom;
  const thumbWidth  = window.thumb.getSizes().width;
  const thumbHeight = window.thumb.getSizes().height;
  const thumbZoom   = window.thumb.getSizes().realZoom;

  const scopeDim = document.getElementById("scope").getBoundingClientRect();

  // const thumbPanX = evt.clientX - dim.left - thumbWidth / 2;
  // const thumbPanY = evt.clientY - dim.top - thumbHeight / 2;
  const thumbPanX = Math.min(Math.max(0, evt.clientX - dim.left), dim.width) - scopeDim.width / 2;
  const thumbPanY = Math.min(Math.max(0, evt.clientY - dim.top), dim.height) - scopeDim.height / 2;
  const mainPanX  = - thumbPanX * mainZoom / thumbZoom;
  const mainPanY  = - thumbPanY * mainZoom / thumbZoom;

  console.log("-----------------------");
  console.log("evt.clientX - dim.left:");
  console.log(evt.clientX - dim.left);
  const panX = Math.min(Math.max(0, evt.clientX - dim.left), dim.width);
  console.log("panX:  " + panX);
  const percentX = panX / dim.width;
  console.log("percentX:  " + percentX);

  console.log("window.thumb.getSizes():");
  console.log(window.thumb.getSizes());

  console.log("evt.clientX: " + evt.clientX); // 25 -> 244
  console.log("dim:    " + JSON.toString(dim));
  console.log(dim)
  console.log("thumbWidth:  " + thumbWidth);
  console.log("thumbPanX:   " + thumbPanX);
  console.log("mainPanX:    " + mainPanX);

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

  window.main.setOnZoom(function(level){
    updateThumbScope(window.main, window.thumb);
  });

  window.main.setOnPan(function(point){
    updateThumbScope(window.main, window.thumb);
  });

  updateThumbScope(window.main, window.thumb);
}

window.addEventListener("load", function(){
  const scopeContainer = document.getElementById("scope-container");

  scopeContainer.addEventListener(
    "wheel",
    function wheelZoom(e) {e.preventDefault()},
    { passive: false }
  );

  // TODO: use document.getElementById('thumb-svg').contentDocument.querySelector("svg")
  scopeContainer.addEventListener('click', function(evt){
    updateMainViewPan(evt);
  });

  scopeContainer.addEventListener('mousemove', function(evt){
    updateMainViewPan(evt);
  });
});

document.getElementById('main-svg').addEventListener('load', function(){
  // prevent zoom scroll events from bubbling up
  const mainSvg = this.contentDocument.querySelector("svg");
  mainSvg.addEventListener(
    "wheel",
    function wheelZoom(e) {e.preventDefault()},
    { passive: false }
  );

  const beforePan = function(oldPan, newPan){
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
    viewportSelector: '.svg-pan-zoom_viewport',
    panEnabled: true,
    controlIconsEnabled: false,
    zoomEnabled: true,
    dblClickZoomEnabled: true,
    mouseWheelZoomEnabled: true,
    preventMouseEventsDefault: false,
    zoomScaleSensitivity: 0.2,
    minZoom: 1,
    maxZoom: 10,
    fit: true,
    contain: false,
    center: true,
    refreshRate: 'auto',
    beforeZoom: function(){},
    onZoom: function(){},
    beforePan: beforePan,
    onPan: function(){},
    onUpdatedCTM: function(){},
    // customEventsHandler: {},
    eventsListenerElement: null,
  });

  bindThumbnail(main, undefined);
})

document.getElementById('thumb-svg').addEventListener('load', function(){
  const thumbSvg = this.contentDocument.querySelector("svg");
  thumbSvg.addEventListener(
    "wheel",
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

  // const scopeContainer = document.getElementById('scopeContainer');
  bindThumbnail(undefined, thumb);
});
