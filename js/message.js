function signup() {
    let payload =
    {
        "userID": "user",
        "username": "username",
        "password": "password",
        "pfp": "pfpbase64"
    }
    fetch(serverAddress + signupEndpoint,
        {
            method: "POST",
            headers: {
                "Content-Type": "application/json",
            },
            body: JSON.stringify(payload)
        }).then(response => {
            if (response.ok) {
                return response.json()
            }
            throw new Error("Network response failed")
        }).then(data => {
            setCookie("RelayJWT",data["RelayJWT"],60)
        })
        .catch(error => {
            console.error("There was a problem with the fetch", error);
        });
}

function login() {
    let payload =
    {
        "userID": "user",
        "password": "password",
    }
    fetch(serverAddress + loginEndpoint,
        {
            method: "POST",
            headers: {
                "Content-Type": "application/json",
            },
            body: JSON.stringify(payload)
        }).then(response => {
            if (response.ok) {
                console.log(response)
                return response.json()
            }
            throw new Error("Network response failed")
        }).then(data => {
            console.log(data)
            setCookie("RelayJWT",data["RelayJWT"],60)
        })
        .catch(error => {
            console.error("There was a problem with the fetch", error);
        });
}

function createServer() {
    payload = {
        "name": "server name",
        "pfp": "too much work"
    }
    let JWTCookie = getCookieByName("RelayJWT")
    fetch(serverAddress + createServerEndpoint,
        {
            method: "POST",
            headers: {
                "Content-Type": "application/json",
                'Authorization': `Bearer ${JWTCookie}`,
            },
            credentials:"include",
            body: JSON.stringify(payload)
        }).then(response => {
            if (response.ok) {
                return response.json()
            }
            throw new Error("Network response failed")
        }).then(data => {
            console.log(data)
        })
        .catch(error => {
            console.error("There was a problem with the fetch", error);
        });
}