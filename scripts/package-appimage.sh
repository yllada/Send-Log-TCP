#!/bin/bash
# =============================================================================
# Script para crear AppImage de SendLog Syslog
# =============================================================================
# AppImage es un formato de distribución que empaqueta la aplicación con
# todas sus dependencias, permitiendo ejecutarla en cualquier distribución
# Linux sin instalación.
#
# Requisitos:
#   - appimagetool (se descarga automáticamente)
#   - La app ya compilada en build/bin/
#
# Uso:
#   ./scripts/package-appimage.sh
#
# Documentación: https://appimage.org/
# =============================================================================

set -e

APP_NAME="Sendlog-Syslog"
APP_VERSION="${1:-1.0.0}"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"
BUILD_DIR="$PROJECT_DIR/build"
APPDIR="$BUILD_DIR/AppDir"
OUTPUT_DIR="$BUILD_DIR/bin"

echo "📦 Creando AppImage para $APP_NAME v$APP_VERSION"

# Verificar que existe el binario
if [ ! -f "$OUTPUT_DIR/$APP_NAME" ] && [ ! -f "$OUTPUT_DIR/Sendlog-Syslog" ]; then
    echo "❌ Error: No se encontró el binario. Ejecuta 'wails build' primero."
    exit 1
fi

BINARY_PATH="$OUTPUT_DIR/Sendlog-Syslog"
if [ -f "$OUTPUT_DIR/$APP_NAME" ]; then
    BINARY_PATH="$OUTPUT_DIR/$APP_NAME"
fi

# Limpiar AppDir anterior
rm -rf "$APPDIR"
mkdir -p "$APPDIR/usr/bin"
mkdir -p "$APPDIR/usr/share/applications"
mkdir -p "$APPDIR/usr/share/icons/hicolor/256x256/apps"
mkdir -p "$APPDIR/usr/share/icons/hicolor/128x128/apps"
mkdir -p "$APPDIR/usr/share/icons/hicolor/64x64/apps"
mkdir -p "$APPDIR/usr/share/icons/hicolor/48x48/apps"
mkdir -p "$APPDIR/usr/share/icons/hicolor/32x32/apps"

# Copiar binario
echo "📋 Copiando binario..."
cp "$BINARY_PATH" "$APPDIR/usr/bin/sendlog-syslog"
chmod +x "$APPDIR/usr/bin/sendlog-syslog"

# Copiar icono (usar el de build si existe, o crear uno placeholder)
ICON_SOURCE="$BUILD_DIR/appicon.png"
if [ -f "$ICON_SOURCE" ]; then
    echo "🎨 Copiando iconos..."
    cp "$ICON_SOURCE" "$APPDIR/usr/share/icons/hicolor/256x256/apps/sendlog-syslog.png"
    cp "$ICON_SOURCE" "$APPDIR/sendlog-syslog.png"
    
    # Crear versiones de diferentes tamaños si imagemagick está disponible
    if command -v convert &> /dev/null; then
        convert "$ICON_SOURCE" -resize 128x128 "$APPDIR/usr/share/icons/hicolor/128x128/apps/sendlog-syslog.png"
        convert "$ICON_SOURCE" -resize 64x64 "$APPDIR/usr/share/icons/hicolor/64x64/apps/sendlog-syslog.png"
        convert "$ICON_SOURCE" -resize 48x48 "$APPDIR/usr/share/icons/hicolor/48x48/apps/sendlog-syslog.png"
        convert "$ICON_SOURCE" -resize 32x32 "$APPDIR/usr/share/icons/hicolor/32x32/apps/sendlog-syslog.png"
    fi
else
    echo "⚠️  No se encontró icono en $ICON_SOURCE"
fi

# Crear archivo .desktop
echo "📝 Creando archivo .desktop..."
cat > "$APPDIR/sendlog-syslog.desktop" << EOF
[Desktop Entry]
Type=Application
Name=SendLog Syslog
Comment=Professional syslog message sender for network testing and monitoring
Exec=sendlog-syslog
Icon=sendlog-syslog
Categories=Network;Utility;Development;
Keywords=syslog;log;network;tcp;udp;monitoring;
Terminal=false
StartupNotify=true
StartupWMClass=sendlog-syslog
EOF

# Copiar .desktop también a usr/share/applications
cp "$APPDIR/sendlog-syslog.desktop" "$APPDIR/usr/share/applications/"

# Crear AppRun script
echo "🔧 Creando AppRun..."
cat > "$APPDIR/AppRun" << 'EOF'
#!/bin/bash
# AppRun script para SendLog Syslog
# Este script configura el entorno y ejecuta la aplicación

SELF=$(readlink -f "$0")
HERE=${SELF%/*}

# Configurar variables de entorno para GTK/WebKit
export PATH="${HERE}/usr/bin:${PATH}"
export LD_LIBRARY_PATH="${HERE}/usr/lib:${HERE}/usr/lib/x86_64-linux-gnu:${LD_LIBRARY_PATH}"
export XDG_DATA_DIRS="${HERE}/usr/share:${XDG_DATA_DIRS}"
export GTK_PATH="${HERE}/usr/lib/gtk-3.0"
export GDK_PIXBUF_MODULE_FILE="${HERE}/usr/lib/gdk-pixbuf-2.0/loaders.cache"
export GDK_PIXBUF_MODULEDIR="${HERE}/usr/lib/gdk-pixbuf-2.0/loaders"

# Ejecutar la aplicación
exec "${HERE}/usr/bin/sendlog-syslog" "$@"
EOF
chmod +x "$APPDIR/AppRun"

# Descargar appimagetool si no existe
APPIMAGETOOL="$BUILD_DIR/appimagetool-x86_64.AppImage"
if [ ! -f "$APPIMAGETOOL" ]; then
    echo "📥 Descargando appimagetool..."
    wget -q -O "$APPIMAGETOOL" "https://github.com/AppImage/AppImageKit/releases/download/continuous/appimagetool-x86_64.AppImage"
    chmod +x "$APPIMAGETOOL"
fi

# Crear AppImage
# Usar --appimage-extract-and-run para evitar problemas con FUSE en CI/CD
echo "🔨 Creando AppImage..."
ARCH=x86_64 "$APPIMAGETOOL" --appimage-extract-and-run "$APPDIR" "$OUTPUT_DIR/${APP_NAME}-${APP_VERSION}-x86_64.AppImage"

# Limpiar
rm -rf "$APPDIR"

echo ""
echo "✅ AppImage creado exitosamente:"
echo "   $OUTPUT_DIR/${APP_NAME}-${APP_VERSION}-x86_64.AppImage"
echo ""
echo "📋 Para ejecutar:"
echo "   chmod +x ${APP_NAME}-${APP_VERSION}-x86_64.AppImage"
echo "   ./${APP_NAME}-${APP_VERSION}-x86_64.AppImage"
