#!/bin/bash

# Build script for retroplayer with size optimization

echo "Compilation de RetroPlayer..."
echo "Compilation pour $(go env GOOS)/$(go env GOARCH)..."

# Create build directory if it doesn't exist
mkdir -p build

# Build with stripped symbols for smaller binary
go build -ldflags="-s -w" -o build/retroplayer

# Check if UPX is available and compress the binary
if command -v upx &> /dev/null
then
    echo "Compressing binary with UPX..."
    upx --best --lzma build/retroplayer
    echo "Binary compressed successfully."
else
    echo "UPX not found. Install UPX for additional size reduction."
fi

echo "Build completed."
echo "Compilation réussie!"
echo "L'exécutable se trouve dans: build/retroplayer"

echo ""
echo "Pour l'exécuter:"
echo "  cd build && ./retroplayer"
echo ""
echo "Options disponibles:"
echo "  ./build/retroplayer          # Lancer l'interface graphique"
echo "  ./build/retroplayer --update # Mettre à jour la base de données"
echo "  ./build/retroplayer --help   # Afficher l'aide"