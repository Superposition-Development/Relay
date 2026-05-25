async function boot() {

    if(getCookie("RelayJWT") == null)
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
        
    }
}

boot()