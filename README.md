# export-komoot

This is a proof-of-concept which allows you to export your planned tours from [Komoot](https://www.komoot.com).

Note that this is a unofficial tool which uses private API's from Komoot and can break at any time…

# Setup

Create a `.env` file which should include your username and password:

```env
KOMOOT_EMAIL=user@host.com
KOMOOT_PASSWD=password
```

# Running a full export

Run: `make run-no-incremental`

# Running an incremental export

Run: `make run`

# Usage

```
$ ./export-komoot -h
Usage of ./export-komoot:
  -concurrency int
        The number of simultaneous downloads (default 16)
  -email string
        Your Komoot email address
  -filter string
        Filter on the given name
  -format string
        The format to export as: gpx or fit (default "gpx")
  -no-incremental
        If specified, all data is redownloaded
  -password string
        Your Komoot password
  -to string
        The path to export to
```

# Caution

Use at your own risk!