# AMP Console

To view the console, hit `/ui` off of the master switch in your browser.  
`https://localhost:7467/ui`  
*Alternatively, you can also pop `index.html` into a browser with cross-origin allowed.*

## Setup
The `AppData.json` file under the `/static/js/` directory contains all available configurations to modify the console at startup.  There shouldn't be a need to change any configurations, but if you aren't running a stock setup (maybe you're "hosting" off of your filesystem) you may have to modify the endpoint info in this file.

#### Changing the switch endpoint
```js
    "endpoints": {
        "switchesIPAddr": "localhost", // change this to the IP of the machine running the switches
        "masterPort": "7467" // change this to the master switch's port
    }
```
Additionally, the `AppData.json` file contains a session object (placeholder for later) that persists user settings like locale, timezones, initial lat/lng on maps, etc...  You can modify these settings if you so desire.

## Architecture
The console is, for now, hosted on each switch, so you can access it from any.  However, the console's current node exploration process only searches down.  This means that if the console is launched off of a child node, any ancestor nodes will not be discovered.

## Troubleshooting
There may be an issue displaying switch data if the browser blocks the health request because of the self-signed cert presented by the server.  If this is the case, you'll have to add securty exceptions for each hosted switch, or disable the browser's web security.

## Further Info
The console is an ongoing effort, and if you have any suggestions for features, layouts, etc, just add to the following google doc:  
[https://drive.google.com/open?id=1sWSTXZte3_k8NX6D2oelXRSvXHokMjZwIS3SHHKuS6I](https://drive.google.com/open?id=1sWSTXZte3_k8NX6D2oelXRSvXHokMjZwIS3SHHKuS6I)
