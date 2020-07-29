# Build instructions

On linux, for linux:
Ensure modvendor is installed by running `go get -u github.com/goware/modvendor` outside of the project directory.
```sh
make build
```
Best is to also use the build container though, to avoid issues with library versions.

On linux, cross-compile for windows:
```sh
sudo docker build -t testbed_build .
sudo docker run -it -v /home/oliver/git/electronics-jam:/build testbed_build
cd /build
make build_cc_windows
```
