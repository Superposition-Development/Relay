function navigate(params) {
    const url = new URL(window.location);

    Object.keys(params).forEach(key => {
        url.searchParams.set(key, params[key]);
    });

    history.replaceState({}, "", url);
    router();
}

function router() {
    const hash = window.location.hash.slice(1);

    const parts = hash.split("/");
    //parts[0...n]
}
function CreateServerDOM(id,name,pfp)
{
    let serverBrowser = document.getElementById("serverBrowser")
    let serverNode = document.createElement("div")
    serverNode.className = "serverIcon"
    serverNode.style.backgroundImage = `url(${pfp})`
    serverNode.dataset.name = name
    serverNode.dataset.serverID = id
    serverNode.addEventListener("click", function(){
                 navigate({"serverID":id})
             });
    serverBrowser.appendChild(serverNode)
}