
async function serverConnect() {
    let url = `ws${location.protocol === 'https:' ? 's' : ''}://${location.host}/ws`;
    let socket = new WebSocket(url);

    socket.onerror = function(e) {
        // TODO
    };

    socket.onopen = function(e) {
        // TODO
    };

    socket.onclose = function(e) {
        // TODO
    };

    socket.onmessage = function(e) {
        let m = JSON.parse(e.data);
        switch(m.type) {
        case 'start': {
            let station = document.getElementsByClassName('station')[0];
            let timer = document.getElementById('timer');
            if(timer) {
                station.removeChild(timer);
            }
            handleProgram(m);
            break;
        }
        case 'wait': {
            let station = document.getElementsByClassName('station')[0];
            let timer = document.getElementById('timer');
            let exists = true;
            cleanPlayer();
            if(!timer) {
                exists = false;
                timer = document.createElement('p');
            }
            timer.id = 'timer';
            startTimer(new Date(m.wait).getTime(), timer);
            if(!exists) {
                station.appendChild(timer);
            }
            break;
        }
        }
    };
}

function handleProgram(m) {
    let player = document.getElementById('player');
    let div = document.createElement('div');
    let overlap = document.createElement('div');
    let video = document.createElement('video');
    let initial_source = '/program/now.m3u8?version=';
    let video_height = "80vmin";
    let version = Math.floor((Math.random() * 10000) + 1);
    let source = initial_source.concat(version.toString());
    let fullscreen_change = function() {
        if(video.style.height === video_height) {
            video.muted = false;
            video.style.height = "100%";
            div.removeChild(overlap);
        } else {
            video.muted = true;
            video.style.height = video_height;
            div.appendChild(overlap);
        }
    };
    let onload = function() {
        video.muted = true;
        video.play();
        div.onclick = function() {
            div.requestFullscreen();
        };
        player.onfullscreenchange = fullscreen_change;
    };

    overlap.id = 'overlap';
    video.id = 'now';
    video.style.height = video_height;

    if(Hls.isSupported()) {
        let hls = new Hls();
        hls.loadSource(source);
        hls.attachMedia(video);
        hls.on(Hls.Events.MANIFEST_PARSED, onload);
    } else if(video.canPlayType('application/vnd.apple.mpegurl')) {
        video.src = source;
        video.addEventListener('loadedmetadata', onload);
    }
    // TODO: else ?

    let timer = document.createElement('p');
    timer.id = 'timer';
    timer.textContent = "It's Alive";
    overlap.appendChild(timer);
    div.style.position = 'relative';

    div.appendChild(video);
    div.appendChild(overlap);
    player.appendChild(div);
}

function cleanPlayer() {
    let player = document.getElementById('player');
    while(player.firstChild) {
        player.removeChild(player.lastChild);
    }
}

function start() {
    try {
        serverConnect();
    } catch(e) {
        console.error(e);
    }
}

start();
