function CreateServerDOM(id,name,pfp)
{
    let serverBrowser = document.getElementById("serverBrowser")
    let serverNode = document.createElement("div")
    serverNode.className = "serverIcon"
    serverNode.style.backgroundImage = `url(${pfp})`
    serverNode.dataset.name = name
    serverNode.dataset.serverID = id
    serverNode.addEventListener("click", function(){
                 navigate(`/server/${id}`)
             });
    serverBrowser.appendChild(serverNode)
}