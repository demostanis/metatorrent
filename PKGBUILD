# Maintainer: demostanis worlds <istillhaventgotten@myemailaddress.back>

_pkgname=metatorrent
pkgname=$_pkgname-git
pkgver=1.0.0
pkgrel=1
pkgdesc="Make searches on many well-known torrenting websites at once, fast."
arch=(any)
makedepends=(git go)
source=(git+https://github.com/demostanis/metatorrent)
sha256sums=('SKIP')

build() {
	cd "$srcdir/$_pkgname"

	export CGO_CPPFLAGS="$CPPFLAGS"
	export CGO_CFLAGS="$CFLAGS"
	export CGO_CXXFLAGS="$CXXFLAGS"
	export CGO_LDFLAGS="$LDFLAGS"
	export GOFLAGS='-buildmode=pie -trimpath -mod=readonly -modcacherw'

	go build .
}

package() {
	install -m755 -d "$pkgdir"/usr/bin
	install -m755 "$srcdir/$_pkgname/$_pkgname" "$pkgdir/usr/bin/$_pkgname"
}
