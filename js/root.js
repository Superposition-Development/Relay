async function boot() {

    if(getCookie("RelayJWT") == null)
    {
        window.location.href = "index.html"
    }

    const jwtStatus = await isUserLoggedIn()
    if(jwtStatus.data.status != "ok")
    {
        window.location.href = "index.html"
    }

    const params = new URLSearchParams(window.location.search);
    const serverID = params.get("serverID"); // "John"
    const channelID = params.get("channelID"); // "30"

    initServers = await getServers()
    for (const server of initServers.data) {
        CreateServerDOM(server.id, server.name, server.pfp)
    }

    if (serverID != null) {
        initChannels = await getChannels(serverID)
        for (const channel of initChannels.data) {
            CreateChannelDOM(channel.id,channel.name)
        }
    }

    if(channelID != null && serverID != null)
    {
        initMessages = await getMessages(serverID,channelID,"0",false,true)
        console.log(initMessages)
        for (const message of initMessages.data) {
            CreateMessageDOM(message.id, message.name, message.pfp,message.content,message.timestamp,true)
        }
    }
}

boot()