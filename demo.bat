@echo off
echo ╔════════════════════════════════════════════════════════════════╗
echo ║              🚢 BATAILLE NAVALE - DÉMO 🚢                      ║
echo ╚════════════════════════════════════════════════════════════════╝
echo.
echo Ce script démarre deux instances du jeu pour tester le client
echo.
echo 📌 Instance 1 : Port 8080
echo 📌 Instance 2 : Port 8081
echo.
echo Appuyez sur une touche pour démarrer...
pause > nul

echo.
echo 🚀 Compilation du projet...
go build -o bataille-navale.exe
if errorlevel 1 (
    echo ❌ Erreur de compilation
    pause
    exit /b 1
)

echo ✓ Compilation réussie
echo.
echo 🎮 Démarrage de l'instance 1 sur le port 8080...
start "Bataille Navale - Joueur 1" cmd /k "bataille-navale.exe --port=8080 --opponents=http://localhost:8081"

timeout /t 2 /nobreak > nul

echo 🎮 Démarrage de l'instance 2 sur le port 8081...
start "Bataille Navale - Joueur 2" cmd /k "bataille-navale.exe --port=8081 --opponents=http://localhost:8080"

echo.
echo ✓ Les deux instances sont démarrées !
echo.
echo 💡 Deux fenêtres devraient s'être ouvertes
echo 💡 Vous pouvez maintenant jouer dans chaque fenêtre
echo.
echo Appuyez sur une touche pour fermer cette fenêtre...
pause > nul
