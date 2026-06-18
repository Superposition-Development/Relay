function navigate(params) {
    const url = new URL(window.location);
    url.searchParams.delete("channelID")

    Object.keys(params).forEach(key => {
        url.searchParams.set(key, params[key]);
    });

    history.replaceState({}, "", url);
    window.dispatchEvent(new CustomEvent("navigate",{detail:"yeehaw"}))
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
    serverBrowser.prepend(serverNode)
}

function ClearChannelDOM()
{
    document.getElementById("channels").innerHTML = ""
}

function CreateChannelDOM(id,name)
{
    let channelHolder = document.getElementById("channels")
    let channelNode = document.createElement("div")

    /*
    <div class="channel">
          <span class="channelName">channel NAme</span>
        </div>
    */

    channelNode.className = "channel"
    channelNode.dataset.channelID = id
    channelNode.addEventListener("click", function(){
                 navigate({"channelID":id})
             });
    channelHolder.appendChild(channelNode)
    let channelNameNode = document.createElement("span")
    channelNameNode.innerText = name
    channelNode.appendChild(channelNameNode)
}

function CreateMessageDOM(id,name,pfp,content,timestamp,shouldPrepend)
{
    let messageField = document.getElementById("messageField")
    let messageNode = document.createElement("div")
    messageNode.className = "message"
    messageNode.dataset.messageID = id
    if(shouldPrepend)
    {
        messageField.prepend(messageNode)
    }
    else
    {
        messageField.appendChild(messageNode)
    }

    let messagePFP = document.createElement("div")
    messagePFP.className = "messagePFP"
    messagePFP.style.backgroundImage = `url(${pfp})`
    messageNode.appendChild(messagePFP)

    let messageContentWrapper = document.createElement("div")
    messageContentWrapper.className = "messageContentWrapper"
    messageNode.appendChild(messageContentWrapper)

    let messageUsername = document.createElement("span")
    messageUsername.className = "messageUsername"
    messageUsername.innerText = name
    messageContentWrapper.appendChild(messageUsername)

    let messageDate = document.createElement("span")
    messageDate.className = "messageDate"
    messageDate.innerText = timestamp
    messageUsername.appendChild(messageDate)

    let messageContent = document.createElement("span")
    messageContent.className = "messageContent"
    messageContent.innerText = content
    messageContentWrapper.appendChild(messageContent)
}
