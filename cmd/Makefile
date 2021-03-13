OUT := sysroot-manager
DISTRO := $(shell cat /etc/os-release | grep '^NAME=' | sed -e 's/.*=//')
GCCGO :=
GO := go

SOURCES = \
	*.go

all: update-deps mainbuild strip

mainbuild: $(SOURCES)
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GO) build -a -tags netgo -ldflags '-w -extldflags "-static"' -o $(OUT) $(SOURCES)

gcc:
ifneq ($GCCGO,)
	$(GO) build -compiler $(GCCGO) -gccgoflags '-static-libgo' -o $(OUT)
else
	@echo "Cannot find gccgo"
endif

update-deps:
	$(GO) get -u

strip:
	strip $(OUT)

deps:
	$(GO) get -v ./... 2>/dev/null ||:

clean:
	rm $(OUT)
	$(GO) clean -x -i

flush-cache:
	$(GO) clean -cache -modcache -i -r

.PHONY: all install clean
