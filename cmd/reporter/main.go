// Command reporter je CLI pro přehled uložených simulací.
// Reporty samotné generuje engine do REPORTS_DIR; tento nástroj je pro inspekci DB.
package main

import (
	"fmt"
	"os"

	"github.com/ol1n/auction-sim/internal/config"
	"github.com/ol1n/auction-sim/internal/store"
)

func main() {
	cfg := config.Load()
	st, err := store.Open(cfg.DBPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "otevření DB:", err)
		os.Exit(1)
	}
	defer st.Close()

	sims, err := st.ListSimulations()
	if err != nil {
		fmt.Fprintln(os.Stderr, "dotaz:", err)
		os.Exit(1)
	}
	if len(sims) == 0 {
		fmt.Println("Žádné simulace.")
		return
	}
	fmt.Printf("%-22s %-10s %-22s %-5s %s\n", "ID", "STATUS", "TYPY", "RUNS", "NÁZEV")
	for _, s := range sims {
		fmt.Printf("%-22s %-10s %-22s %-5d %s\n", s.ID, s.Status, s.Types, s.Runs, s.Name)
	}
}
