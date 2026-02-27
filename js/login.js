function signup() {
    let payload =
    {
        "userID": "user",
        "username": "username",
        "password": "password",
        "pfp": "pfpbase64"
    }

    data = POST(payload,
        null,
        serverAddress + signupEndpoint
    )

    if(data["data"] != null)
    {
        setCookie("RelayJWT",data["data"]["RelayJWT"],60)
    }
}

function login() {
    let payload =
    {
        "userID": "user",
        "password": "password",
    }

    data = POST(payload,
        null,
        serverAddress + loginEndpoint
    )

    if(data["data"] != null)
    {
        setCookie("RelayJWT",data["data"]["RelayJWT"],60)
    }
}