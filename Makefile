

OUT := $(realpath .)/build/out

.PHONY: build
build:
	mkdir -p $(OUT)/resources
	cp -rv ./resources/* $(OUT)/resources/.

	go mod vendor

	modvendor -copy="**/*.c **/*.h **/*.m"

	CGO_ENABLED=1 \
	go build -mod vendor -o $(OUT)/autopilot_testbed main.go

.PHONY: build_cc_windows
build_cc_windows:
	mkdir -p $(OUT)/resources
	cp -rv ./resources/* $(OUT)/resources/.

	GOOS=windows \
	GOARCH=amd64 \
	go mod vendor

	GOOS=windows \
	GOARCH=amd64 \
	modvendor -copy="**/*.c **/*.h **/*.m"

	CGO_ENABLED=1 \
	CC=x86_64-w64-mingw32-gcc \
	GOOS=windows \
	GOARCH=amd64 \
	CGO_LDFLAGS_ALLOW="-Wl,-luuid" \
	CGO_CFLAGS_ALLOW="-Wl,-luuid" \
	go build -mod=vendor -v -o $(OUT)/autopilot_testbed.exe  -ldflags="-H=windowsgui" main.go

.PHONY: package_windows
package_windows:
	cd build/out && rm -f autopilot_testbed_windows.zip && zip -r autopilot_testbed_windows.zip resources/ autopilot_testbed.exe

.PHONY: package_linux
package_linux:
	cd build/out && rm -f autopilot_testbed_linux.tar.gz && tar czf autopilot_testbed_linux.tar.gz resources/ autopilot_testbed
