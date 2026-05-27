async function boot() {

    // console.log("bruh")
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
    prepareDOM()

}

async function prepareDOM()
{
    const params = new URLSearchParams(window.location.search);
    const serverID = params.get("serverID"); // "John"
    const channelID = params.get("channelID"); // "30"

    if (serverID != null) {
        initChannels = await getChannels(serverID)
        ClearChannelDOM()
        for (const channel of initChannels.data) {
            CreateChannelDOM(channel.id,channel.name)
        }
    }

    if(channelID != null && serverID != null)
    {
        initMessages = await getMessages(serverID,channelID,"0",false,true)
        for (const message of initMessages.data) {
            const epochSeconds = message.timestamp;
            const date = new Date(epochSeconds * 1000);
            //local time string date.toString()
            //utc time date.toUTCString()
            //locale specific date.toLocalestring()
            CreateMessageDOM(message.id, message.name, message.pfp,message.content," "+date.toLocaleString(),true)
        }
    }
}

boot()
window.addEventListener('navigate', (e) => {
    // console.log("navigate",e.detail);
    prepareDOM()
});