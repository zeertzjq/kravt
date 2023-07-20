PREFIX ?= /usr/local

all: build

build:
	go build

clean:
	rm -f kravt

install:
	mkdir -p $(DESTDIR)$(PREFIX)/bin
	cp -f kravt $(DESTDIR)$(PREFIX)/bin/
	chmod 755 $(DESTDIR)$(PREFIX)/bin/kravt

uninstall:
	rm -f $(DESTDIR)$(PREFIX)/bin/kravt
