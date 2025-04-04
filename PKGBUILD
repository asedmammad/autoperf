# Maintainer: Ased Mammad <asedmammad@gmail.com>
pkgname=autoperf
pkgver=1.1.0
pkgrel=1
pkgdesc="Automatic CPU performance bias adjustment based on system load and temperature"
arch=('x86_64')
url="https://github.com/asedmammad/autoperf"
license=('MIT')
depends=('glibc' 'x86_energy_perf_policy')
makedepends=('go')
backup=('etc/autoperf.conf')
source=("${pkgname}-${pkgver}.tar.gz::${url}/archive/v${pkgver}.tar.gz")
sha256sums=('SKIP')

build() {
    cd "$pkgname-$pkgver"
    export CGO_CPPFLAGS="${CPPFLAGS}"
    export CGO_CFLAGS="${CFLAGS}"
    export CGO_CXXFLAGS="${CXXFLAGS}"
    export CGO_LDFLAGS="${LDFLAGS}"
    # export GOFLAGS="-buildmode=pie -trimpath -ldflags=-linkmode=external -mod=readonly -modcacherw"
    
    go build -ldflags "-s -w" -o autoperf .
}

package() {
    cd "$pkgname-$pkgver"
    
    # Install binary
    install -Dm755 autoperf "$pkgdir/usr/bin/autoperf"
    
    # Install man page
    install -Dm644 doc/autoperf.conf.5 "$pkgdir/usr/share/man/man5/autoperf.conf.5"
    
    # Install config file
    install -Dm644 autoperf.conf "$pkgdir/etc/autoperf.conf"
    
    # Install systemd service file
    install -Dm644 systemd/autoperf.service "$pkgdir/usr/lib/systemd/system/autoperf.service"
    
    # Install license file (assuming you have one, if not you should create it)
    # install -Dm644 LICENSE "$pkgdir/usr/share/licenses/$pkgname/LICENSE"
}
