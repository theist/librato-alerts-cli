# librato-alerts

Small commandline client to enable and disable alerts in librato legacy 
accounts.

Usage: ` librato-alerts [help | disable | enable | list | status | recent]`

`enable` and `disable` requires a list of alerts to disable passed by standard 
input thru a pipe, the output of `list` can be used for this purpose like this:
```
   librato-alerts list | grep <pattern> | librato-alerts disable
```
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

 * It does not support pagination yet if there are more alerts than the ones 
   which fits in an API call it will not list them.
 * This is tested against an old, no tagged metrics librato account may work
   in the modern ones.
