const stunServers = { iceServers: [{ url: 'stun:stun.l.google.com:19302' }] }
let localConnection;

async function joinCall()
{
    const connAttempt = new RTCPeerConnection(stunServers)
    connAttempt.addTransceiver("video")
    connAttempt.addTransceiver("audio")
    const localOffer = await connAttempt.createOffer()
    await connAttempt.setLocalDescription(localOffer)

    connAttempt.onicecandidate = (e) => {
        if (e.candidate)
        {
            //send smth to the SFU
        }
    }
    localConnection = connAttempt
    //send smethh to the SFU via localOffer.sdp
}

//brb, reminder ot add smth abt globalPeer.addIceCandidate(new RTCIceCandidate())
//make the websocket fire an event and add event listeners across the files 

document.addEventListener("WebsocketMessage",function(e)
{
    console.log(e)
    switch(e["message"])
    {
        
    }
})