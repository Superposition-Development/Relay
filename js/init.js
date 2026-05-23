async function boot()
{
    initServers = await getServers()
    for (const server of initServers.data) {
        CreateServerDOM(server.id,server.name,server.pfp)
    }
}

boot()