function POST(payload,authKey,address)
{
    result = {"data":null,
              "error":null
    }

    fetch(address,
        {
            method: "POST",
            headers: {
                "Content-Type": "application/json",
                'Authorization': `Bearer ${authKey}`,
            },
            credentials:"include",
            body: JSON.stringify(payload)
        }).then(response => {
            if (response.ok) {
                return response.json()
            }
            throw new Error("Network response failed")
        }).then(data => {
            result["data"] = data
        })
        .catch(error => {
            result["error"] = error
        });

    return result 
}

function GET(authKey,address)
{
    result = {"data":null,
              "error":null
    }

    fetch(address,
        {
            method: "GET",
            headers: {
                "Content-Type": "application/json",
                'Authorization': `Bearer ${authKey}`,
            },
            credentials:"include"
        }).then(response => {
            if (response.ok) {
                return response.json()
            }
            throw new Error("Network response failed")
        }).then(data => {
            result["data"] = data
        })
        .catch(error => {
            result["error"] = error
        });

    return result 
}