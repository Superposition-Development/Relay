function test() {
    let payload =
    {
        "userID":"user",
        "username":"username",
        "password":"password",
        "pfp":"pfpbase64"
    }
    fetch(serverAddress + signupEndpoint,
        {
            method: "POST",
            headers: {
                "Content-Type": "application/json",
                // 'Authorization': `Bearer ${localStorage.getItem("jwt")}`
            },
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