function createModal() {
    let modalWrapper = document.getElementById("modalWrapper")
    document.getElementById("modalWrapper").style.display = "flex"
    modalWrapper.style.filter = "opacity(100%)"
}

function closeModal() {
    let modalWrapper = document.getElementById("modalWrapper")
    modalWrapper.style.filter = "opacity(0%)"
    window.setTimeout(function () {
        modalWrapper.style.display = "none"
        modalWrapper.innerHTML = ""
    }, 1000)
}