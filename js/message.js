function createServer() {
    payload = {
        "name": "server name",
        "pfp": "too much work"
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
    console.log(res)
}

function createChannel() {
    payload = {
        "name": "server name",
        "serverID": 1
    }
    let JWTCookie = getCookie("RelayJWT")
    POST(payload,
        JWTCookie,
        serverAddress + createChannelEndpoint)
}


async function getChannels() {
    payload = {
        "serverID": 1
    }
    let JWTCookie = getCookie("RelayJWT")
    let res  = await POST(payload,
        JWTCookie,
        serverAddress + getChannelEndpoint)
    console.log(res)
}