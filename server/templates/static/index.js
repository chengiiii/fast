class FastClient {
  files = []
  el = null
  orders = { 'Name': 1, 'Type': 1, 'Path': 1, 'ModTime': 1, 'Size': 1 }
  constructor(files, el) {
    this.files = files
    this.el = el
  }
  render() {
    this.el.innerHTML = ""
    this.files.forEach(file => {
      const li = document.createElement("li")
      li.innerHTML = `
      <div class="file">
        <a href="${file.Path}">${file.Name}${file.IsDir ? "/" : ""}</a>
        <span class="type">${file.Type}</span>
        <span class="path">${file.Path}</span>
        <span class="time">${formatTime(file.ModTime * 1000)}</span >
        <span class="size">${formatSize(file.Size)}</span>
      </div >
        `
      this.el.appendChild(li.firstElementChild)
    })
  }
  sortBy(filed) {
    const order = this.orders[filed]
    this.files.sort((a, b) => {
      if (a[filed] > b[filed]) {
        return -order
      } else if (a[filed] < b[filed]) {
        return order
      } else {
        return 0
      }
    })
    this.orders[filed] = -order
    this.render()
  }
}

function back() {
  window.location.href = window.location.pathname.split("/").slice(0, -1).join("/") || "/";
}

// format timestamp to 2021-01-01 52:17
function formatTime(timestamp) {
  const date = new Date(timestamp)
  return `${date.getFullYear()}-${date.getMonth() + 1}-${date.getDate()} ${date.getHours()}:${date.getMinutes()}`
}

// format file size form bytes to KB, MB, GB
function formatSize(bytes) {
  if (bytes < 1024) {
    return bytes + "B"
  } else if (bytes < 1024 * 1024) {
    return (bytes / 1024).toFixed(2) + "KB"
  } else if (bytes < 1024 * 1024 * 1024) {
    return (bytes / 1024 / 1024).toFixed(2) + "MB"
  } else {
    return (bytes / 1024 / 1024 / 1024).toFixed(2) + "GB"
  }
}

function post(url, data) {
  return fetch(url, {
    method: "POST",
    body: data
  })
}

const dropArea = document.querySelector("#drop-area");
dropArea && dropArea.addEventListener("drop", function drop(event) {
  event.preventDefault()
  const files = event.dataTransfer.files
  const formData = new FormData()
  for (let i = 0; i < files.length; i++) {
    formData.append("files", files[i])
  }

  post(window.location.href, formData).then(res => {
    console.log("upload success", res)
    window.location.reload()
  })

  dropArea.style.display = "none"
})

window.addEventListener("dragover", function (e) {
  e.preventDefault()
  if (!dropArea) return
  dropArea.style.display = "block"
})

dropArea && dropArea.addEventListener("dragleave", function (e) {
  e.preventDefault()
  dropArea.style.display = "none"
})