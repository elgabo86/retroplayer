# RetroPlayer

Lecteur de musiques rétro avec interface TUI (Terminal User Interface).

## Fonctionnalités

- Interface utilisateur dans le terminal (TUI)
- Recherche en temps réel des fichiers musicaux
- Navigation fluide avec les flèches du clavier
- Téléchargement et lecture automatique des fichiers
- Base de données locale dans `~/.cache/retroplayer`
- Support de 25 formats de musiques rétro

## Dépendances

- **zxtune-qt** : lecteur de musiques rétro (requis pour la lecture)
- **7z** : extraction des archives (p7zip-full)

## Installation

```bash
curl -fsSL https://raw.githubusercontent.com/elgabo86/retroplayer/main/install.sh | bash
```

## Utilisation

```bash
retroplayer --update   # Mettre à jour la base de données (première utilisation)
retroplayer            # Lancer l'interface
retroplayer --help     # Afficher l'aide
```

## Navigation

- **Flèches ↑↓** : Naviguer dans la liste
- **Gauche/Droite** : Sauter de 5 entrées
- **Entrée** : Jouer le fichier sélectionné
- **ESC** : Annuler le téléchargement / Quitter
- **Ctrl+C** : Quitter
- **Ctrl+U** : Effacer la recherche
- **Backspace** : Effacer un caractère
- **Saisie texte** : Recherche en temps réel

## Formats supportés

SPC (SNES), NSF (NES), USF (N64), SMD (Genesis), PSF (PS1), PSF2 (PS2), PSF3 (PS3),
PSF4 (PS4), GBS (Game Boy), GSF (GBA), 2SF (NDS), 3SF (3DS), DSF (Dreamcast),
Wii, WiiU, GCN, SWITCH, SSF (Saturn), Xbox, X360, PSP, VITA, PC, 3DO

## Licence

MIT
