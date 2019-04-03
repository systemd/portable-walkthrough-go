all: portable-walkthrough-go.raw

# Build the static service binary from the Go source file
portable-walkthrough-go: main.go
	CGO_ENABLED=0 go build -tags netgo -compiler gccgo -gccgoflags '-static'

# Build a a squashfs file system from the static service binary, the
# two unit files and the os-release file (together with some auxiliary
# empty directories and files that can be over-mounted from the host.
portable-walkthrough-go.raw: portable-walkthrough-go portable-walkthrough-go.service portable-walkthrough-go.socket os-release
	( rm -rf build && \
	  mkdir -p build/usr/bin build/usr/lib/systemd/system build/etc build/proc build/sys build/dev build/run build/tmp build/var/tmp build/var/lib/walkthrough-go && \
	  cp portable-walkthrough-go build/usr/bin/ && \
	  cp portable-walkthrough-go.service portable-walkthrough-go.socket build/usr/lib/systemd/system && \
	  cp os-release build/usr/lib/os-release && \
	  touch build/etc/resolv.conf build/etc/machine-id && \
	  rm -f portable-walkthrough-go.raw && \
	  mksquashfs build/ portable-walkthrough-go.raw )

# A shortcut for installing all needed build-time dependencies on Fedora
install-tools:
	dnf install gcc-go libgo-static glibc-static squashfs-tools

clean:
	go clean
	rm -rf build
	rm -f portable-walkthrough-go.raw

.PHONY: all install-tools clean
