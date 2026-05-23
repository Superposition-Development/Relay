function router() {
    const hash = window.location.hash.slice(1); // remove #

    const parts = hash.split("/");

    if (parts[1] === "server") {
        const serverId = parts[2];
        console.log("server:", serverId);
    }

    if (parts[1] === "channels") {
        const type = parts[2];
        const channelId = parts[3];
        console.log(type, channelId);
    }
}

window.addEventListener("hashchange", router);
router();

function navigate(path) {
    window.location.hash = path;
    router();
}




// navigate("/server/1");
