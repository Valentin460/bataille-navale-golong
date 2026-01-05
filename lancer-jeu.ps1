# Script pour lancer 2 joueurs automatiquement
Write-Host "╔════════════════════════════════════════════════════════════════╗" -ForegroundColor Cyan
Write-Host "║         🚢 BATAILLE NAVALE - LANCEMENT AUTO 🚢                 ║" -ForegroundColor Cyan
Write-Host "╚════════════════════════════════════════════════════════════════╝" -ForegroundColor Cyan
Write-Host ""

Write-Host "🎮 Démarrage du Joueur 1 (port 8080)..." -ForegroundColor Green
Start-Process powershell -ArgumentList "-NoExit", "-Command", "cd '$PWD'; go run main.go --port=8080 --opponents=http://localhost:8081"

Write-Host "⏳ Attente de 2 secondes..." -ForegroundColor Yellow
Start-Sleep -Seconds 2

Write-Host "🎮 Démarrage du Joueur 2 (port 8081)..." -ForegroundColor Green
Start-Process powershell -ArgumentList "-NoExit", "-Command", "cd '$PWD'; go run main.go --port=8081 --opponents=http://localhost:8080"

Write-Host ""
Write-Host "✅ Deux fenêtres ont été ouvertes !" -ForegroundColor Green
Write-Host "💡 Vous pouvez maintenant jouer dans chaque fenêtre" -ForegroundColor Yellow
Write-Host ""
Write-Host "Appuyez sur une touche pour fermer cette fenêtre..." -ForegroundColor Gray
$null = $Host.UI.RawUI.ReadKey("NoEcho,IncludeKeyDown")
