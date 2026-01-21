# Meteo App

Sito web per consultare le previsioni meteo, ricevere notifiche Telegram e scegliere la posizione su mappa.

## Funzionalità

- Visualizzazione meteo attuale e previsioni
- Notifiche Telegram automatiche
- Selezione posizione personalizzata su mappa
- Interfaccia moderna e responsive

## Deploy automatico

Ad ogni push su `main`:

- I file sorgente (.go, go.mod, go.sum) vengono copiati su `/home/zairo/prove-go` su lucaairo.it
- Viene eseguito lo script `deploy.sh` per compilare e avviare l'app

## Link

- [Sito online](https://lucaairo.it/meteo)
