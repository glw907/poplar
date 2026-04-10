// Package theme provides compiled color themes for mailrender and poplar.
//
// Themes are defined as Go values in themes.go. Each theme is built
// from a Palette (16 hex colors) via NewCompiledTheme, which constructs all
// lipgloss styles. There is no runtime file loading.
//
// The styleset generator (styleset.go) writes aerc stylesets from
// the same palette values.
package theme
