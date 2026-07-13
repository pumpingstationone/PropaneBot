# PropaneBot
This program monitors weight readings from an MQTT server and does three things:
* Provides a Discord bot (`/weight`) to show the current weight and percentage remaining (-ish). This is configured in `cylinder.json` and has to be adjusted every time the cylinder is replaced (because they don't always have the same tare or fill weights).
* Provide a web server to display the weight and amount remaining. This is used by a RPI Zero W that shows the page in kiosk mode on a screen in the Hot Metals area.
* A background thread monitors the weight and after it drops below a certain percentage will notify a specific user in a specific channel (set in `config.json`). This is meant to serve as a reminder to said person that maybe they should think about putting in a call to the gas supplier.


