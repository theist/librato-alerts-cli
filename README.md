# librato-alerts

Small commandline client to enable and disable alerts in librato legacy 
accounts.

Usage: ` librato-alerts [help | disable | enable | list | status | recent]`

`enable` and `disable` requires a list of alerts to disable passed by standard 
input thru a pipe, the output of `list` can be used for this purpose like this:
```
   librato-alerts list | grep <pattern> | librato-alerts disable
```

## CONFIGURATION

This requires two environment varables to store the librato credentials, 
`LIBRATO_MAIL` with the librato user's mail and `LIBRATO_TOKEN`
with a valid librato API token. API token must have read / write access to allow update alarms state.
The environment variables can also be placed in an `.env` file.

## MODES

```
   list:    List all alerts, telling if they are enabled or disabled.
   status:  Lists the alert names which are in alarm state.
   recent:  Lists the alert names of alert which were resolved recently.
   enable:  Enable alerts passed by stdin. Alerts must be pased one by line,
            and it will be updated only if they are disabled
   disable: Disable alerts passed by stdin. Alerts must be pased one by line,
            and it will be updated only if they are enabled
   help:    This help.
```

## ALMOST KNOWN BUGS or TODO's:

 * This is tested against an old, no tagged metrics librato account may work
   in the modern ones.
