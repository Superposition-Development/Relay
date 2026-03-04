async function POST(payload, authKey, address) {
    try {
        const response = await fetch(address, {
            method: "POST",
            headers: {
                "Content-Type": "application/json",
                "Authorization": `Bearer ${authKey}`,
            },
            credentials: "include",
            body: JSON.stringify(payload)
        });

        if (!response.ok) {
            throw new Error("Network response failed");
        }

        const data = await response.json();
        return { data: data, error: null };

    } catch (error) {
        return { data: null, error: error };
    }
}

async function GET(authKey, address) {
    try {
        const response = await fetch(address, {
            method: "GET",
            headers: {
                "Content-Type": "application/json",
                "Authorization": `Bearer ${authKey}`,
            },
            credentials: "include",
        });

        if (!response.ok) {
            throw new Error("Network response failed");
        }

        const data = await response.json();
        return { data: data, error: null };

    } catch (error) {
        return { data: null, error: error };
    }
}

//note that websocket addresses are just "ws://[server address just without http://]"
//and given https exists wss also does, could be complicated cause localhost uses wss and i
//dont know too much abt port forwarding or wtv when ppl go to deploy their own

/*

websocket endpoint docs until we get a better oen

client->server

{
 clause:sendMessage,
 content:{}
}


server->client
{
 clause:recieveMessage,
 content:{}
}

*/

let socket = null

function registerWebsocket(address) {
    socket = new WebSocket(address)
    socket.addEventListener("open", (e) => {
        // socket.send("data")
    })

    socket.addEventListener("message", (event) => {
        console.log("Message from server:", event.data);
        console.log(JSON.parse(event.data))
    });

    socket.addEventListener("close", () => {
        console.log("Connection closed");
    });

    socket.addEventListener("error", (error) => {
        console.error("WebSocket error:", error);
    });
}

/*
expected data

{
 message:caseType,
 authKey:JWT,
 content: (everything else is dependent on what the endpoint wants, just make message:a valid casetype)
}
*/
function sendWebsocketJSON(message)
{
    message = JSON.stringify(message)
    if(socket == null || socket.readyState != WebSocket.OPEN)
    {
        console.error("Websocket offline") //dont do it like this find some modal or smth
        return
    }
    socket.send(message)
}

