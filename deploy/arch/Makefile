
pkgname=gopilot
pkgver=0.1.0
pkgrel=1

.PHONY: $(pkgname)-$(pkgver).tar.gz

archive: $(pkgname)-$(pkgver).tar.gz
$(pkgname)-$(pkgver).tar.gz:
	tar \
	-C ../../ \
	--transform 's,^\.,$(pkgname)-$(pkgver),' \
	-caf $@ \
	./Makefile \
	./deploy/common \
	./src/core \
	./src/plugins \
	./src/copilotg.go 

package:
	makepkg --skipchecksums

clean:
	@rm -vfR pkg src
	@rm -vf *.gz *.xz