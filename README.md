# go-cowin-bot
A cowin bot that will give you an update  on discord whenever it finds a vacancy for the parameters provided

## Setup: 
* download go-cowin-bot executable from the releases 
* for discord bot updates:
  * make a discord bot (https://discord.com/developers/applications)
  * copy bot token and set it in environment variable (`export DISCORD_TOKEN=<>`)
  * get Channel ID for the channel on which bot should send updates (`export CHANNEL_ID=<>`)
    * open discord.com in browser, navigate to desired channel, copy last section of URL
  * make sure bot has "send embedded links" permission on channel
  

## Usage:
```
./go-cowin-bot
  -cmd     [enables output on cmd line (disbales discord)]
  -discord [sends updates to discord channel]
  
  -district_id [district_id that needs to be checked: default = 294(BBMP)]
  -age         [minimum age check: default=18]
  -poll        [amount of time between polls to API: default=15]
  -days        [number of days to check for: default = 10]
```
