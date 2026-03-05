(function () {
  'use strict';

  // ---- State ----
  const state = {
    components: {},   // name -> {name, type, state}
    edges: [],        // [{from, to, fieldName, depType}]
    events: [],
    breakpoints: {},  // name -> bool
    selectedComp: null,
    mode: 0
  };

  // ---- DOM refs ----
  const compList = document.getElementById('comp-list');
  const timeline = document.getElementById('timeline');
  const graphCanvas = document.getElementById('graph-canvas');
  const detailInfo = document.getElementById('detail-info');
  const detailDeps = document.getElementById('detail-deps');
  const detailDependents = document.getElementById('detail-dependents');
  const btnNext = document.getElementById('btn-next');
  const modeButtons = document.querySelectorAll('.mode-group button');

  // ---- API ----
  function post(path, body) {
    return fetch('/api/' + path, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(body)
    });
  }

  btnNext.addEventListener('click', function () { post('next'); });

  modeButtons.forEach(function (btn) {
    btn.addEventListener('click', function () {
      var mode = parseInt(btn.dataset.mode);
      state.mode = mode;
      post('mode', { mode: mode });
      modeButtons.forEach(function (b) { b.classList.toggle('active', b === btn); });
    });
  });

  // ---- SSE ----
  var evtSource = new EventSource('/api/events');
  evtSource.onmessage = function (e) {
    var ev = JSON.parse(e.data);
    state.events.push(ev);
    handleEvent(ev);
    renderTimeline();
    renderComponents();
    renderGraph();
  };

  function handleEvent(ev) {
    switch (ev.action) {
      case 'component_registered':
        state.components[ev.componentName] = {
          name: ev.componentName,
          type: (ev.details && ev.details.type) || '',
          state: 'registered'
        };
        break;
      case 'definition_scanned':
        if (ev.details && Array.isArray(ev.details.components)) {
          ev.details.components.forEach(function (n) {
            if (state.components[n]) state.components[n].state = 'scanned';
          });
        } else if (state.components[ev.componentName]) {
          state.components[ev.componentName].state = 'scanned';
        }
        break;
      case 'component_creating':
        if (state.components[ev.componentName]) state.components[ev.componentName].state = 'creating';
        break;
      case 'populating':
        if (state.components[ev.componentName]) state.components[ev.componentName].state = 'populating';
        break;
      case 'dependency_injected':
        if (ev.details && ev.details.dependency) {
          state.edges.push({
            from: ev.componentName,
            to: ev.details.dependency,
            fieldName: ev.details.field || '',
            depType: ev.details.depType || 'pointer'
          });
        }
        break;
      case 'before_initialization':
        if (state.components[ev.componentName]) state.components[ev.componentName].state = 'initializing';
        break;
      case 'component_ready':
        if (state.components[ev.componentName]) state.components[ev.componentName].state = 'ready';
        break;
    }
  }

  // ---- Render Components ----
  function renderComponents() {
    var names = Object.keys(state.components).sort();
    compList.innerHTML = '';
    names.forEach(function (name) {
      var c = state.components[name];
      var el = document.createElement('div');
      el.className = 'comp-item' + (state.selectedComp === name ? ' selected' : '');
      var shortName = name.length > 28 ? '...' + name.slice(-25) : name;
      el.innerHTML =
        '<input type="checkbox"' + (state.breakpoints[name] ? ' checked' : '') + '>' +
        '<span class="state-dot ' + c.state + '"></span>' +
        '<span class="comp-name" title="' + escHtml(name) + '">' + escHtml(shortName) + '</span>';
      el.querySelector('input').addEventListener('change', function (e) {
        state.breakpoints[name] = e.target.checked;
        post('breakpoint', { component: name, enabled: e.target.checked });
      });
      el.addEventListener('click', function (e) {
        if (e.target.tagName === 'INPUT') return;
        state.selectedComp = name;
        renderComponents();
        renderDetail();
        highlightGraphNode(name);
      });
      compList.appendChild(el);
    });
  }

  // ---- Render Timeline ----
  function renderTimeline() {
    timeline.innerHTML = '';
    state.events.forEach(function (ev, i) {
      var el = document.createElement('div');
      el.className = 'tl-event' + (i === state.events.length - 1 ? ' current' : '');
      if (ev.action === 'phase_start' || ev.action === 'phase_end') {
        el.classList.add('phase');
      }
      el.innerHTML = formatEvent(ev);
      timeline.appendChild(el);
    });
    timeline.scrollTop = timeline.scrollHeight;
  }

  function formatEvent(ev) {
    var parts = ['<span class="action">' + escHtml(ev.action) + '</span>'];
    if (ev.componentName) parts.push('<span class="comp">' + escHtml(displayName(ev.componentName)) + '</span>');
    if (ev.processorName) parts.push('<span class="processor">' + escHtml(displayName(ev.processorName)) + '</span>');
    if (ev.details) {
      if (Array.isArray(ev.details.components)) {
        var cnames = ev.details.components.map(function (c) { return displayName(c); });
        parts.push('<span class="comp">[' + escHtml(cnames.join(', ')) + ']</span>');
      }
      if (Array.isArray(ev.details.processors)) {
        var pnames = ev.details.processors.map(function (p) { return displayName(p); });
        parts.push('<span class="processor">[' + escHtml(pnames.join(', ')) + ']</span>');
      }
      var extra = [];
      Object.keys(ev.details).forEach(function (k) {
        if (k === 'type' || k === 'components' || k === 'processors') return;
        extra.push(k + '=' + ev.details[k]);
      });
      if (extra.length) parts.push('<span class="detail">' + escHtml(extra.join(', ')) + '</span>');
    }
    return parts.join(' ');
  }

  // ---- Render Detail ----
  function renderDetail() {
    var name = state.selectedComp;
    if (!name || !state.components[name]) {
      detailInfo.textContent = 'Select a component to view details';
      detailDeps.innerHTML = '';
      detailDependents.innerHTML = '';
      return;
    }
    var c = state.components[name];
    detailInfo.innerHTML = '<strong>' + escHtml(name) + '</strong><br>Type: ' + escHtml(c.type) + '<br>State: ' + c.state;

    var deps = state.edges.filter(function (e) { return e.from === name; });
    detailDeps.innerHTML = deps.length === 0 ? '<li style="color:var(--text2)">None</li>' : '';
    deps.forEach(function (d) {
      var li = document.createElement('li');
      var cls = d.depType === 'interface' ? 'dep-interface' : 'dep-pointer';
      li.innerHTML = '<span class="' + cls + '">' + escHtml(displayName(d.to)) + '</span>' +
        '<span class="dep-type">(' + escHtml(d.depType) + (d.fieldName ? ', field: ' + escHtml(d.fieldName) : '') + ')</span>';
      detailDeps.appendChild(li);
    });

    var dependents = state.edges.filter(function (e) { return e.to === name; });
    detailDependents.innerHTML = dependents.length === 0 ? '<li style="color:var(--text2)">None</li>' : '';
    dependents.forEach(function (d) {
      var li = document.createElement('li');
      var cls = d.depType === 'interface' ? 'dep-interface' : 'dep-pointer';
      li.innerHTML = '<span class="' + cls + '">' + escHtml(displayName(d.from)) + '</span>' +
        '<span class="dep-type">(' + escHtml(d.depType) + ')</span>';
      detailDependents.appendChild(li);
    });
  }

  // ---- Graph Rendering (SVG force-directed) ----
  var graphNodes = [], graphEdges = [];
  var svgNS = 'http://www.w3.org/2000/svg';
  var simulation = null;
  var highlightedNode = null;

  function graphSize() {
    return {
      w: graphCanvas.clientWidth || 600,
      h: graphCanvas.clientHeight || 400
    };
  }

  function layoutGraph(names, sz) {
    var w = sz.w, h = sz.h;
    var padX = 80, padY = 50;

    var outMap = {}, inMap = {};
    names.forEach(function (n) { outMap[n] = []; inMap[n] = []; });
    state.edges.forEach(function (e) {
      if (outMap[e.from] && inMap[e.to]) {
        outMap[e.from].push(e.to);
        inMap[e.to].push(e.from);
      }
    });

    var depth = {};
    var visiting = {};
    function dfs(node) {
      if (depth[node] !== undefined) return depth[node];
      if (visiting[node]) return 0;
      visiting[node] = true;
      var d = 0;
      outMap[node].forEach(function (dep) {
        d = Math.max(d, dfs(dep) + 1);
      });
      depth[node] = d;
      visiting[node] = false;
      return d;
    }
    names.forEach(function (n) { dfs(n); });

    var maxDepth = 0;
    names.forEach(function (n) { if (depth[n] > maxDepth) maxDepth = depth[n]; });
    if (maxDepth === 0) maxDepth = 1;

    var layers = {};
    names.forEach(function (n) {
      var r = depth[n] || 0;
      if (!layers[r]) layers[r] = [];
      layers[r].push(n);
    });

    var nodeIdx = {};
    for (var r in layers) {
      layers[r].forEach(function (n, i) { nodeIdx[n] = i; });
    }

    for (var pass = 0; pass < 4; pass++) {
      for (var rank = maxDepth; rank >= 0; rank--) {
        if (!layers[rank]) continue;
        layers[rank].forEach(function (n) {
          var neighbors = outMap[n].concat(inMap[n]);
          if (neighbors.length === 0) return;
          var sum = 0;
          neighbors.forEach(function (nb) { sum += (nodeIdx[nb] || 0); });
          nodeIdx[n] = sum / neighbors.length;
        });
        layers[rank].sort(function (a, b) { return (nodeIdx[a] || 0) - (nodeIdx[b] || 0); });
        layers[rank].forEach(function (n, i) { nodeIdx[n] = i; });
      }
      for (var rank2 = 0; rank2 <= maxDepth; rank2++) {
        if (!layers[rank2]) continue;
        layers[rank2].forEach(function (n) {
          var neighbors = outMap[n].concat(inMap[n]);
          if (neighbors.length === 0) return;
          var sum = 0;
          neighbors.forEach(function (nb) { sum += (nodeIdx[nb] || 0); });
          nodeIdx[n] = sum / neighbors.length;
        });
        layers[rank2].sort(function (a, b) { return (nodeIdx[a] || 0) - (nodeIdx[b] || 0); });
        layers[rank2].forEach(function (n, i) { nodeIdx[n] = i; });
      }
    }

    var positions = {};
    var usableW = w - 2 * padX;
    var layerH = (h - 2 * padY) / (maxDepth || 1);

    for (var rk in layers) {
      var row = layers[rk];
      var count = row.length;
      var spacing = usableW / (count + 1);
      var yPos = padY + (maxDepth - parseInt(rk)) * layerH;
      row.forEach(function (n, i) {
        positions[n] = {
          x: padX + spacing * (i + 1),
          y: yPos,
          rank: parseInt(rk)
        };
      });
    }

    return positions;
  }

  function renderGraph() {
    var names = Object.keys(state.components);
    if (names.length === 0) return;

    var sz = graphSize();
    var positions = layoutGraph(names, sz);

    graphNodes = names.map(function (name) {
      var existing = graphNodes.find(function (n) { return n.name === name; });
      var pos = positions[name] || { x: sz.w / 2, y: sz.h / 2, rank: 0 };
      if (existing) {
        existing.state = state.components[name].state;
        existing.rank = pos.rank;
        existing.targetX = pos.x;
        existing.targetY = pos.y;
        return existing;
      }
      return {
        name: name,
        state: state.components[name].state,
        rank: pos.rank,
        targetX: pos.x,
        targetY: pos.y,
        x: pos.x,
        y: pos.y,
        vx: 0, vy: 0
      };
    });

    graphEdges = state.edges.map(function (e) {
      return {
        source: graphNodes.find(function (n) { return n.name === e.from; }),
        target: graphNodes.find(function (n) { return n.name === e.to; }),
        depType: e.depType
      };
    }).filter(function (e) { return e.source && e.target; });

    if (!simulation) {
      simulation = createSimulation();
    }
    drawSVG();
  }

  function createSimulation() {
    var sz = graphSize();
    var w = sz.w, h = sz.h;
    var alpha = 1;
    var running = true;
    var pad = 40;

    function tick() {
      if (!running) return;
      alpha *= 0.96;
      if (alpha < 0.001) { alpha = 0; }

      var n = graphNodes.length;

      for (var i = 0; i < n; i++) {
        for (var j = i + 1; j < n; j++) {
          var a = graphNodes[i], b = graphNodes[j];
          var dx = b.x - a.x, dy = b.y - a.y;
          var dist = Math.sqrt(dx * dx + dy * dy) || 1;
          var minDist = 90;
          if (dist < minDist) {
            var push = (minDist - dist) * 0.3 * alpha;
            var px = dx / dist * push, py = dy / dist * push;
            a.vx -= px; a.vy -= py;
            b.vx += px; b.vy += py;
          }
        }
      }

      graphNodes.forEach(function (nd) {
        if (nd.targetX !== undefined) {
          nd.vx += (nd.targetX - nd.x) * 0.12 * alpha;
          nd.vy += (nd.targetY - nd.y) * 0.12 * alpha;
        }
        nd.vx *= 0.4; nd.vy *= 0.4;
        nd.x += nd.vx; nd.y += nd.vy;
        nd.x = Math.max(pad, Math.min(w - pad, nd.x));
        nd.y = Math.max(pad, Math.min(h - pad, nd.y));
      });

      drawSVG();
      if (alpha > 0) requestAnimationFrame(tick);
    }

    return {
      reheat: function () {
        alpha = 0.8;
        sz = graphSize();
        w = sz.w; h = sz.h;
        var positions = layoutGraph(Object.keys(state.components), sz);
        graphNodes.forEach(function (nd) {
          var pos = positions[nd.name];
          if (pos) {
            nd.targetX = pos.x;
            nd.targetY = pos.y;
          }
        });
        tick();
      },
      stop: function () { running = false; }
    };
  }

  var cam = { x: 0, y: 0, zoom: 1 };
  var svgEl = null, worldG = null;
  var isPanning = false, panStart = { x: 0, y: 0 }, camStart = { x: 0, y: 0 };

  function ensureSVG() {
    if (svgEl) return;
    var w = graphCanvas.clientWidth || 600;
    var h = graphCanvas.clientHeight || 400;
    svgEl = document.createElementNS(svgNS, 'svg');
    svgEl.setAttribute('viewBox', '0 0 ' + w + ' ' + h);
    svgEl.style.cursor = 'grab';

    var defs = document.createElementNS(svgNS, 'defs');
    var marker = document.createElementNS(svgNS, 'marker');
    marker.setAttribute('id', 'arrow');
    marker.setAttribute('viewBox', '0 0 10 10');
    marker.setAttribute('refX', '22'); marker.setAttribute('refY', '5');
    marker.setAttribute('markerWidth', '6'); marker.setAttribute('markerHeight', '6');
    marker.setAttribute('orient', 'auto-start-reverse');
    var path = document.createElementNS(svgNS, 'path');
    path.setAttribute('d', 'M 0 0 L 10 5 L 0 10 z');
    marker.appendChild(path);
    defs.appendChild(marker);
    var marker2 = marker.cloneNode(true);
    marker2.setAttribute('id', 'arrow-iface');
    marker2.classList.add('interface-marker');
    defs.appendChild(marker2);
    svgEl.appendChild(defs);

    worldG = document.createElementNS(svgNS, 'g');
    worldG.setAttribute('id', 'graph-world');
    svgEl.appendChild(worldG);

    graphCanvas.innerHTML = '';
    graphCanvas.appendChild(svgEl);

    svgEl.addEventListener('mousedown', function (e) {
      if (e.button !== 0) return;
      isPanning = true;
      panStart.x = e.clientX; panStart.y = e.clientY;
      camStart.x = cam.x; camStart.y = cam.y;
      svgEl.style.cursor = 'grabbing';
    });
    window.addEventListener('mousemove', function (e) {
      if (!isPanning) return;
      cam.x = camStart.x + (e.clientX - panStart.x);
      cam.y = camStart.y + (e.clientY - panStart.y);
      applyTransform();
    });
    window.addEventListener('mouseup', function () {
      if (!isPanning) return;
      isPanning = false;
      svgEl.style.cursor = 'grab';
    });
    svgEl.addEventListener('wheel', function (e) {
      e.preventDefault();
      var rect = svgEl.getBoundingClientRect();
      var mx = e.clientX - rect.left;
      var my = e.clientY - rect.top;
      var factor = e.deltaY < 0 ? 1.02 : 1 / 1.02;
      var newZoom = Math.max(0.15, Math.min(5, cam.zoom * factor));
      // keep world point under cursor fixed
      cam.x = mx - (mx - cam.x) / cam.zoom * newZoom;
      cam.y = my - (my - cam.y) / cam.zoom * newZoom;
      cam.zoom = newZoom;
      applyTransform();
    }, { passive: false });
  }

  function applyTransform() {
    if (!worldG) return;
    worldG.setAttribute('transform',
      'translate(' + cam.x + ',' + cam.y + ') scale(' + cam.zoom + ')');
  }

  function drawSVG() {
    var w = graphCanvas.clientWidth || 600;
    var h = graphCanvas.clientHeight || 400;
    ensureSVG();
    svgEl.setAttribute('viewBox', '0 0 ' + w + ' ' + h);

    while (worldG.firstChild) worldG.removeChild(worldG.firstChild);

    graphEdges.forEach(function (e) {
      if (!e.source || !e.target) return;
      var line = document.createElementNS(svgNS, 'line');
      line.classList.add('edge');
      line.classList.add(e.depType === 'interface' ? 'edge-interface' : 'edge-pointer');
      line.setAttribute('x1', e.source.x); line.setAttribute('y1', e.source.y);
      line.setAttribute('x2', e.target.x); line.setAttribute('y2', e.target.y);
      line.setAttribute('marker-end', e.depType === 'interface' ? 'url(#arrow-iface)' : 'url(#arrow)');
      worldG.appendChild(line);
    });

    graphNodes.forEach(function (n) {
      var g = document.createElementNS(svgNS, 'g');
      g.classList.add('gnode');
      var circle = document.createElementNS(svgNS, 'circle');
      circle.classList.add('node-circle');
      if (n.state) circle.classList.add(n.state);
      if (highlightedNode === n.name) circle.classList.add('highlight');
      circle.setAttribute('cx', n.x); circle.setAttribute('cy', n.y);
      circle.setAttribute('r', '12');
      circle.addEventListener('click', function (ev) {
        ev.stopPropagation();
        state.selectedComp = n.name;
        highlightedNode = n.name;
        renderComponents();
        renderDetail();
        drawSVG();
      });
      g.appendChild(circle);

      var label = document.createElementNS(svgNS, 'text');
      label.classList.add('node-label');
      label.setAttribute('x', n.x); label.setAttribute('y', n.y + 22);
      label.setAttribute('text-anchor', 'middle');
      label.textContent = shortName(n.name);
      g.appendChild(label);
      worldG.appendChild(g);
    });

    applyTransform();
  }

  function highlightGraphNode(name) {
    highlightedNode = name;
    drawSVG();
  }

  // reheat graph when data changes
  var graphTimer = null;
  var origRender = renderGraph;
  renderGraph = function () {
    origRender();
    if (simulation) {
      clearTimeout(graphTimer);
      graphTimer = setTimeout(function () { simulation.reheat(); }, 50);
    }
  };

  // ---- Helpers ----
  function shortName(name) {
    if (!name) return '';
    var parts = name.split('/');
    var last = parts[parts.length - 1];
    return last.length > 18 ? last.slice(0, 15) + '...' : last;
  }

  function displayName(name) {
    if (!name) return '';
    var parts = name.split('/');
    return parts[parts.length - 1];
  }

  function escHtml(s) {
    if (!s) return '';
    return s.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;').replace(/"/g, '&quot;');
  }

  // load initial state and check dry run
  fetch('/api/state').then(function (r) { return r.json(); }).then(function (data) {
    if (data.dryRun) {
      var badge = document.getElementById('dry-run-badge');
      if (badge) badge.classList.remove('hidden');
    }
  }).catch(function () {});

  // initial render
  renderComponents();
  renderTimeline();
})();
