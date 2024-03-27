// Last State
let lastGetStream = { "page": 1, "search": null };

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
const addLog = async (text, color = "blue", useProgress = false, title) => {
    const now = new Date();
    if (title == undefined) {
        title = addZero(now.getHours()) + ":" + addZero(now.getMinutes()) + ":" + addZero(now.getSeconds());
        localStorage.setItem('log-time', `${now.getFullYear()}-${addZero(now.getMonth() + 1)}-${addZero(now.getDate())} ${title}`);
    }
    const index = logList.push({ time: title, text: text, color });
    if (useProgress)
        logBox.innerHTML += `<div id="log-${index - 1}"><div class="ui olive horizontal label">${title}</div> ${text} <div class="ui progress success log-progress"><div class="bar" id="log-pb-${index - 1}"><div class="progress"></div></div></div></div>`
    else
        logBox.innerHTML += `<div id="log-${index - 1}"><div class="ui ${color} horizontal label">${title}</div> ${text.replace(/\n/g, "<br />")}</div>`
    localStorage.setItem('log', JSON.stringify(logList));
    logBox.scrollTop = logBox.scrollHeight;

    return index - 1;
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
const getStream = (page, search, f = false) => {
    lastGetStream = { page, search };
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
                if (f) addLog("Connected to server...");
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

// Get Status
let lastStatus = {};
let lastStatusLength = {}
const getStatus = (key) => {
    const slen = lastStatusLength[key] == undefined ? 0 : lastStatusLength[key];
    fetch('/api/status/get?key=' + key).then(response => response.json()).then(data => {
        if (data.status == 200) {
            for (let i = slen; i < data.data.length; i++) {
                if (data.data[i].percent == -1) {
                    addLog(data.data[i].msg, data.data[i].msg.includes("Error") ? "red" : "blue")

                    if (data.data[i].msg.includes("Stream Upload success")) getStream(lastGetStream.page, lastGetStream.search);
                    else if (data.data[i].msg.includes("Error")) return
                } else {
                    addLog(data.data[i].msg, "olive", true).then(index => lastStatus[i] = { msg: data.data[i].msg, index });
                }
            }

            for (let i = 0; i < slen; i++) {
                if (lastStatus[i] != undefined) {
                    document.getElementById(`log-pb-${lastStatus[i].index}`).style.width = data.data[i].percent + "%";
                    logList[lastStatus[i].index].text = data.data[i].msg + " " + data.data[i].percent + "%";
                    if (data.data[i].percent >= 100) {
                        delete lastStatus[i];
                    }
                }
            }

            lastStatusLength[key] = data.data.length;

            setTimeout(() => getStatus(key), 1000)
        }
    })
};

// Upload Stream
const uploadFile = document.getElementById('upload-file');
const uploadFileName = document.getElementById('upload-file-name');
const uploadArchiveIcon = document.getElementById('upload-archive-icon');
const uploadBtn = document.getElementById('upload-btn');
const uploadPreview = document.getElementById('upload-preview');
const uploadProgress = document.getElementById("upload-progress")
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

    // fetch('/api/stream/upload', {
    //     method: 'POST',
    //     body: form,
    // }).then(response => response.json()).then(data => {
    //     uploadBtn.classList.remove('loading');
    //     switch (data.status) {
    //         case 200:
    //             // getStream(lastGetStream.page, lastGetStream.search);
    //             getStatus(data.task_id);
    //             break;
    //         case 500: openDialog('Error', data.msg); addLog(data.msg, "red"); break;
    //         case 404: openDialog('Error', data.msg); addLog(data.msg, "red"); break;
    //     }

    //     // End
    //     uploadFile.value = "";
    //     uploadId.value = "";
    //     uploadName.value = "";
    //     uploadPreview.style.display = 'none';
    //     uploadArchiveIcon.style.display = 'block';
    //     uploadFileName.innerText = "No file selected";
    // })

    let xhr = new XMLHttpRequest();
    xhr.open('POST', '/api/stream/upload', true);
    xhr.upload.onprogress = function (e) {
        if (e.lengthComputable) {
            let percentComplete = (e.loaded / e.total) * 100;
            uploadProgress.style.width = percentComplete + '%';
        }
    };
    xhr.onload = function () {
        if (this.status == 200) {
            let data = JSON.parse(this.response);

            uploadBtn.classList.remove('loading');
            uploadProgress.style.width = "0%";
            switch (data.status) {
                case 200:
                    getStatus(data.task_id);
                    break;
                case 500: openDialog('Error', data.msg); addLog(data.msg, "red"); break;
                case 404: openDialog('Error', data.msg); addLog(data.msg, "red"); break;
            }

            // End
            uploadFile.value = "";
            uploadId.value = "";
            uploadName.value = "";
            uploadPreview.style.display = 'none';
            uploadArchiveIcon.style.display = 'block';
            uploadFileName.innerText = "No file selected";
        };
    };
    xhr.send(form);
}

// System Info
const bToGB = (b) => { return (b / 1024 / 1024 / 1024).toFixed(1) }
const cpuValue = document.getElementById('cpu-value');
const cpuInfo = document.getElementById('cpu-info');
const memValue = document.getElementById('mem-value');
const memInfo = document.getElementById('mem-info');
const gpuValue = document.getElementById('gpu-value');
const gpuInfo = document.getElementById('gpu-info');
const getSystemInfo = () => {
    fetch('/api/system/info').then(response => response.json()).then(data => {
        cpuValue.innerText = data.cpu.percent.toFixed(2) + '%'
        cpuInfo.innerText = " " + data.cpu.temp + "°C"
        memValue.innerText = data.mem.percent.toFixed(2) + '%'
        memInfo.innerText = " " + bToGB(data.mem.used) + '/' + bToGB(data.mem.total) + "G"
        gpuValue.innerText = data.gpu.util.toFixed(2) + '%'
        gpuInfo.innerText = data.gpu.mem + "% " + data.gpu.temp + "°C"

        if (dialogStatus && dialogTitle.innerText == "Network Error") closeDialog('mini');
    }).catch(e => {
        openDialog("Network Error", "Failed to get system information. Please check the server status.", "mini");
        console.error(e)
    })
    setTimeout(getSystemInfo, 2000)
}; getSystemInfo();

// Open Dialog
let dialogStatus = false;
const dialog = document.getElementById('dialog');
const dialogModal = document.getElementById('dialog-modal');
const dialogTitle = document.getElementById('dialog-title');
const dialogCancel = document.getElementsByClassName('close');
const closeDialog = (size) => {
    dialogStatus = false;
    dialog.classList.remove('active');
    setTimeout(() => {
        if (size != undefined) dialogModal.classList.remove(size);
        dialog.style.visibility = "hidden";
    }, 300)
}
const openDialog = (title, content, size) => {
    dialogStatus = true;
    dialog.style.visibility = "visible";
    dialogTitle.innerText = title;
    dialog.querySelector('.content').innerHTML = content;
    dialog.classList.add('active');
    if (size != undefined) dialogModal.classList.add(size);
    for (let i in dialogCancel) dialogCancel[i].onclick = () => closeDialog(size);
    dialogModal.onclick = (e) => {
        e.stopPropagation();
    }
}

// Stream Sync
const streamSyncBtn = document.getElementById("stream-sync-btn");
streamSyncBtn.onclick = () => {
    streamSyncBtn.classList.add('loading');
    fetch('/api/stream/sync', {
        method: 'POST'
    }).then(response => response.json()).then(data => {
        streamSyncBtn.classList.remove('loading');
        if (data.status == 200) getStream(1)
        else addLog("Sync Error", "red")
    })
}

// Add Stream
const addStreamBtn = document.getElementById("add-stream-btn");
const addStreamDialog = document.getElementById("add-stream-dialog");
const addStreamDialogModal = document.getElementById("add-stream-dialog-modal");
addStreamBtn.onclick = () => {
    addStreamDialog.style.visibility = "visible";
    addStreamDialog.classList.add('active');
    for (let i in dialogCancel) dialogCancel[i].onclick = () => {
        addStreamDialog.classList.remove('active');
        setTimeout(() => addStreamDialog.style.visibility = "hidden", 300)
    }
    addStreamDialogModal.onclick = (e) => {
        e.stopPropagation();
    }
}
const addStreamForm = document.getElementById("add-stream-form");
const addStreamName = document.getElementById("add-stream-name");
const addStreamDes = document.getElementById("add-stream-des");
const addStreamTags = document.getElementById("add-stream-tags");
const addStreamTagsPreview = document.getElementById("add-stream-tags-preview");
const addStreamType = document.getElementById("add-stream-type");
const addStreamImage = document.getElementById("add-stream-image")
const addStreamPreview = document.getElementById("add-stream-preview");
addStreamImage.onchange = () => {
    addStreamPreview.style.display = 'block';
    addStreamPreview.src = URL.createObjectURL(addStreamImage.files[0]);
}
addStreamTags.onkeyup = () => {
    if (addStreamTags.value != "") {
        addStreamTagsPreview.style.display = "block";
        addStreamTagsPreview.innerHTML = "";
        const tags = addStreamTags.value.split(",");
        for (let i = 0; i < tags.length; i++) {
            addStreamTagsPreview.innerHTML += `<div class="ui label">${tags[i]}</div>`
        }
    } else addStreamTagsPreview.style.display = "none";
}
document.getElementById("add-stream-submit").onclick = () => {
    if (addStreamName.value == "" || addStreamDes.value == "" || addStreamTags.value == "" || addStreamType.value == "") {
        openDialog("Error", "Please fill in all fields.", 'mini')
        return
    }
    addStreamForm.classList.add('loading');
    const form = new FormData();
    form.append('name', addStreamName.value);
    form.append('des', addStreamDes.value);
    form.append('tags', addStreamTags.value);
    form.append('type', addStreamType.value);
    form.append('file', addStreamImage.files[0]);
    fetch('/api/stream/add', {
        method: 'POST',
        body: form
    }).then(response => response.json()).then(data => {
        addStreamForm.classList.remove('loading');
        switch (data.status) {
            case 200:
                addLog("Stream added successfully.", "green")
                addStreamDialog.classList.remove('active');
                setTimeout(() => addStreamDialog.style.visibility = "hidden", 300)
                getStream(1);
                break;
            case 500:
                addLog(data.msg, "red")
                openDialog('Error', data.msg);
                break;
            case 404: openDialog('Error', data.msg); break;
        }
    })
}

// Send Command
const commandInput = document.getElementById('command');
const commandBtn = document.getElementById('command-btn');
const sendCommand = () => {
    const command = commandInput.value;
    if (command != "") {
        commandBtn.classList.add('loading');
        commandInput.value = "";
        addLog(command, "green", false, username);
        fetch('/api/system/command', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/x-www-form-urlencoded'
            },
            body: "command=" + command
        }).then(response => response.json()).then(data => {
            commandBtn.classList.remove('loading');
            switch (data.status) {
                case 200:
                    addLog(data.msg);
                    break;
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
        document.getElementById('log').innerHTML += `<div><div class="ui ${logs[i].color} horizontal label"> ${logs[i].time}</div> ${logs[i].text}<div>`
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

// Fan Ctrl
let fanValueNow = 0;
const fanBar = document.getElementById("fan-bar");
const fanValue = document.getElementById("fan-value");
const fanSetSave = document.getElementById("fan-set-save");
const setFanBar = (value) => {
    if (value > 254) value = 254;
    else if (value < 0) value = 0;
    const pt = (value / 254 * 100).toFixed(2) + "%";
    fanValue.innerText = `${pt} (${value}/254)`;
    fanBar.style.width = pt;
    if (value > 203)
        fanBar.style.backgroundColor = "#db2828"
    else if (value > 152)
        fanBar.style.backgroundColor = "#f2c037"
    else if (value > 0)
        fanBar.style.backgroundColor = "#21ba45"
    else
        fanBar.style.backgroundColor = "#888"

    fanValueNow = value;
}
const changeFan = (value, set = false) => setFanBar(set ? value : fanValueNow + value);
fanSetSave.onclick = () => fetch('/api/system/fan', {
    method: 'POST',
    headers: {
        'Content-Type': 'application/x-www-form-urlencoded'
    },
    body: "value=" + fanValueNow
}).then(response => response.json()).then(data => {
    if (data.status == 200) addLog("Fan speed set to " + fanValueNow, "green")
    else addLog(data.msg, "red")
})

// System Init Info
fetch('/api/system/init').then(response => response.json()).then(data => {
    if (data.status == 200) {
        console.log(data)
        document.getElementById("info").onclick = () => openDialog(
            "System Info",
            `<b>CPU: </b>${data.device.cpu}<br/><b>Memory: </b>${bToGB(data.device.mem)}GB<br/><b>GPU: </b>${data.device.gpu}<br/><b>Platform: </b>${data.device.platform}`,
            "mini"
        )
        if (data.fans == -1) {
            document.getElementById("fan-box").innerHTML = `<div class="ui red horizontal label">Not Support</div> Failed to get fan control.`
        } else setFanBar(data.fans);
    } else if (data.status == 403) location = "/login";
    else openDialog("Error", data.msg, 'mini')
})

// Main
if (window.screen.width < 650) {
    openDialog("Warning", "This website is not supported on mobile devices.")
}

getStream(1, undefined, true);