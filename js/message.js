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

function getServers() {
    let JWTCookie = getCookie("RelayJWT")
    GET(JWTCookie,
        serverAddress + getServerEndpoint)
    
}