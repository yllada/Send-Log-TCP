#!/bin/bash
# =============================================================================
# Script para crear paquete DEB de SendLog Syslog
# =============================================================================
# Crea un paquete .deb para instalación en Debian/Ubuntu/derivados.
# El paquete declara las dependencias, que se instalan automáticamente.
#
# Uso:
#   ./scripts/package-deb.sh [version]
#
# Ejemplo:
#   ./scripts/package-deb.sh 1.4.0
#
# Documentación: https://www.debian.org/doc/debian-policy/
# =============================================================================

set -e

APP_NAME="sendlog-syslog"
APP_VERSION="${1:-1.0.0}"
MAINTAINER="Yadian Llada Lopez <yadian.llada@gmail.com>"
DESCRIPTION="Professional syslog message sender for network testing and monitoring"
HOMEPAGE="https://github.com/yllada/Send-Log-TCP"
ARCH="amd64"

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"
BUILD_DIR="$PROJECT_DIR/build"
PACKAGE_DIR="$BUILD_DIR/${APP_NAME}_${APP_VERSION}_${ARCH}"
OUTPUT_DIR="$BUILD_DIR/bin"

echo "📦 Creando paquete DEB para $APP_NAME v$APP_VERSION"

# Verificar que existe el binario
BINARY_PATH="$OUTPUT_DIR/Sendlog-Syslog"
if [ ! -f "$BINARY_PATH" ]; then
    echo "❌ Error: No se encontró el binario. Ejecuta 'wails build' primero."
    exit 1
fi

# Limpiar directorio anterior
rm -rf "$PACKAGE_DIR"

# Crear estructura de directorios del paquete
mkdir -p "$PACKAGE_DIR/DEBIAN"
mkdir -p "$PACKAGE_DIR/usr/bin"
mkdir -p "$PACKAGE_DIR/usr/share/applications"
mkdir -p "$PACKAGE_DIR/usr/share/icons/hicolor/256x256/apps"
mkdir -p "$PACKAGE_DIR/usr/share/icons/hicolor/128x128/apps"
mkdir -p "$PACKAGE_DIR/usr/share/icons/hicolor/48x48/apps"
mkdir -p "$PACKAGE_DIR/usr/share/doc/$APP_NAME"
mkdir -p "$PACKAGE_DIR/usr/share/metainfo"

# Copiar binario
echo "📋 Copiando binario..."
cp "$BINARY_PATH" "$PACKAGE_DIR/usr/bin/$APP_NAME"
chmod 755 "$PACKAGE_DIR/usr/bin/$APP_NAME"

# Copiar icono
ICON_SOURCE="$BUILD_DIR/appicon.png"
if [ -f "$ICON_SOURCE" ]; then
    echo "🎨 Copiando iconos..."
    cp "$ICON_SOURCE" "$PACKAGE_DIR/usr/share/icons/hicolor/256x256/apps/$APP_NAME.png"
    
    if command -v convert &> /dev/null; then
        convert "$ICON_SOURCE" -resize 128x128 "$PACKAGE_DIR/usr/share/icons/hicolor/128x128/apps/$APP_NAME.png"
        convert "$ICON_SOURCE" -resize 48x48 "$PACKAGE_DIR/usr/share/icons/hicolor/48x48/apps/$APP_NAME.png"
    fi
fi

# Crear archivo .desktop
echo "📝 Creando archivo .desktop..."
cat > "$PACKAGE_DIR/usr/share/applications/$APP_NAME.desktop" << EOF
[Desktop Entry]
Type=Application
Name=SendLog Syslog
GenericName=Syslog Client
Comment=$DESCRIPTION
Exec=$APP_NAME
Icon=$APP_NAME
Categories=Network;Utility;Development;System;
Keywords=syslog;log;network;tcp;udp;monitoring;testing;
Terminal=false
StartupNotify=true
StartupWMClass=Sendlog-Syslog
EOF

# Crear archivo AppStream metainfo (para GNOME Software, etc.)
echo "📝 Creando AppStream metadata..."
cat > "$PACKAGE_DIR/usr/share/metainfo/$APP_NAME.appdata.xml" << EOF
<?xml version="1.0" encoding="UTF-8"?>
<component type="desktop-application">
  <id>com.github.yllada.sendlog-syslog</id>
  <metadata_license>MIT</metadata_license>
  <project_license>Apache-2.0</project_license>
  <name>SendLog Syslog</name>
  <summary>Professional syslog message sender</summary>
  <description>
    <p>
      SendLog Syslog is a modern desktop application for sending syslog messages
      to remote servers via TCP or UDP protocols.
    </p>
    <p>Features:</p>
    <ul>
      <li>Multi-protocol support: TCP, UDP, TCP+TLS</li>
      <li>RFC 5424 and RFC 3164 message formats</li>
      <li>TLS/SSL encryption with certificate verification</li>
      <li>Modern UI with dark/light theme support</li>
    </ul>
  </description>
  <launchable type="desktop-id">$APP_NAME.desktop</launchable>
  <url type="homepage">$HOMEPAGE</url>
  <url type="bugtracker">$HOMEPAGE/issues</url>
  <developer_name>Yadian Llada Lopez</developer_name>
  <provides>
    <binary>$APP_NAME</binary>
  </provides>
  <content_rating type="oars-1.1" />
  <releases>
    <release version="$APP_VERSION" date="$(date +%Y-%m-%d)" />
  </releases>
</component>
EOF

# Copiar documentación
echo "📚 Copiando documentación..."
if [ -f "$PROJECT_DIR/README.md" ]; then
    cp "$PROJECT_DIR/README.md" "$PACKAGE_DIR/usr/share/doc/$APP_NAME/"
fi
if [ -f "$PROJECT_DIR/LICENSE" ]; then
    cp "$PROJECT_DIR/LICENSE" "$PACKAGE_DIR/usr/share/doc/$APP_NAME/copyright"
fi
if [ -f "$PROJECT_DIR/CHANGELOG.md" ]; then
    cp "$PROJECT_DIR/CHANGELOG.md" "$PACKAGE_DIR/usr/share/doc/$APP_NAME/"
fi

# Crear archivo de control
# Detectamos la versión de Ubuntu para las dependencias correctas
echo "📋 Creando archivo de control..."
cat > "$PACKAGE_DIR/DEBIAN/control" << EOF
Package: $APP_NAME
Version: $APP_VERSION
Section: net
Priority: optional
Architecture: $ARCH
Depends: libgtk-3-0, libwebkit2gtk-4.1-0 | libwebkit2gtk-4.0-37
Recommends: ca-certificates
Maintainer: $MAINTAINER
Homepage: $HOMEPAGE
Description: $DESCRIPTION
 SendLog Syslog is a modern desktop application built with Wails
 for sending syslog messages to remote servers via TCP or UDP protocols.
 .
 Features:
  - Multi-protocol support: TCP, UDP, TCP+TLS
  - RFC 5424 (modern) and RFC 3164 (legacy) message formats
  - TLS/SSL encryption with certificate verification
  - TCP framing according to RFC 6587
  - Modern UI with dark/light theme support
EOF

# Crear script postinst para actualizar cache de iconos
cat > "$PACKAGE_DIR/DEBIAN/postinst" << 'EOF'
#!/bin/bash
set -e

# Actualizar cache de iconos
if command -v gtk-update-icon-cache &> /dev/null; then
    gtk-update-icon-cache -f -t /usr/share/icons/hicolor || true
fi

# Actualizar base de datos de aplicaciones
if command -v update-desktop-database &> /dev/null; then
    update-desktop-database /usr/share/applications || true
fi

exit 0
EOF
chmod 755 "$PACKAGE_DIR/DEBIAN/postinst"

# Crear script postrm para limpiar después de desinstalar
cat > "$PACKAGE_DIR/DEBIAN/postrm" << 'EOF'
#!/bin/bash
set -e

# Actualizar cache de iconos
if command -v gtk-update-icon-cache &> /dev/null; then
    gtk-update-icon-cache -f -t /usr/share/icons/hicolor || true
fi

# Actualizar base de datos de aplicaciones
if command -v update-desktop-database &> /dev/null; then
    update-desktop-database /usr/share/applications || true
fi

exit 0
EOF
chmod 755 "$PACKAGE_DIR/DEBIAN/postrm"

# Calcular tamaño instalado
INSTALLED_SIZE=$(du -sk "$PACKAGE_DIR" | cut -f1)
echo "Installed-Size: $INSTALLED_SIZE" >> "$PACKAGE_DIR/DEBIAN/control"

# Establecer permisos correctos
find "$PACKAGE_DIR" -type d -exec chmod 755 {} \;
find "$PACKAGE_DIR/usr" -type f -exec chmod 644 {} \;
chmod 755 "$PACKAGE_DIR/usr/bin/$APP_NAME"

# Construir el paquete
echo "🔨 Construyendo paquete DEB..."
dpkg-deb --build --root-owner-group "$PACKAGE_DIR"

# Mover a output
DEB_FILE="${APP_NAME}_${APP_VERSION}_${ARCH}.deb"
mv "$BUILD_DIR/$DEB_FILE" "$OUTPUT_DIR/"

# Limpiar
rm -rf "$PACKAGE_DIR"

echo ""
echo "✅ Paquete DEB creado exitosamente:"
echo "   $OUTPUT_DIR/$DEB_FILE"
echo ""
echo "📋 Para instalar:"
echo "   sudo apt install ./$DEB_FILE"
echo ""
echo "📋 Para desinstalar:"
echo "   sudo apt remove $APP_NAME"
