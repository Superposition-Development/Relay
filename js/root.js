async function boot()
{
    const params = new URLSearchParams(window.location.search);
    const serverID = params.get("serverID"); // "John"
    const channelID = params.get("channelID"); // "30"

    initServers = await getServers()
    for (const server of initServers.data) {
        CreateServerDOM(server.id,server.name,server.pfp)
    }


}

boot()