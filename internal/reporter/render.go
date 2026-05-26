package reporter

import (
	"encoding/json"
	"html/template"
	"os"
	"path/filepath"
)

var reportTmpl = template.Must(template.New("report").Parse(`<!DOCTYPE html>
<html lang="cs"><head><meta charset="utf-8">
<title>Auction Sim Report — {{.Name}}</title>
<script src="https://cdn.jsdelivr.net/npm/chart.js@4"></script>
<style>
 body{font-family:system-ui,sans-serif;max-width:1000px;margin:2rem auto;padding:0 1rem;color:#1a1a1a}
 h1,h2{border-bottom:2px solid #eee;padding-bottom:.3rem}
 table{border-collapse:collapse;width:100%;margin:1rem 0}
 th,td{border:1px solid #ddd;padding:.4rem .6rem;text-align:right;font-size:.9rem}
 th:first-child,td:first-child{text-align:left}
 .insights{background:#f7f7f9;padding:1rem 1.5rem;border-radius:8px;white-space:pre-wrap}
 canvas{max-height:340px}
</style></head><body>
<h1>{{.Name}}</h1>
<p>Zboží: <strong>{{.Item.Name}}</strong> (retail {{printf "%.0f" .Item.RetailPrice}} Kč)</p>

<h2>Srovnání typů aukcí (průměr přes běhy)</h2>
<table><thead><tr><th>Typ</th><th>Běhů</th><th>Ø Final cena</th><th>Ø Profit platformy</th><th>Ø Profit prodejce</th><th>Ø Bidů</th><th>Ø Conversion</th><th>Ø Viral cost</th><th>Ø Efficiency</th></tr></thead>
<tbody>
{{range $t, $a := .Aggregates}}<tr><td>{{$t}}</td><td>{{$a.Runs}}</td><td>{{printf "%.1f" $a.FinalPrice.Mean}}</td><td>{{printf "%.1f" $a.PlatformProfit.Mean}}</td><td>{{printf "%.1f" $a.SellerProfit.Mean}}</td><td>{{printf "%.1f" $a.TotalBids.Mean}}</td><td>{{printf "%.2f" $a.Conversion.Mean}}</td><td>{{printf "%.1f" $a.ViralCost.Mean}}</td><td>{{printf "%.2f" $a.Efficiency.Mean}}</td></tr>
{{end}}</tbody></table>

<h2>Profit platformy dle typu</h2><canvas id="profitChart"></canvas>
<h2>Viral cost vs. profit platformy</h2><canvas id="viralChart"></canvas>

<h2>Jednotlivé běhy</h2>
<table><thead><tr><th>Typ</th><th>Run</th><th>Vítěz</th><th>Final cena</th><th>Profit platf.</th><th>Loser ztráty</th><th>Bidů</th><th>Efficiency</th></tr></thead>
<tbody>
{{range .Reports}}<tr><td>{{.Type}}</td><td>{{.RunIndex}}</td><td>{{.WinnerName}}</td><td>{{printf "%.1f" .FinalPrice}}</td><td>{{printf "%.1f" .PlatformProfit}}</td><td>{{printf "%.1f" .TotalLoserLosses}}</td><td>{{.TotalBidsPlaced}}</td><td>{{printf "%.0f" .EfficiencyScore}}</td></tr>
{{end}}</tbody></table>

<h2>LLM Insights</h2>
<div class="insights">{{.Insights}}</div>

<script>
const labels = {{.ChartLabels}};
new Chart(document.getElementById('profitChart'),{type:'bar',data:{labels:labels,datasets:[{label:'Ø Profit platformy',data:{{.ProfitData}}}]}});
new Chart(document.getElementById('viralChart'),{type:'scatter',data:{datasets:[{label:'typ',data:{{.ViralScatter}}}]},options:{scales:{x:{title:{display:true,text:'Viral cost'}},y:{title:{display:true,text:'Profit platformy'}}}}});
</script>
</body></html>`))

type renderData struct {
	SimulationReport
	ChartLabels  template.JS
	ProfitData   template.JS
	ViralScatter template.JS
}

// Render zapíše HTML report do reportsDir a vrátí cestu k souboru.
func Render(rep SimulationReport, reportsDir string) (string, error) {
	var labels []string
	var profit []float64
	type pt struct {
		X float64 `json:"x"`
		Y float64 `json:"y"`
	}
	var scatter []pt
	for typ, agg := range rep.Aggregates {
		labels = append(labels, string(typ))
		profit = append(profit, agg.PlatformProfit.Mean)
		scatter = append(scatter, pt{X: agg.ViralCost.Mean, Y: agg.PlatformProfit.Mean})
	}
	labelsJSON, _ := json.Marshal(labels)
	profitJSON, _ := json.Marshal(profit)
	scatterJSON, _ := json.Marshal(scatter)

	if err := os.MkdirAll(reportsDir, 0o755); err != nil {
		return "", err
	}
	path := filepath.Join(reportsDir, rep.SimulationID+".html")
	f, err := os.Create(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	err = reportTmpl.Execute(f, renderData{
		SimulationReport: rep,
		ChartLabels:      template.JS(labelsJSON),
		ProfitData:       template.JS(profitJSON),
		ViralScatter:     template.JS(scatterJSON),
	})
	return path, err
}
