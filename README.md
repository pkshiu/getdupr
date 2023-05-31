# DUPR Rating data downloader in golang

This DUPR data downloader pulls player and ratings data for all players belonging
to a club.

Usage:

```
    # To run with a full go installation:
    go run main.go -username <your_dupr_username> -password <your_dupr_password> -club_id <your club id>

    # To run just the binaries downloaded, e.g. on a windows machine
    getdupr.exe  -username <your_dupr_username> -password <your_dupr_password> -club_id <your club id>

    # You can copy getdupr.exe to wherever you need, or invoke it with the full directory
    c:\\where-you-put-the-file/getdupr.exe -username <your_dupr_username> -password <your_dupr_password> -club_id <your club id>
```

## History

0.1 Initial release, much more cleanup and test is needed

## Logging In

DUPR requires a login operation to access the data. You can specify your
DUPR username and password on the command line, or save it in a `.env` file.
Unless your computer is private and secure, you should not store your password
in the local `.env` file.

## Club Id

This downloader uses a DUPR Club ID to pull the set of members/players.
You can find the clubs ID in the club information page's URL.
Your own clubs are listed in your dashboard. You can search for other clubs in the clubs page.

## Implementation Note

I converted a more full feature tool in Python to this version in golang to learn golang.
More features will be added at a later stage. The dupr client will be moved to a
separate package later.

## Executable

The justfile builds executable for my M1 mac, and for amd64 windows. The binaries
are placed in the `.bin` subdirectory.


