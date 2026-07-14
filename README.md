# PropaneBot
## What is this?
This program monitors weight readings from an MQTT server and does three things:
* Provides a Discord bot (`/weight`) to show the current weight and percentage remaining (-ish). This is configured in `cylinder.json` and has to be adjusted every time the cylinder is replaced (because they don't always have the same tare or fill weights).
* Provide a web server to display the weight and amount remaining. This is used by a RPI Zero W that shows the page in kiosk mode on a screen in the Hot Metals area.
* A background thread monitors the weight and after it drops below a certain percentage will notify a specific user in a specific channel (set in `config.json`). This is meant to serve as a reminder to said person that maybe they should think about putting in a call to the gas supplier.

## How to run it as a container
```
docker build -t propane-bot .
docker run -d --name propanebot \
  --log-driver=local \
  --restart unless-stopped \
  --network host \
  -v $(pwd)/config.json:/app/config.json:ro \
  -v $(pwd)/cylinder.json:/app/cylinder.json \
  propanebot
```
Note that the `--network host` option is required for the bot to be able to connect to the MQTT server and for the web server to be accessible on the local network. Also, make sure to adjust the paths to `config.json` and `cylinder.json` as needed.
