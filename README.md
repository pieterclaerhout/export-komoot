# export-komoot

This is a tool which allows you to export your planned and recorded tours from [Komoot](https://www.komoot.com).

> [!NOTE]
> This is an unofficial tool which uses private API's from Komoot and can break at any time…

# Installing

To install, you can download the latest release from [the releases](https://github.com/pieterclaerhout/export-komoot/releases).

> [!NOTE]
> You don't need to have Goland installed when you download the binaries from the release page.

# Finding your Komoot User ID

To find your Komoot user ID, login to Komoot, click on your user name in the upper right corner of the screen and
select the option "Profile". The URL you will navigate to will look like this:

```
https://www.komoot.com/<lang>>-<localle>/user/<userid>
```

Your user ID is the number in the last part of the URL.

# Usage

## Getting the help info

```
$ ./export-komoot -h
Usage: export-komoot --email EMAIL --password PASSWORD --userid USERID [--filter FILTER] --to TO [--fulldownload] [--concurrency CONCURRENCY] [--tourtype TOURTYPE]

Options:
  --email EMAIL          Your Komoot email address [env: KOMOOT_EMAIL]
  --password PASSWORD    Your Komoot password [env: KOMOOT_PASSWD]
  --userid USERID        Your Komoot user ID [env: KOMOOT_USER_ID]
  --filter FILTER        Filter tours with name matching this pattern
  --to TO                The path to export to
  --fulldownload         If specified, all data is redownloaded [default: false]
  --concurrency CONCURRENCY
                         The number of simultaneous downloads [default: 16]
  --tourtype TOURTYPE    The type of tours to download
  --help, -h             display this help and exit
```

## Running a full export

To download all planned and recorded tours, you can run:

```
./export-komoot --email "<email>" --password "<password>" --userid "<user_id>" --to "<destination_path>" --fulldownload
```

This will download all tours, even if they already exist in the target location.

## Running an incremental export (the default)

To only download the tours which aren't downloaded yet or those that were updated, you can run it like this:

```
./export-komoot --email "<email>" --password "<password>" --userid "<user_id>" --to "<destination_path>"
```

## Filtering the list of tours

To add a filter to the list of tours that need to be exported, you can use the `--filter` parameter. The filter works
the same way as the search field in the Komoot user interface.

```
./export-komoot --email "<email>" --password "<password>" --userid "<user_id>" --to "<destination_path>" --filter "<filter>"
```

## Using a `.env` file

To avoid that you always have to specify the email, password and user ID, you can store them in a `.env` file.

Create a `.env` file which should include your username, password and user ID in the current working directory:

```env
KOMOOT_EMAIL=user@host.com
KOMOOT_PASSWD=password
KOMOOT_USER_ID=123456
```

Once this is set, you can omit the following parameters from the command:

- `--email "<email>"`
- `--password "<password>"`
- `--userid "<user_id>"`

# About the generated filenames

The generated filenames use the following structure:

```
<date>_<id>_<name>_<type>_<changed-timestamp>.gpx
```

- `<date>`: the date of the tour in the format `YYYM-MM-DD`
- `<id>`: the unique ID of the tour in Komoot
- `<name>`: the name of the tour in a filesystem friendly way
- `<type>`: the type of the tour (`tour_planned` or `tour_recorded`)
- `<changed-timestamp>`: the last changed datetime of the tour as a unix timestamp (needed for the incremental export)

# Limitations

The tool will only export the first 5000 tours at the moment.

# Building

If you prefer to compile the binaries yourself, you will need to have [Golang version 1.24](https://go.dev) or higher
installed. To make building easier, you also need to have the [`make`](https://www.gnu.org/software/make/) utility
installed. Once installed, you can build the binary for the platform you are running using:

```
make build
```

If you want to compile all supported architectures and operating systems, you can execute:

```
make build-all
```

To run the tests, you can execute:

```
make test
```

# References

- [The new Komoot authentication](https://github.com/Woeler/komoot-php/commit/21065fcf517cc0fac646a6a216b5cf2d851f7975#diff-17339dceedd73393b090f1db8e636e6a8a5a161944c87d85dcd8ec3789dd6112)
