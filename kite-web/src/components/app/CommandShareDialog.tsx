const PREFIX = "LemonCube's Kite Command Share";

// SENDER SCRIPT 
export const generateCommandShare = () => {
    if (typeof window.fetch !== 'function') {
        alert('Error: Fetch not available.');
        return;
    }

    const originalFetch = window.fetch;
    window.fetch = function(url, options) {
        if (options?.method?.toUpperCase() === 'PATCH' && options.body) {
            const body = options.body.toString();
            navigator.clipboard.writeText(body).then(() => {
                alert('SUCCESS: Sharing code copied to clipboard!');
            }).catch(() => {
                console.log("Payload:", body);
                alert('Copy failed. Check console.');
            });
        }
        return originalFetch.apply(this, arguments as any);
    };
    
    alert(PREFIX + ' Sender active! Click "Save Changes" to generate code.');
};

// RECEIVER SCRIPT
export const receiveCommandShare = () => {
    let capturedData = { name: null as string | null, description: null as string | null };
    let isMonitoring = false;

    const sendPatchRequest = (appId: string, commandId: string, name: string, description: string, payload: string) => {
        try {
            const parsed = JSON.parse(payload);
            parsed.flow_source.nodes = parsed.flow_source.nodes.map((node: any) => {
                if (node.type === 'entry_command') {
                    node.data.name = name;
                    node.data.description = description;
                }
                return node;
            });

            fetch(`https://api.kite.onl/v1/apps/${appId}/commands/${commandId}`, {
                method: "PATCH",
                headers: { "accept": "application/json", "Content-Type": "application/json" },
                body: JSON.stringify(parsed),
                credentials: 'include'
            }).then(res => {
                if (res.ok) alert('Success! Reload to see changes.');
            });
        } catch (e) {
            console.error(PREFIX + " Error:", e);
        }
    };

    const urlMatch = window.location.pathname.match(/\/apps\/([^\/]+)/);
    if (!urlMatch) {
        alert('Please run this on the Kite apps page.');
        return;
    }
    const appId = urlMatch[1];
    
    const storedPayload = prompt(PREFIX + " Enter sharing code:");
    if (!storedPayload) return;

    const originalFetch = window.fetch;
    window.fetch = function(url, options) {
        const urlStr = url.toString();
        if (options?.method?.toUpperCase() === 'POST' && urlStr.includes('/commands') && options.body) {
            try {
                const parsedBody = JSON.parse(options.body.toString());
                const entry = parsedBody.flow_source.nodes.find((n: any) => n.type === 'entry_command');
                if (entry?.data.name) {
                    capturedData = { name: entry.data.name, description: entry.data.description || "" };
                    isMonitoring = true;
                    return originalFetch.apply(this, arguments as any).then((res: any) => {
                        setTimeout(() => {
                            const newId = window.location.pathname.split('/').pop();
                            if (newId) sendPatchRequest(appId, newId, capturedData.name!, capturedData.description!, storedPayload);
                        }, 800);
                        return res;
                    });
                }
            } catch (e) {}
        }
        return originalFetch.apply(this, arguments as any);
    };

    alert('Receiver active! Now create a new command to load the shared code.');
};
