include $(GOROOT)/src/Make.$(GOARCH)
 
TARG=godwulf
GOFILES=godwulf.go

CLEANFILES+=${TARG}

include $(GOROOT)/src/Make.pkg

DESTDIR=/usr/local/bin

all: godwulf

%: %.go
	@$(QUOTED_GOBIN)/$(GC) $*.go
	@$(QUOTED_GOBIN)/$(LD) -o $@ $*.$O

install: all
	@echo installing executable file to ${DESTDIR}/bin
	@mkdir -p ${DESTDIR}/bin
	@cp -f ${TARG} ${DESTDIR}/bin
	@chmod 755 ${DESTDIR}/bin

uninstall:
	@echo removing executable file from ${DESTDIR}/bin
	@rm -f ${DESTDIR}/bin/${TARG}
