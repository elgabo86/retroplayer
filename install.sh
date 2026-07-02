#!/usr/bin/bash

# Script d'installation de RetroPlayer
# Télécharge et installe retroplayer et sa dépendance zxtune

set -eoux pipefail

echo "Installation de RetroPlayer..."

# Installation de zxtune (dépendance runtime)
echo "Téléchargement de zxtune..."
curl -fsSL https://storage.zxtune.ru/builds/public/r5100/linux/x86_64/zxtune_r5100_linux_x86_64.tar.gz | \
    sudo tar -xzf - -C / --exclude=./usr/bin/zxtune123 --exclude=./usr/bin/xtractor

# Installation de retroplayer depuis la dernière release GitHub
echo "Téléchargement de retroplayer..."
sudo curl -fsSL "https://github.com/elgabo86/retroplayer/releases/latest/download/retroplayer" \
    -o /usr/bin/retroplayer
sudo chmod +x /usr/bin/retroplayer

echo "RetroPlayer installé avec succès."
echo ""
echo "Première utilisation :"
echo "  retroplayer --update   # Mettre à jour la base de données"
echo "  retroplayer            # Lancer l'interface"
