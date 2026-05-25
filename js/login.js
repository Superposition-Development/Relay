async function signup(userID,username,password,pfp) {
    let payload =
    {
        "userID": userID,
        "username": username,
        "password": password,
        "pfp": pfp
    }

    data = await POST(payload,
        null,
        serverAddress + signupEndpoint
    )

    console.log(data)

    if(data["data"] != null)
    {
        setCookie("RelayJWT",data["data"]["RelayJWT"],60)
    }
}

async function login(userID,password) {
    let payload =
    {
        "userID": userID,
        "password": password,
    }

    data = await POST(payload,
        null,
        serverAddress + loginEndpoint
    )

    console.log(data)

    if(data["data"] != null)
    {
        setCookie("RelayJWT",data["data"]["RelayJWT"],60)
    }
}

async function isUserLoggedIn()
{
    let JWTCookie = getCookie("RelayJWT")
    let res = await GET(JWTCookie,
        serverAddress + validateUserEndpoint)
    return res
}