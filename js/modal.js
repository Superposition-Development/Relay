function setModalHeader(title)
{
    document.getElementById("modalTitle").innerText = title
}

function createModalPictureUpload()
{

}

function createServerModal()
{
    createModal(() => setModalHeader("title"),createModalPictureUpload,)
}

function createModal(...modules) {
    let modalWrapper = document.getElementById("modalWrapper")
    modalWrapper.style.filter = "opacity(0%)"
    document.getElementById("modalWrapper").style.display = "flex"
    window.setTimeout(function () {
        modalWrapper.style.filter = "opacity(100%)"
    }, 20)

    for (let module of modules) {
        module();
    }
}

/*createModal(
    () => setModalHeader("title")
);
*/

function closeModal() {
    let modalWrapper = document.getElementById("modalWrapper")
    modalWrapper.style.filter = "opacity(0%)"
    window.setTimeout(function () {
        modalWrapper.style.display = "none"
        document.getElementById("modalContent").innerHTML = ""
    }, 1000)
}