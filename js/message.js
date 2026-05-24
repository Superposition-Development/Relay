function createServer(serverName, pfp) {
    payload = {
        "name": serverName,
        "pfp": pfp
    }
    let JWTCookie = getCookie("RelayJWT")
    POST(payload,
        JWTCookie,
        serverAddress + createServerEndpoint)
}

async function getServers() {
    let JWTCookie = getCookie("RelayJWT")
    let res = await GET(JWTCookie,
        serverAddress + getServerEndpoint)
    return res
}

async function createChannel(channelName, serverID) {
    payload = {
        "name": channelName,
        "serverID": serverID
    }
    let JWTCookie = getCookie("RelayJWT")
    await POST(payload,
        JWTCookie,
        serverAddress + createChannelEndpoint)
}


async function getChannels(serverID) {
    payload = {
        "serverID": serverID
    }
    let JWTCookie = getCookie("RelayJWT")
    let res  = await POST(payload,
        JWTCookie,
        serverAddress + getChannelEndpoint)
    return res
}

async function joinServer(serverID)
{
    payload = {
        "serverID": serverID
    }
    let JWTCookie = getCookie("RelayJWT")
    let res  = await POST(payload,
        JWTCookie,
        serverAddress + joinServerEndpoint)
    console.log(res)
}

async function sendMessage(serverID,channelID,content)
{
    payload = {
        "serverID": serverID,
        "channelID":channelID,
        "content":content
    }
    let JWTCookie = getCookie("RelayJWT")
    let res  = await POST(payload,
        JWTCookie,
        serverAddress + sendMessageEndpoint)
    console.log(res)
}

