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
