# export-komoot

This is a proof-of-concept which allows you to export your planned tours from [Komoot](https://www.komoot.com).

Note that this is a unofficial tool which uses private API's from Komoot and can break at any time…

# Setup

Create a `.env` file which should include your username and password:

```env
KOMOOT_EMAIL=user@host.com
KOMOOT_PASSWD=password
KOMOOT_USER_ID=123456
```

# Finding your Komoot User ID

To find your Komoot user ID, login to Komoot, click on your user name in the upper right corner of the screen and
select the option "Profile". The URL you will navigate to will look like this:

```
https://www.komoot.com/<lang>>-<localle>/user/<userid>
```

Your user ID is the number in the last part of the URL.

# Running a full export

Run: `make run-full`

# Running an incremental export

Run: `make run-incremental`

# Usage

```
$ ./export-komoot -h
Usage: export-komoot --email EMAIL --password PASSWORD --userid USERID [--filter FILTER] [--format FORMAT] [--to TO] [--fulldownload] [--concurrency CONCURRENCY] [--tourtype TOURTYPE]

Options:
  --email EMAIL          Your Komoot email address
  --password PASSWORD    Your Komoot password
  --userid USERID        Your Komoot user ID
  --filter FILTER        Filter tours with name matching this pattern
  --format FORMAT        The format to export as: gpx or fit [default: gpx]
  --to TO                The path to export to
  --fulldownload         If specified, all data is redownloaded [default: false]
  --concurrency CONCURRENCY
                         The number of simultaneous downloads [default: 16]
  --tourtype TOURTYPE    The type of tours to download [default: tour_planned]
  --help, -h             display this help and exit
```

# Caution

Use at your own risk!

# References

https://github.com/Woeler/komoot-php/commit/21065fcf517cc0fac646a6a216b5cf2d851f7975#diff-17339dceedd73393b090f1db8e636e6a8a5a161944c87d85dcd8ec3789dd6112