package assets

import _ "embed"

// OST c'est la musique compressee (mono 22kHz 24k), embarquee dans le binaire.
// (go:embed interdit les "..", d'ou ce petit package a cote du fichier.)

//go:embed audio/ost.mp3
var OST []byte
