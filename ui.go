package main

import (
	"html/template"
	"net/http"
)

const htmlTemplate = `
<!DOCTYPE html>
<html lang="it">
<head>
<meta charset="UTF-8">
<title>Meteo App</title>
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<link rel="stylesheet" href="https://unpkg.com/leaflet@1.9.4/dist/leaflet.css" />
<script src="https://unpkg.com/leaflet@1.9.4/dist/leaflet.js"></script>
<style>
*{margin:0;padding:0;box-sizing:border-box;}
body{
    font-family:"Segoe UI",Tahoma,Geneva,Verdana,sans-serif;
    background:linear-gradient(135deg,#667eea 0%,#764ba2 100%);
    min-height:100vh;
    display:flex;
    justify-content:center;
    align-items:center;
    padding:20px;
}
.container{
    background:white;
    border-radius:20px;
    padding:40px;
    box-shadow:0 20px 60px rgba(0,0,0,0.3);
    max-width:700px;
    width:100%;
}
h1{
    color:#667eea;
    text-align:center;
    margin-bottom:10px;
    font-size:2em;
}
.location{
    text-align:center;
    color:#666;
    margin-bottom:20px;
    font-size:1.1em;
}
.notification-toggle{
    background:#dc3545;
    color:white;
    padding:12px 24px;
    border:none;
    border-radius:25px;
    text-align:center;
    margin:0 auto 15px;
    font-size:1em;
    cursor:pointer;
    transition:all .3s ease;
    display:block;
    font-weight:600;
}
.notification-toggle.enabled{
    background:#28a745;
}
.notification-toggle:hover{
    transform:scale(1.05);
    box-shadow:0 5px 15px rgba(0,0,0,0.3);
}
.time{
    text-align:center;
    color:#999;
    margin-bottom:15px;
    font-size:.9em;
}
.config-panel{
    margin-bottom:20px;
    padding:15px;
    border-radius:12px;
    background:#f8f9fa;
    font-size:.9em;
}
.config-panel h3{
    margin-bottom:8px;
    color:#444;
}
.config-panel label{
    display:block;
    margin-bottom:6px;
}
.config-panel input{
    width:80px;
    padding:4px 6px;
    margin-left:4px;
}
.config-save{
    margin-top:8px;
    padding:8px 16px;
    border-radius:20px;
    border:none;
    background:#007bff;
    color:white;
    cursor:pointer;
}
.config-save:hover{
    background:#0056b3;
}
.current{
    background:linear-gradient(135deg,#667eea 0%,#764ba2 100%);
    color:white;
    padding:30px;
    border-radius:15px;
    margin-bottom:20px;
}
.current h2{font-size:1.3em;margin-bottom:20px;}
.temp-big{font-size:4em;font-weight:bold;margin:20px 0;}
.details{
    display:grid;
    grid-template-columns:1fr 1fr;
    gap:15px;
    margin-top:20px;
}
.detail-item{
    background:rgba(255,255,255,0.2);
    padding:15px;
    border-radius:10px;
}
.detail-label{font-size:.9em;opacity:.9;}
.detail-value{font-size:1.3em;font-weight:bold;margin-top:5px;}
.forecast{
    display:grid;
    grid-template-columns:1fr 1fr;
    gap:20px;
}
.forecast-card{
    background:#f8f9fa;
    padding:20px;
    border-radius:15px;
    border:2px solid #e9ecef;
}
.forecast-card h3{color:#667eea;margin-bottom:10px;}
.forecast-temp{font-size:1.4em;color:#333;margin:10px 0;}
.location-btn{background:#667eea;color:white;padding:8px 16px;border:none;border-radius:20px;cursor:pointer;font-size:.9em;margin-top:10px;}
.location-btn:hover{background:#5568d3;}
.modal{display:none;position:fixed;z-index:1000;left:0;top:0;width:100%;height:100%;background:rgba(0,0,0,0.5);}
.modal-content{background:white;margin:5% auto;padding:20px;border-radius:15px;max-width:600px;width:90%;}
.modal-header{display:flex;justify-content:space-between;align-items:center;margin-bottom:15px;}
.close{font-size:28px;font-weight:bold;cursor:pointer;color:#999;}
.close:hover{color:#333;}
#map{height:400px;border-radius:10px;margin:15px 0;}
.map-buttons{display:flex;gap:10px;justify-content:flex-end;}
.btn{padding:10px 20px;border:none;border-radius:20px;cursor:pointer;font-weight:600;}
.btn-primary{background:#667eea;color:white;}
.btn-primary:hover{background:#5568d3;}
.btn-secondary{background:#6c757d;color:white;}
.btn-secondary:hover{background:#5a6268;}
.toast-container{position:fixed;top:20px;right:20px;z-index:9999;display:flex;flex-direction:column;gap:10px;}
.toast{background:white;padding:16px 20px;border-radius:10px;box-shadow:0 4px 12px rgba(0,0,0,0.15);min-width:300px;max-width:400px;display:flex;align-items:center;gap:12px;animation:slideIn 0.3s ease;}
.toast.success{border-left:4px solid #28a745;}
.toast.error{border-left:4px solid #dc3545;}
.toast.info{border-left:4px solid #007bff;}
.toast-icon{font-size:24px;}
.toast-message{flex:1;color:#333;}
.toast-close{cursor:pointer;color:#999;font-size:20px;font-weight:bold;}
.toast-close:hover{color:#333;}
@keyframes slideIn{from{transform:translateX(400px);opacity:0;}to{transform:translateX(0);opacity:1;}}
@keyframes slideOut{from{transform:translateX(0);opacity:1;}to{transform:translateX(400px);opacity:0;}}
.toast.hiding{animation:slideOut 0.3s ease;}
.footer{text-align:center;margin-top:30px;padding-top:20px;border-top:1px solid #e9ecef;color:#999;font-size:0.85em;}
</style>
</head>
<body>
<div class="toast-container" id="toastContainer"></div>
<div class="container">
    <h1>🌤️ Meteo App</h1>

    <button class="notification-toggle {{if .NotificationsEnabled}}enabled{{end}}" id="notificationToggle">
        {{if .NotificationsEnabled}}
            📢 Notifiche Attive
        {{else}}
            🔕 Notifiche Disattivate
        {{end}}
    </button>

    <div class="config-panel">
        <h3>⚙️ Configurazione notifiche</h3>
        <label>
            Intervallo (minuti):
            <input id="intervalInput" type="number" min="1" max="180" value="{{.IntervalMinutes}}">
        </label>
        <label>
            Dalle (ora):
            <input id="startHourInput" type="number" min="0" max="23" value="{{.StartHour}}">
        </label>
        <label>
            Alle (ora):
            <input id="endHourInput" type="number" min="0" max="23" value="{{.EndHour}}">
        </label>
        <button id="saveConfigBtn" class="config-save">💾 Salva configurazione</button>
    </div>

    <div class="location">
        📍 {{.City}}, {{.Country}}<br>
        <small>{{printf "%.4f" .Lat}}, {{printf "%.4f" .Lon}}</small><br>
        <button class="location-btn" id="openMapBtn">🗺️ Scegli posizione sulla mappa</button>
    </div>
    <div class="time">🕐 {{.Time}}</div>

    <div class="current">
        <h2>Condizioni Attuali</h2>
        <div>{{.CurrentCondition}}</div>
        <div class="temp-big">{{printf "%.1f" .CurrentTemp}}°C</div>

        <div class="details">
            <div class="detail-item">
                <div class="detail-label">💧 Umidità</div>
                <div class="detail-value">{{printf "%.0f" .Humidity}}%</div>
            </div>
            <div class="detail-item">
                <div class="detail-label">💨 Vento</div>
                <div class="detail-value">{{printf "%.1f" .WindSpeed}} km/h</div>
            </div>
            <div class="detail-item">
                <div class="detail-label">👁️ Visibilità</div>
                <div class="detail-value">{{printf "%.1f" .Visibility}} km</div>
            </div>
            <div class="detail-item">
                <div class="detail-label">🌧️ Precipitazioni</div>
                <div class="detail-value">{{printf "%.1f" .Precipitation}} mm</div>
            </div>
        </div>
    </div>

    <div class="forecast">
        <div class="forecast-card">
            <h3>📅 Oggi</h3>
            <div>{{.TodayCondition}}</div>
            <div class="forecast-temp">
                Max: {{printf "%.1f" .TodayMax}}°C<br>
                Min: {{printf "%.1f" .TodayMin}}°C
            </div>
        </div>
        <div class="forecast-card">
            <h3>📅 Domani</h3>
            <div>{{.TomorrowCondition}}</div>
            <div class="forecast-temp">
                Max: {{printf "%.1f" .TomorrowMax}}°C<br>
                Min: {{printf "%.1f" .TomorrowMin}}°C
            </div>
        </div>
    </div>
    
    <div class="footer">
        ⚙️ Meteo App v{{.Version}}
    </div>
</div>

<div id="mapModal" class="modal">
    <div class="modal-content">
        <div class="modal-header">
            <h2>🗺️ Scegli posizione meteo</h2>
            <span class="close" id="closeModal">&times;</span>
        </div>
        <p>Clicca sulla mappa per selezionare la posizione desiderata</p>
        <div id="map"></div>
        <div class="map-buttons">
            <button class="btn btn-secondary" id="resetLocationBtn">🔄 Usa posizione automatica</button>
            <button class="btn btn-primary" id="saveLocationBtn">💾 Salva posizione</button>
        </div>
    </div>
</div>

<script>
const toggleBtn = document.getElementById("notificationToggle");
const intervalInput = document.getElementById("intervalInput");
const startHourInput = document.getElementById("startHourInput");
const endHourInput = document.getElementById("endHourInput");
const saveConfigBtn = document.getElementById("saveConfigBtn");
const openMapBtn = document.getElementById("openMapBtn");
const mapModal = document.getElementById("mapModal");
const closeModal = document.getElementById("closeModal");
const saveLocationBtn = document.getElementById("saveLocationBtn");
const resetLocationBtn = document.getElementById("resetLocationBtn");

let map;
let marker;
let selectedLat = {{.Lat}};
let selectedLon = {{.Lon}};

// Sistema toast
function showToast(message, type = "info") {
    const container = document.getElementById("toastContainer");
    const toast = document.createElement("div");
    toast.className = "toast " + type;
    
    const icons = {
        success: "✅",
        error: "❌",
        info: "ℹ️"
    };
    
    const icon = icons[type] || icons.info;
    toast.innerHTML = "<span class=\"toast-icon\">" + icon + "</span>" +
        "<span class=\"toast-message\">" + message + "</span>" +
        "<span class=\"toast-close\" onclick=\"this.parentElement.remove()\">×</span>";
    
    container.appendChild(toast);
    
    setTimeout(() => {
        toast.classList.add("hiding");
        setTimeout(() => toast.remove(), 300);
    }, 4000);
}

// Inizializza mappa
function initMap() {
    if (!map) {
        map = L.map("map").setView([selectedLat, selectedLon], 10);
        L.tileLayer("https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png", {
            attribution: "© OpenStreetMap contributors"
        }).addTo(map);
        
        marker = L.marker([selectedLat, selectedLon], {draggable: true}).addTo(map);
        
        map.on("click", function(e) {
            selectedLat = e.latlng.lat;
            selectedLon = e.latlng.lng;
            marker.setLatLng(e.latlng);
        });
        
        marker.on("dragend", function(e) {
            const pos = marker.getLatLng();
            selectedLat = pos.lat;
            selectedLon = pos.lng;
        });
    }
}

// Apri modale mappa
openMapBtn.addEventListener("click", () => {
    mapModal.style.display = "block";
    setTimeout(() => {
        initMap();
        map.invalidateSize();
    }, 100);
});

// Chiudi modale
closeModal.addEventListener("click", () => {
    mapModal.style.display = "none";
});

window.addEventListener("click", (e) => {
    if (e.target === mapModal) {
        mapModal.style.display = "none";
    }
});

// Salva posizione personalizzata
saveLocationBtn.addEventListener("click", async () => {
    try {
        const res = await fetch("/meteo/location/set", {
            method: "POST",
            headers: {"Content-Type": "application/json"},
            body: JSON.stringify({lat: selectedLat, lon: selectedLon})
        });
        if (!res.ok) throw new Error("Errore salvataggio posizione");
        showToast("Posizione salvata! Ricaricamento...", "success");
        mapModal.style.display = "none";
        setTimeout(() => location.reload(), 1500);
    } catch (e) {
        console.error(e);
        showToast("Errore: " + e.message, "error");
    }
});

// Ripristina posizione automatica
resetLocationBtn.addEventListener("click", async () => {
    try {
        const res = await fetch("/meteo/location/reset", {method: "POST"});
        if (!res.ok) throw new Error("Errore reset posizione");
        showToast("Posizione automatica ripristinata! Ricaricamento...", "success");
        mapModal.style.display = "none";
        setTimeout(() => location.reload(), 1500);
    } catch (e) {
        console.error(e);
        showToast("Errore: " + e.message, "error");
    }
});

toggleBtn.addEventListener("click", async () => {
    try {
        const res = await fetch("/meteo/toggle-notification", { method: "POST" });
        if (!res.ok) throw new Error("Errore server");
        const data = await res.json();
        if (data.enabled) {
            toggleBtn.classList.add("enabled");
            toggleBtn.textContent = "📢 Notifiche Attive";
            showToast("Notifiche attivate con successo", "success");
        } else {
            toggleBtn.classList.remove("enabled");
            toggleBtn.textContent = "🔕 Notifiche Disattivate";
            showToast("Notifiche disattivate", "info");
        }
    } catch (e) {
        console.error(e);
        showToast("Errore toggle notifiche: " + e.message, "error");
    }
});

saveConfigBtn.addEventListener("click", async () => {
    try {
        const payload = {
            interval_minutes: parseInt(intervalInput.value, 10),
            start_hour: parseInt(startHourInput.value, 10),
            end_hour: parseInt(endHourInput.value, 10)
        };
        const res = await fetch("/meteo/config/update", {
            method: "POST",
            headers: {"Content-Type": "application/json"},
            body: JSON.stringify(payload)
        });
        if (!res.ok) throw new Error("Errore salvataggio");
        const cfg = await res.json();
        intervalInput.value = cfg.interval_minutes;
        startHourInput.value = cfg.start_hour;
        endHourInput.value = cfg.end_hour;
        showToast("Configurazione aggiornata con successo", "success");
    } catch (e) {
        console.error(e);
        showToast("Errore configurazione: " + e.message, "error");
    }
});
</script>
</body>
</html>
`

// homeHandler gestisce la pagina principale con l'interfaccia utente
func homeHandler(w http.ResponseWriter, r *http.Request) {
	data, err := getWeather()
	if err != nil {
		http.Error(w, "Errore meteo: "+err.Error(), http.StatusInternalServerError)
		return
	}

	t := template.Must(template.New("weather").Parse(htmlTemplate))
	_ = t.Execute(w, data)
}
