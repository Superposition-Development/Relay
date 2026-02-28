async function signup() {
    let payload =
    {
        "userID": "user",
        "username": "username",
        "password": "password",
        "pfp": "pfpbase64"
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

async function login() {
    let payload =
    {
        "userID": "user",
        "password": "password",
    }

    data = await POST(payload,
        null,
        serverAddress + loginEndpoint
    )

    if(data["data"] != null)
    {
        setCookie("RelayJWT",data["data"]["RelayJWT"],60)
    }
}