# Currency-trends-monitor
**What's this?** This is a project to test Go Lang's capabilities.


## Start service for development:
Fetch the project: <br> ``` git clone https://github.com/BuzinD/currency-trends.git ```
For the first running: 
- prepare variabls in app/env/okx.env
use a command: ` make first-run-dev` It build docker images for DB, migrations also run up DB, migrations, app

## Available commands
For getting info about available commands run: <br>
`make help`

## Features of project:
### data-fetcher
- Getting available currencies (scheduled twice a day)
- Getting trade history candles (scheduled on every hour)
- Getting real-time trade info