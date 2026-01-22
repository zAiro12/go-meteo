package main

import (
	"html/template"
	"net/http"
	"net/url"
	"strconv"
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
/* Header layout: title left, auth icon button right */
.header{display:flex;align-items:center;justify-content:space-between;margin-bottom:10px}
.auth-btn{background:transparent;border:none;font-size:1.6em;cursor:pointer;padding:6px;border-radius:8px}
.auth-btn:hover{background:rgba(0,0,0,0.05)}
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
/* Telegram input & button styles */
.user-panel{display:flex;flex-direction:column;gap:8px;margin-bottom:18px}
.input-row{display:flex;align-items:center;gap:12px;width:100%}
.telegram-input{padding:12px 16px;border-radius:12px;border:1px solid #e6eef5;background:#fff;min-width:160px;flex:1;font-size:1rem;box-shadow:0 6px 18px rgba(0,0,0,0.06)}
.btn-telegram{background:#0088cc;color:white;padding:12px 22px;border-radius:14px;border:none;cursor:pointer;font-weight:700;font-size:1rem;box-shadow:0 8px 20px rgba(2,136,204,0.18);min-height:54px;display:inline-flex;align-items:center;justify-content:center}
.btn-telegram:hover{background:#007ab8;transform:translateY(-2px)}
.center-actions{width:100%;display:flex;justify-content:center;margin-top:8px}
.notification-toggle.user-toggle{padding:12px 20px;border-radius:14px;min-height:54px}

@media (max-width:700px){
    .input-row{flex-direction:column;align-items:stretch}
    .btn-telegram{width:100%}
    .notification-toggle.user-toggle{width:100%}
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
    <div class="header">
        <h1>🌤️ Meteo App</h1>
        <div id="authArea" style="display:flex;align-items:center;gap:8px;">
            <span id="userNameDisplay" style="font-size:0.95em;color:#444;display:none"></span>
            <button id="openAuthBtn" class="auth-btn" title="Accedi o Registrati">👤</button>
            <button id="logoutBtn" class="auth-btn" title="Logout" style="display:none">🔓</button>
        </div>
    </div>

    

    <!-- Login admin per mostrare configurazione notifiche -->
    <div id="adminLoginPanel" style="display:none; margin-bottom:20px;">
        <input type="password" id="adminPasswordInput" placeholder="Password admin" />
        <button id="adminLoginBtn" class="btn btn-primary">Login Admin</button>
    </div>
    <div id="adminConfigPanel" class="config-panel" style="display:none;">
        <h3>⚙️ Configurazione notifiche (admin)</h3>
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

        <!-- Pulsante login (icona) posizionato nella header sopra -->

        <!-- Auth Modal -->
        <div id="authModal" class="modal">
            <div class="modal-content">
                <div class="modal-header">
                    <h2>Accesso / Registrazione</h2>
                    <span class="close" id="authClose">&times;</span>
                </div>
                <div style="padding:10px;">
                    <label>Username: <input id="authUsername" type="text" /></label><br/><br/>
                    <label>Password: <input id="authPassword" type="password" /></label><br/><br/>
                    <button id="authLoginBtn" class="btn btn-primary">Accedi</button>
                    <button id="authRegisterBtn" class="btn btn-secondary">Registrati</button>
                </div>
            </div>
        </div>
    <div id="userPanel" style="display:none;">
        <div class="user-panel">
            <label style="font-weight:600;color:#333">Telegram ID:</label>
            <div class="input-row">
                <input id="telegramInput" type="text" class="telegram-input" />
                <button id="saveTelegramBtn" class="btn-telegram">Salva Telegram</button>
            </div>
            <div class="center-actions">
                <button id="userNotificationToggle" class="notification-toggle user-toggle" style="padding:10px 18px;font-size:0.95em">Attiva notifiche</button>
            </div>
        </div>
    </div>

    <!-- Meteo sempre visibile per città client di default -->

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
// If page was rendered with preview params, remove them from the URL
(function(){
    try {
        const params = new URLSearchParams(window.location.search);
        if (params.has('preview_lat') && params.has('preview_lon')) {
            const newUrl = window.location.pathname + window.location.hash;
            history.replaceState(null, '', newUrl);
        }
    } catch (e) {
        // ignore older browsers
    }
})();

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

// Salva posizione personalizzata: comportamento differenziato
saveLocationBtn.addEventListener("click", async () => {
    try {
        // Se non loggato -> anteprima (non salva): mostra meteo per le coordinate selezionate
        if (!currentUser) {
            mapModal.style.display = "none";
            showToast("Anteprima: la posizione non verrà salvata. Effettua il login per salvarla.", "info");
            const url = '/?preview_lat=' + encodeURIComponent(selectedLat) + '&preview_lon=' + encodeURIComponent(selectedLon);
            window.location.href = url;
            return;
        }

        // Se utente loggato e admin -> salva posizione globale
        if (currentUser.is_admin) {
            const res = await fetch("/location/set", {
                method: "POST",
                headers: {"Content-Type": "application/json"},
                body: JSON.stringify({lat: selectedLat, lon: selectedLon, admin_username: currentUser.username})
            });
            if (!res.ok) throw new Error("Errore salvataggio posizione admin");
            // persist username so session is restored after reload
            try { if (currentUser && currentUser.username) localStorage.setItem('meteo_username', currentUser.username); } catch(e){}
            showToast("Posizione globale aggiornata! Ricaricamento...", "success");
            mapModal.style.display = "none";
            setTimeout(() => { window.location.href = '/'; }, 1200);
            return;
        }

        // Utente normale -> salva solo per il profilo utente
        const payload = {...currentUser, lat: selectedLat, lon: selectedLon};
        const res = await fetch("/meteo/user/update", {
            method: "POST",
            headers: {"Content-Type": "application/json"},
            body: JSON.stringify(payload)
        });
        if (!res.ok) throw new Error("Errore salvataggio posizione utente");
        // ensure session persistence and reload to show new default for this user
        try { if (currentUser && currentUser.username) localStorage.setItem('meteo_username', currentUser.username); } catch(e){}
        // Recupera il nome della città e della provincia via reverse geocoding per mostrare il nome nella home
        let cityName = '';
        let provinceName = '';
        let countryName = '';
        try {
            const rev = await fetch('https://nominatim.openstreetmap.org/reverse?format=json&lat=' + encodeURIComponent(selectedLat) + '&lon=' + encodeURIComponent(selectedLon));
            if (rev.ok) {
                const j = await rev.json();
                if (j && j.address) {
                    cityName = j.address.city || j.address.town || j.address.village || j.address.hamlet || '';
                    provinceName = j.address.county || j.address.state || j.address.region || j.address.province || '';
                    countryName = j.address.country || j.address.country_code || '';
                }
            }
        } catch (e) { /* ignore reverse geocode errors */ }

        // Convert province full name to common 2-letter code when possible (e.g., Milano -> MI)
        function provinceCodeFromName(name) {
            if (!name) return '';
            const n = name.toLowerCase();
            const map = {
                'milano':'MI','milan':'MI','roma':'RM','torino':'TO','torino città':'TO','napoli':'NA','naples':'NA','genova':'GE','venezia':'VE','venice':'VE','verona':'VR','bari':'BA','palermo':'PA','catania':'CT','firenze':'FI','florence':'FI','bologna':'BO','cagliari':'CA','trento':'TN','bolzano':'BZ','brescia':'BS','bergamo':'BG','monza':'MB','modena':'MO','padova':'PD','vicenza':'VI','como':'CO','pavia':'PV','messina':'ME','taranto':'TA','perugia':'PG','ancona':'AN','siena':'SI','arezzo':'AR','lecce':'LE','salerno':'SA','reggio calabria':'RC','cosenza':'CS','veneto':'VE'
            };
            for (const k in map) {
                if (n.includes(k)) return map[k];
            }
            // fallback: take first two consonant letters of name
            const cleaned = n.replace(/[^a-z]/g, '');
            if (cleaned.length >= 2) return cleaned.substr(0,2).toUpperCase();
            return cleaned.toUpperCase();
        }
        const provinceCode = provinceCodeFromName(provinceName);
        showToast("Posizione salvata sul profilo personale", "success");
        mapModal.style.display = "none";
        // Ricarica la home mostrando subito la città e la provincia (se trovate) o le coordinate
        setTimeout(() => {
            let url = '/?preview_lat=' + encodeURIComponent(selectedLat) + '&preview_lon=' + encodeURIComponent(selectedLon);
            if (cityName) url += '&preview_city=' + encodeURIComponent(cityName);
            if (provinceCode) url += '&preview_province=' + encodeURIComponent(provinceCode);
            if (countryName) url += '&preview_country=' + encodeURIComponent(countryName);
            window.location.href = url;
        }, 900);
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

// --- INIZIO LOGICA UI AUTENTICAZIONE E PROFILO ---

// Admin login
const adminLoginPanel = document.getElementById("adminLoginPanel");
const adminConfigPanel = document.getElementById("adminConfigPanel");
const adminPasswordInput = document.getElementById("adminPasswordInput");
const adminLoginBtn = document.getElementById("adminLoginBtn");

adminLoginPanel.style.display = "none";
adminLoginBtn.addEventListener("click", async () => {
    // legacy: attempt admin login via header password (kept for compatibility)
    const pass = adminPasswordInput.value;
    const res = await fetch("/meteo/admin/notifications", {headers: {"X-Admin-Password": pass}});
    if (res.ok) {
        adminConfigPanel.style.display = "block";
        adminLoginPanel.style.display = "none";
        showToast("Accesso admin riuscito", "success");
    } else {
        showToast("Password admin errata", "error");
    }
});

// Auth modal and user profile management
const openAuthBtn = document.getElementById("openAuthBtn");
const authModal = document.getElementById("authModal");
const authClose = document.getElementById("authClose");
const authUsername = document.getElementById("authUsername");
const authPassword = document.getElementById("authPassword");
const authLoginBtn = document.getElementById("authLoginBtn");
const authRegisterBtn = document.getElementById("authRegisterBtn");
const logoutBtn = document.getElementById("logoutBtn");
const userNameDisplay = document.getElementById("userNameDisplay");

// Restore session from localStorage (username) on page load
async function restoreSession() {
    const stored = localStorage.getItem('meteo_username');
    if (!stored) return;
    try {
        const res = await fetch('/meteo/user/profile?username=' + encodeURIComponent(stored));
        if (!res.ok) { localStorage.removeItem('meteo_username'); return; }
        const user = await res.json();
        currentUser = user;
        telegramInput.value = user.telegram_user || '';
        userNotificationToggle.textContent = user.notify ? 'Disattiva notifiche Telegram' : 'Attiva notifiche Telegram';
        userPanel.style.display = 'block';
        openAuthBtn.style.display = 'none';
        logoutBtn.style.display = 'inline-block';
        userNameDisplay.style.display = 'inline';
        userNameDisplay.textContent = user.username;
        // per-user toggle aggiornato più in basso (userNotificationToggle)
        if (user.is_admin) {
            adminConfigPanel.style.display = 'block';
        } else {
            adminConfigPanel.style.display = 'none';
        }
        // If the page wasn't rendered for this user, reload so server shows user's saved city
        try {
            const params = new URLSearchParams(window.location.search);
            if (!params.has('username')) {
                window.location.href = '/?username=' + encodeURIComponent(user.username);
                return;
            }
        } catch (e) {}
    } catch (e) {
        localStorage.removeItem('meteo_username');
    }
}

// try restore immediately
restoreSession();

const userPanel = document.getElementById("userPanel");
const telegramInput = document.getElementById("telegramInput");
const saveTelegramBtn = document.getElementById("saveTelegramBtn");
const userNotificationToggle = document.getElementById("userNotificationToggle");

let currentUser = null;

openAuthBtn.addEventListener("click", () => {
    authModal.style.display = "block";
    authPassword.value = "";
});
authClose.addEventListener("click", () => { authModal.style.display = "none"; });
window.addEventListener("click", (e) => { if (e.target === authModal) authModal.style.display = "none"; });

authRegisterBtn.addEventListener("click", async () => {
    const username = authUsername.value.trim();
    const password = authPassword.value;
    if (!username || !password) { showToast("Inserisci username e password", "error"); return; }
    try {
        const res = await fetch('/meteo/auth/register', {
            method: 'POST', headers: {'Content-Type':'application/json'},
            body: JSON.stringify({username, password})
        });
        if (res.status === 201) {
            showToast('Registrazione avvenuta. Effettua il login.', 'success');
            authPassword.value = '';
            // leave username prefilled for login
        } else {
            const txt = await res.text();
            showToast('Errore registrazione: '+txt, 'error');
        }
    } catch (e) { showToast('Errore registrazione', 'error'); }
});

authLoginBtn.addEventListener("click", async () => {
    const username = authUsername.value.trim();
    const password = authPassword.value;
    if (!username || !password) { showToast("Inserisci username e password", "error"); return; }
    try {
        const res = await fetch('/meteo/auth/login', {
            method: 'POST', headers: {'Content-Type':'application/json'},
            body: JSON.stringify({username, password})
        });
        if (res.ok) {
            const user = await res.json();
            currentUser = user;
            // show profile area
            telegramInput.value = user.telegram_user || '';
            userNotificationToggle.textContent = user.notify ? 'Disattiva notifiche Telegram' : 'Attiva notifiche Telegram';
            userPanel.style.display = 'block';
            // update header: hide login, show logout and user name
            openAuthBtn.style.display = 'none';
            // persist session
            try { localStorage.setItem('meteo_username', user.username); } catch(e){}
            logoutBtn.style.display = 'inline-block';
            userNameDisplay.style.display = 'inline';
            userNameDisplay.textContent = user.username;
                    // mostra solo il toggle per utente (userNotificationToggle) già aggiornato sopra
                    // if user is admin, additionally show admin config
                    if (user.is_admin) {
                        adminConfigPanel.style.display = 'block';
                    } else {
                        adminConfigPanel.style.display = 'none';
                    }
            authModal.style.display = 'none';
            showToast('Login utente riuscito', 'success');
            // ricarica la pagina con username per mostrare la città salvata dal server
            try { window.location.href = '/?username=' + encodeURIComponent(user.username); } catch(e){}
        } else {
            showToast('Credenziali errate', 'error');
        }
    } catch (e) { showToast('Errore login', 'error'); }
});

// Logout client-side
logoutBtn.addEventListener('click', async () => {
    try {
        // optional server logout call
        await fetch('/meteo/user/logout', {method: 'POST'}).catch(()=>{});
    } catch (e) {}
    currentUser = null;
    try { localStorage.removeItem('meteo_username'); } catch(e) {}
    userPanel.style.display = 'none';
    adminConfigPanel.style.display = 'none';
    // rimuove riferimento a toggle globale (usiamo solo il toggle per utente)
    openAuthBtn.style.display = 'inline-block';
    logoutBtn.style.display = 'none';
    userNameDisplay.style.display = 'none';
    userNameDisplay.textContent = '';
    showToast('Logout effettuato', 'info');
    // dopo logout ricarica la home senza parametri in modo che venga mostrata la città di default (admin)
    setTimeout(() => { window.location.href = '/'; }, 500);
});

saveTelegramBtn.addEventListener("click", async () => {
    if (!currentUser) return;
    const telegram_user = telegramInput.value.trim();
    const payload = {...currentUser, telegram_user};
    const res = await fetch("/meteo/user/update", {
        method: "POST",
        headers: {"Content-Type": "application/json"},
        body: JSON.stringify(payload)
    });
    if (res.ok) {
        showToast("Telegram ID salvato", "success");
        currentUser.telegram_user = telegram_user;
    } else {
        showToast("Errore salvataggio Telegram", "error");
    }
});

userNotificationToggle.addEventListener("click", async () => {
    if (!currentUser) return;
    const payload = {...currentUser, notify: !currentUser.notify};
    const res = await fetch("/meteo/user/update", {
        method: "POST",
        headers: {"Content-Type": "application/json"},
        body: JSON.stringify(payload)
    });
    if (res.ok) {
        currentUser.notify = !currentUser.notify;
        userNotificationToggle.textContent = currentUser.notify ? "Disattiva notifiche Telegram" : "Attiva notifiche Telegram";
        showToast(currentUser.notify ? "Notifiche attivate" : "Notifiche disattivate", "success");
    } else {
        showToast("Errore aggiornamento notifiche", "error");
    }
});

// --- FINE LOGICA UI AUTENTICAZIONE E PROFILO ---

</script>
</body>
</html>
`

// homeHandler gestisce la pagina principale con l'interfaccia utente
func renderTemplate(w http.ResponseWriter, data *WeatherData) error {
	t := template.Must(template.New("weather").Parse(htmlTemplate))
	return t.Execute(w, data)
}

func tryRenderForUser(w http.ResponseWriter, q url.Values) bool {
	username := q.Get("username")
	if username == "" {
		return false
	}
	u, _ := GetUserByUsername(username)
	if u == nil {
		return false
	}
	if u.Lat != 0 || u.Lon != 0 {
		data, err := getWeatherFor(u.Lat, u.Lon)
		if err == nil {
			if u.City != "" {
				data.City = u.City
			}
			_ = renderTemplate(w, data)
			return true
		}
	}
	return false
}

func tryRenderPreview(w http.ResponseWriter, q url.Values) bool {
	if q.Get("preview_lat") == "" || q.Get("preview_lon") == "" {
		return false
	}
	lat, err := strconv.ParseFloat(q.Get("preview_lat"), 64)
	if err != nil {
		return false
	}
	lon, err := strconv.ParseFloat(q.Get("preview_lon"), 64)
	if err != nil {
		return false
	}
	data, err := getWeatherFor(lat, lon)
	if err != nil {
		return false
	}
	if q.Get("preview_city") != "" {
		city := q.Get("preview_city")
		if q.Get("preview_province") != "" {
			city = city + " (" + q.Get("preview_province") + ")"
		}
		data.City = city
	}
	if q.Get("preview_country") != "" {
		data.Country = q.Get("preview_country")
	}
	_ = renderTemplate(w, data)
	return true
}

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	// 1) show user's saved city if requested
	if tryRenderForUser(w, q) {
		return
	}
	// 2) preview (anonymous)
	if tryRenderPreview(w, q) {
		return
	}
	// 3) default weather
	data, err := getWeather()
	if err != nil {
		http.Error(w, "Errore meteo: "+err.Error(), http.StatusInternalServerError)
		return
	}
	// If there is no admin default location configured, and we're on the public homepage
	// (no preview and no username), do not show a perceived city from IP geolocation.
	// Leave City/Country empty so the homepage does not display the last user's city.
	if q.Get("username") == "" && q.Get("preview_lat") == "" && q.Get("preview_lon") == "" {
		locationMutex.RLock()
		localUseCustom := useCustom
		locationMutex.RUnlock()
		if !localUseCustom {
			data.City = ""
			data.Country = ""
		}
	}
	_ = renderTemplate(w, data)
}
