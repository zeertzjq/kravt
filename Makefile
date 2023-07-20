PREFIX ?= /usr/local

all: build

build:
	go build

clean:
	rm -f kravt

install:
	install -Dm755 kravt -t "$(DESTDIR)$(PREFIX)/bin/"

uninstall:
	rm -f "$(DESTDIR)$(PREFIX)/bin/kravt"
