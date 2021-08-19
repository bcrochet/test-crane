# Test Crane Functionality

Using Crane to pull an image from a a registry, write it as a tarball, and
extract the tarball to disk.

## Build

```shell
go build .
```

## Run

Without credentials:

```shell
./test-crane <ImageWithRegistry>
```

With credentials:

```shell
read -s YOUR_PASS
<enter your password>
REG_USERNAME=your_username REG_PASSWORD=$YOUR_PASS ./test-crane <ImageWithRegistry>
```
