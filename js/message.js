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
        "content":content,
        "authKey":getCookie("RelayJWT"),
        "message":"sendMessage"
    }
    let rest = await sendWebsocketJSON(payload)
    // let res  = await POST(payload,
    //     JWTCookie,
    //     serverAddress + sendMessageEndpoint)
    // console.log(res)
}

/*MoreThan  string `json:"moreThan"`
	ChannelID string `json:"channelID"`
	ServerID  string `json:"serverID"`
	MessageID string `json:"messageID"`
	Ascending string `json:"ascending"`*/
async function getMessages(serverID,channelID,messageID,ascending,moreThan)
{
    payload = {
        "serverID": serverID,
        "channelID":channelID,
        "messageID":messageID,
        "ascending":ascending.toString(),
        "moreThan":moreThan.toString()
    }
    let JWTCookie = getCookie("RelayJWT")
    let res  = await POST(payload,
        JWTCookie,
        serverAddress + getMessagesEndpoint)
    return res
}

//eventually move this type of stuff into like a "control.js" or smth
document.addEventListener("keypress",function(e){

    //don't do this!! load this from an external file somewhere!!!
    const params = new URLSearchParams(window.location.search);
    const serverID = params.get("serverID"); // "John"
    const channelID = params.get("channelID"); // "30"
    //end segment of bad practice
    if(e.key == "Enter")
    {
        if(serverID == null || channelID == null)
        {
            return;
        }
        if(!e.shiftKey)
        {
            e.preventDefault()
            let content = document.getElementById("chatTextarea").value
            sendMessage(serverID,channelID,content)
            document.getElementById("chatTextarea").value = ""
        }
    }
})

function shouldAutoScroll()
{
    return messageField.scrollHeight - messageField.scrollTop - messageField.offsetHeight < 60
}

function scrollToLatestMessage()
{
    let messageField = document.getElementById("messageField")
    messageField.scrollTo(0,messageField.scrollHeight)
}

