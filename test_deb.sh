#!/bin/bash
mkdir -p build/deb/opt/nems
mkdir -p build/deb/DEBIAN
mkdir -p build/deb/etc/systemd/system

# Assuming files are already in build/dist and build/nems-server
cp build/nems-server build/deb/opt/nems/
cp -r build/dist build/deb/opt/nems/
cp nems.service build/deb/etc/systemd/system/

cat << 'CTRL' > build/deb/DEBIAN/control
Package: nems
Version: 1.0.0
Section: base
Priority: optional
Architecture: arm64
Maintainer: NEMS Team <team@nems.local>
Description: NEMS Energy Management System
 A complete energy management system for Raspberry Pi.
CTRL

cat << 'POSTINST' > build/deb/DEBIAN/postinst
#!/bin/sh
set -e
# Create nems user if it doesn't exist
if ! id -u nems >/dev/null 2>&1; then
    useradd -r -s /bin/false nems
fi
chown -R nems:nems /opt/nems
echo "nems ALL=(ALL) NOPASSWD: /sbin/reboot, /usr/sbin/reboot, /usr/bin/reboot, /bin/reboot, /bin/systemctl" > /etc/sudoers.d/nems
chmod 0440 /etc/sudoers.d/nems
systemctl daemon-reload
systemctl enable nems.service
systemctl restart nems.service
POSTINST
chmod +x build/deb/DEBIAN/postinst

cat << 'PRERM' > build/deb/DEBIAN/prerm
#!/bin/sh
set -e
systemctl stop nems.service || true
systemctl disable nems.service || true
PRERM
chmod +x build/deb/DEBIAN/prerm

dpkg-deb --build build/deb build/nems_1.0.0_arm64.deb
