// User
const username = localStorage.getItem('username');
const signOutBtn = document.getElementById("sign-out-btn")
signOutBtn.onclick = () => {
    signOutBtn.classList.add('loading')
    fetch('/api/logout').then(response => response.json()).then(data => {
        if (data.status == 200) location = "/login"
        else signOutBtn.classList.remove('loading')
    })
}

// Logs
const logList = [];
const logBox = document.getElementById('log');
const addZero = (number) => {
    return number < 10 ? "0" + number : number
}
const addLog = async (text, color = "blue") => {
    const now = new Date();
    const time = now.getHours() + ":" + addZero(now.getMinutes()) + ":" + addZero(now.getSeconds());
    logList.push({ time: time, text: text });
    logBox.innerHTML += `<div><div class="ui ${color} horizontal label">${time}</div> ${text}</div>`
    localStorage.setItem('log', JSON.stringify(logList));
    localStorage.setItem('log-time', `${now.getFullYear()}-${addZero(now.getMonth() + 1)}-${addZero(now.getDate())} ${time}`);
    logBox.scrollTop = logBox.scrollHeight;
}

// Set Upload Info
const uploadName = document.getElementById('upload-name');
const uploadId = document.getElementById('upload-id');
const setStreamUploadInfo = (id, name) => {
    uploadName.value = name;
    uploadId.value = id;
}

// Get Stream
const streamSearch = document.getElementById('stream-search');
const getStream = (page, search) => {
    const isSearchMode = search != undefined;
    if (isSearchMode) streamSearch.classList.add('loading');
    fetch(!isSearchMode ? '/api/stream/get' : "/api/stream/search", {
        method: 'POST',
        headers: {
            'Content-Type': 'application/x-www-form-urlencoded'
        },
        body: !isSearchMode ? "page=" + page : "page=" + page + "&search=" + search
    }).then(response => response.json()).then(data => {
        switch (data.status) {
            case 200:
                addLog("Connected to server...");

                if (isSearchMode) streamSearch.classList.remove('loading');

                const streamBox = document.getElementById('stream-item-box');
                streamBox.innerHTML = "";
                for (let i = 0; i < data.data.length; i++) {
                    streamBox.innerHTML += `<div class="stream-item" onclick="setStreamUploadInfo(${data.data[i].id},'${data.data[i].name}')">
                    <div class="ui orange horizontal label">ID ${data.data[i].id}</div>
                    <b>${data.data[i].name}</b>
                    <span class="right">${data.data[i].count} Ep</span>
                </div>`
                }
                document.getElementById('stream-total').innerText = data.count

                setStreamPage(page, data.count)
                break;
            case 500: addLog(data.msg, "red"); break;
            case 403: location = "/login"; break;
        }
    })
}
let searchTimeout = null;
streamSearch.onkeydown = () => {
    clearTimeout(searchTimeout);
    searchTimeout = setTimeout(() => streamSearch.value != "" ?
        getStream(1, streamSearch.value) :
        getStream(1), 500)
}

// Upload Stream
const uploadFile = document.getElementById('upload-file');
const uploadFileName = document.getElementById('upload-file-name');
const uploadArchiveIcon = document.getElementById('upload-archive-icon');
const uploadBtn = document.getElementById('upload-btn');
const uploadPreview = document.getElementById('upload-preview');
uploadFile.onchange = () => {
    uploadArchiveIcon.style.display = 'none';
    uploadPreview.style.display = 'block';
    uploadPreview.src = URL.createObjectURL(uploadFile.files[0]);
    uploadFileName.innerText = uploadFile.files[0].name;
}
uploadBtn.onclick = () => {
    if (uploadFile.files.length == 0) {
        openDialog("Error", "Please select a file to upload.", 'mini')
        return
    } else if (uploadName.value == "" || uploadId.value == "") {
        openDialog("Error", "Please select a stream to upload.", 'mini')
        return
    }
    uploadBtn.classList.add('loading');

    const form = new FormData();
    form.append('file', uploadFile.files[0]);
    form.append('id', uploadId.value);

    fetch('/api/stream/upload', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/x-www-form-urlencoded'
        },
        body: form,
    }).then(response => response.json()).then(data => {
        uploadBtn.classList.remove('loading');
        switch (data.status) {
            case 200: addLog(data.msg); break;
            case 500: addLog(data.msg, "red"); break;
        }
    })

    // End
    uploadFile.value = "";
    uploadPreview.style.display = 'none';
    uploadArchiveIcon.style.display = 'block';
    uploadFileName.innerText = "No file selected";
}

// System Info
const bToGB = (b) => { return (b / 1024 / 1024 / 1024).toFixed(1) }
const cpuValue = document.getElementById('cpu-value');
const memValue = document.getElementById('mem-value');
const memInfo = document.getElementById('mem-info');
const tempValue = document.getElementById('temp-value');
const getSystemInfo = () => {
    fetch('/api/system/info').then(response => response.json()).then(data => {
        cpuValue.innerText = data.cpu.percent.toFixed(2) + '%'
        memValue.innerText = data.mem.percent + '%'
        memInfo.innerText = " " + bToGB(data.mem.used) + '/' + bToGB(data.mem.total) + "G"
        tempValue.innerText = data.cpu.temp
    })
    setTimeout(getSystemInfo, 2000)
}; getSystemInfo();

// Open Dialog
const dialog = document.getElementById('dialog');
const dialogModal = document.getElementById('dialog-modal');
const dialogTitle = document.getElementById('dialog-title');
const dialogCancel = document.getElementsByClassName('close');
const openDialog = (title, content, size) => {
    dialog.style.visibility = "visible";
    dialogTitle.innerText = title;
    dialog.querySelector('.content').innerText = content;
    dialog.classList.add('active');
    if (size != undefined) dialogModal.classList.add(size);
    for (let i in dialogCancel) dialogCancel[i].onclick = () => {
        dialog.classList.remove('active');
        setTimeout(() => {
            if (size != undefined) dialogModal.classList.remove(size);
            dialog.style.visibility = "hidden";
        }, 300)
    }
}

// Send Command
const commandInput = document.getElementById('command');
const commandBtn = document.getElementById('command-btn');
const sendCommand = () => {
    const command = commandInput.value;
    if (command != "") {
        commandBtn.classList.add('loading');
        commandInput.value = "";
        document.getElementById('log').innerHTML += `<div><div class="ui green horizontal label">${username}</div> ${command}</div>`
        fetch('/api/system/command', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/x-www-form-urlencoded'
            },
            body: "command=" + command
        }).then(response => response.json()).then(data => {
            commandBtn.classList.remove('loading');
            switch (data.status) {
                case 200: addLog(data.msg); break;
                case 500: addLog(data.msg, "red"); break;
                case 403: location = "/login"; break;
            }
        })
    }
}
commandBtn.onclick = sendCommand;
commandInput.onkeyup = (e) => {
    if (e.key == "Enter") sendCommand()
}

// Get History
const logHistory = localStorage.getItem('log');
const logHistoryTime = localStorage.getItem('log-time');
if (logHistory != null && logHistoryTime != null) {
    const logs = JSON.parse(logHistory);
    for (let i = 0; i < logs.length; i++) {
        document.getElementById('log').innerHTML += `<div><div class="ui blue horizontal label"> ${logs[i].time}</div> ${logs[i].text}<div>`
    }
    document.getElementById('log').innerHTML += `<div class="ui horizontal divider" style="margin: 0">Last Logs ${logHistoryTime}</div>`
}

// Clear History
document.getElementById("clear-history").onclick = () => {
    localStorage.removeItem('log');
    localStorage.removeItem('log-time');
    document.getElementById('log').innerHTML = `<div class="ui horizontal divider" style="margin: 0">Cleared</div>`
}

// Stream Page
const streamBtnGroup = document.getElementById('stream-btn-group');
const setStreamPage = (page, count) => {
    const maxPage = Math.ceil(count / 20);
    let html = "";
    if (page > 1) {
        html += `<button class="ui icon button" onclick="getStream(${page - 1})"><i class="icon angle left"></i></button>`;
    }
    for (let i = Math.max(1, page - 2); i < page; i++) {
        html += `<button class="ui button" onclick="getStream(${i})">${i}</button>`;
    }
    for (let i = page; i < page + 2 && i <= maxPage; i++) {
        html += `<button class="ui button${i == page ? " disabled" : ""}" onclick="getStream(${i})">${i}</button>`;
    }
    if (page < maxPage) {
        html += `<button class="ui icon button" onclick="getStream(${page + 1})"><i class="icon angle right"></i></button>`;
    }
    streamBtnGroup.innerHTML = html;
}

// Main
if (window.screen.width < 650) {
    openDialog("Warning", "This website is not supported on mobile devices.")
}

getStream(1)