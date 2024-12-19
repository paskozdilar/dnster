.PHONY: build install uninstall

build: dnster

install:
	cp dnster /usr/local/bin/dnster
	cp dnster.conf /etc/dnster.conf
	cp dnster.service /etc/systemd/system/dnster.service
	systemctl daemon-reload
	systemctl enable dnster
	systemctl start dnster

uninstall:
	systemctl stop dnster
	systemctl disable dnster
	rm -f /etc/systemd/system/dnster.service
	rm -f /etc/dnster.conf
	rm -f /usr/local/bin/dnster
	systemctl daemon-reload

dnster: $(wildcard *.go)
	go build .
