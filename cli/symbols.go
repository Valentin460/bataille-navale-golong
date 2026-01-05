package cli

// Symboles pour l'affichage du jeu
const (
	// Statuts des adversaires
	SymbolAlive       = "[+]" // Vivant
	SymbolDefeated    = "[X]" // Éliminé
	SymbolUnreachable = "[!]" // Injoignable
	SymbolUnknown     = "[?]" // Inconnu
	
	// États des cases
	SymbolWater    = "~"  // Eau / Non révélé
	SymbolMiss     = "o"  // Tir dans l'eau
	SymbolHit      = "X"  // Bateau touché
	SymbolBoat     = "B"  // Bateau non touché (vue du joueur)
	
	// Notifications
	SymbolSuccess = "[OK]"  // Succès
	SymbolError   = "[!!]"  // Erreur
	SymbolWarning = "[!]"   // Avertissement
	SymbolInfo    = "[i]"   // Information
	
	// Tirs
	SymbolHitShot  = "[*]"  // Tir touché
	SymbolMissShot = "[.]"  // Tir raté
)

// Symboles colorés (optionnel, si le terminal supporte les couleurs ANSI)
const (
	// Codes couleur ANSI
	ColorReset   = "\033[0m"
	ColorRed     = "\033[31m"
	ColorGreen   = "\033[32m"
	ColorYellow  = "\033[33m"
	ColorBlue    = "\033[34m"
	ColorMagenta = "\033[35m"
	ColorCyan    = "\033[36m"
	ColorGray    = "\033[37m"
	ColorBold    = "\033[1m"
)

// FormatStatus retourne le symbole formaté pour un statut
func FormatStatus(alive, defeated, unreachable bool) string {
	if defeated {
		return ColorRed + SymbolDefeated + ColorReset
	}
	if unreachable {
		return ColorYellow + SymbolUnreachable + ColorReset
	}
	if alive {
		return ColorGreen + SymbolAlive + ColorReset
	}
	return SymbolUnknown
}

// FormatCell retourne le symbole formaté pour une case
func FormatCell(state int, colored bool) string {
	var symbol string
	switch state {
	case 0:
		symbol = SymbolWater
	case 1:
		symbol = SymbolMiss
		if colored {
			symbol = ColorCyan + symbol + ColorReset
		}
	case 2:
		symbol = SymbolHit
		if colored {
			symbol = ColorRed + symbol + ColorReset
		}
	case 3:
		symbol = SymbolBoat
		if colored {
			symbol = ColorGreen + symbol + ColorReset
		}
	default:
		symbol = "?"
	}
	return symbol
}

// FormatNotification retourne le symbole formaté pour une notification
func FormatNotification(msgType string, colored bool) string {
	var symbol string
	switch msgType {
	case "success":
		symbol = SymbolSuccess
		if colored {
			symbol = ColorGreen + symbol + ColorReset
		}
	case "error":
		symbol = SymbolError
		if colored {
			symbol = ColorRed + symbol + ColorReset
		}
	case "warning":
		symbol = SymbolWarning
		if colored {
			symbol = ColorYellow + symbol + ColorReset
		}
	case "info":
		symbol = SymbolInfo
		if colored {
			symbol = ColorBlue + symbol + ColorReset
		}
	default:
		symbol = SymbolInfo
	}
	return symbol
}
