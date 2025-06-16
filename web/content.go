package web

// HTMLContent contains the dashboard HTML
const HTMLContent = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Syslog Analyzer Dashboard</title>
    <link rel="stylesheet" href="/static/style.css">
</head>
<body>
    <div class="container">
        <header>
            <h1>Professional Syslog Analyzer</h1>
            <div class="status-indicator">
                <div class="status-dot" id="connectionStatus"></div>
                <span id="statusText">Connecting...</span>
            </div>
        </header>

        <div class="dashboard">
            <div class="global-metrics">
                <h2>Global Summary</h2>
                <div class="metrics-grid">
                    <div class="metric-card">
                        <h3>Total Real-time EPS</h3>
                        <div class="metric-value" id="globalEPS">0</div>
                    </div>
                    <div class="metric-card">
                        <h3>Total Real-time GB/s</h3>
                        <div class="metric-value" id="globalGBps">0.000000</div>
                    </div>
                    <div class="metric-card">
                        <h3>Total Hourly Avg GB</h3>
                        <div class="metric-value" id="totalHourlyAvg">0.00000</div>
                    </div>
                    <div class="metric-card">
                        <h3>Total Daily Avg GB</h3>
                        <div class="metric-value" id="totalDailyAvg">0.00000</div>
                    </div>
                    <div class="metric-card">
                        <h3>Active Sources</h3>
                        <div class="metric-value" id="activeSources">0</div>
                    </div>
                    <div class="metric-card">
                        <h3>Total Sources</h3>
                        <div class="metric-value" id="totalSources">0</div>
                    </div>
                </div>
            </div>

            <div class="sources-section">
                <div class="section-header">
                    <h2>Syslog Sources</h2>
                    <div class="actions">
                        <button onclick="generateReport()" class="btn btn-secondary">Export Report</button>
                        <button onclick="showAddSourceModal()" class="btn btn-primary">Add Source</button>
                    </div>
                </div>
                
                <div class="sources-table">
                    <table id="sourcesTable">
                        <thead>
                            <tr>
                                <th>Source</th>
                                <th>Real-time Metrics</th>
                                <th>Hourly Averages</th>
                                <th>Daily Averages</th>
                                <th>Actions</th>
                            </tr>
                        </thead>
                        <tbody id="sourcesTableBody">
                        </tbody>
                    </table>
                </div>
            </div>
        </div>
    </div>

    <!-- Add Source Modal -->
    <div id="addSourceModal" class="modal">
        <div class="modal-content">
            <div class="modal-header">
                <h3>Add New Syslog Source</h3>
                <span class="close" onclick="hideAddSourceModal()">&times;</span>
            </div>
            <form id="addSourceForm">
                <div class="form-group">
                    <label for="sourceName">Source Name:</label>
                    <input type="text" id="sourceName" required>
                </div>
                <div class="form-group">
                    <label for="sourceIP">Source IP Address:</label>
                    <input type="text" id="sourceIP" required placeholder="192.168.1.100">
                </div>
                <div class="form-group">
                    <label for="sourcePort">Port:</label>
                    <input type="number" id="sourcePort" value="514" min="1" max="65535" required>
                </div>
                <div class="form-group">
                    <label for="sourceProtocol">Protocol:</label>
                    <select id="sourceProtocol" required>
                        <option value="UDP" selected>UDP</option>
                        <option value="TCP">TCP</option>
                    </select>
                </div>
                
                <!-- Destinations Section -->
                <div class="form-group">
                    <label>Destinations:</label>
                    <div id="destinationsContainer">
                        <!-- Destinations will be added here dynamically -->
                    </div>
                    <button type="button" onclick="addDestination()" class="btn btn-secondary btn-small">Add Destination</button>
                </div>
                
                <div class="form-group">
                    <label for="simulationMode">Simulation Mode:</label>
                    <div class="toggle-container">
                        <input type="checkbox" id="simulationMode" class="toggle-input" checked>
                        <label for="simulationMode" class="toggle-label">
                            <span class="toggle-slider"></span>
                            <span class="toggle-text">
                                <span class="on-text">ON</span>
                                <span class="off-text">OFF</span>
                            </span>
                        </label>
                    </div>
                </div>
                <div class="form-actions">
                    <button type="button" onclick="hideAddSourceModal()" class="btn btn-secondary">Cancel</button>
                    <button type="submit" class="btn btn-primary">Add Source</button>
                </div>
            </form>
        </div>
    </div>

    <script src="/static/app.js"></script>
</body>
</html>`

// CSSContent contains the dashboard CSS
const CSSContent = `* {
    margin: 0;
    padding: 0;
    box-sizing: border-box;
}

body {
    font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
    background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
    min-height: 100vh;
    color: #333;
}

.container {
    max-width: 1400px;
    margin: 0 auto;
    padding: 20px;
}

header {
    background: rgba(255, 255, 255, 0.95);
    padding: 20px 30px;
    border-radius: 15px;
    margin-bottom: 20px;
    backdrop-filter: blur(10px);
    box-shadow: 0 8px 32px rgba(0, 0, 0, 0.1);
    display: flex;
    justify-content: space-between;
    align-items: center;
}

header h1 {
    color: #2c3e50;
    font-size: 2rem;
    font-weight: 600;
}

.status-indicator {
    display: flex;
    align-items: center;
    gap: 10px;
}

.status-dot {
    width: 12px;
    height: 12px;
    border-radius: 50%;
    background: #e74c3c;
    animation: pulse 2s infinite;
}

.status-dot.connected {
    background: #2ecc71;
}

@keyframes pulse {
    0% { opacity: 1; }
    50% { opacity: 0.5; }
    100% { opacity: 1; }
}

.dashboard {
    display: grid;
    gap: 20px;
}

.global-metrics {
    background: rgba(255, 255, 255, 0.95);
    padding: 25px;
    border-radius: 15px;
    backdrop-filter: blur(10px);
    box-shadow: 0 8px 32px rgba(0, 0, 0, 0.1);
}

.global-metrics h2 {
    color: #2c3e50;
    margin-bottom: 20px;
    font-size: 1.5rem;
}

.metrics-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
    gap: 15px;
}

.metric-card {
    background: linear-gradient(135deg, #74b9ff, #0984e3);
    padding: 20px;
    border-radius: 12px;
    color: white;
    text-align: center;
    transition: transform 0.3s ease;
}

.metric-card:hover {
    transform: translateY(-5px);
}

.metric-card h3 {
    font-size: 0.9rem;
    margin-bottom: 10px;
    opacity: 0.9;
}

.metric-value {
    font-size: 1.8rem;
    font-weight: bold;
}

.sources-section {
    background: rgba(255, 255, 255, 0.95);
    padding: 25px;
    border-radius: 15px;
    backdrop-filter: blur(10px);
    box-shadow: 0 8px 32px rgba(0, 0, 0, 0.1);
}

.section-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 20px;
}

.section-header h2 {
    color: #2c3e50;
    font-size: 1.5rem;
}

.actions {
    display: flex;
    gap: 10px;
}

.btn {
    padding: 10px 20px;
    border: none;
    border-radius: 8px;
    cursor: pointer;
    font-weight: 500;
    transition: all 0.3s ease;
    text-decoration: none;
    display: inline-block;
}

.btn-primary {
    background: linear-gradient(135deg, #6c5ce7, #a29bfe);
    color: white;
}

.btn-primary:hover {
    transform: translateY(-2px);
    box-shadow: 0 5px 15px rgba(108, 92, 231, 0.3);
}

.btn-secondary {
    background: linear-gradient(135deg, #74b9ff, #0984e3);
    color: white;
}

.btn-secondary:hover {
    transform: translateY(-2px);
    box-shadow: 0 5px 15px rgba(116, 185, 255, 0.3);
}

.btn-danger {
    background: linear-gradient(135deg, #fd79a8, #e84393);
    color: white;
    padding: 5px 15px;
    font-size: 0.8rem;
}

.btn-small {
    padding: 8px 16px;
    font-size: 0.9rem;
}

.sources-table {
    overflow-x: auto;
}

table {
    width: 100%;
    border-collapse: collapse;
    margin-top: 10px;
}

th, td {
    padding: 15px;
    text-align: left;
    border-bottom: 1px solid #ecf0f1;
}

th {
    background: linear-gradient(135deg, #ddd6fe, #c7d2fe);
    color: #2c3e50;
    font-weight: 600;
}

.source-info {
    display: flex;
    flex-direction: column;
    gap: 5px;
}

.source-name {
    font-weight: 600;
    color: #2c3e50;
}

.source-address {
    font-size: 0.9rem;
    color: #7f8c8d;
}

.metrics-column {
    display: flex;
    flex-direction: column;
    gap: 3px;
    min-width: 120px;
}

.metric-row {
    display: flex;
    justify-content: space-between;
    padding: 2px 0;
}

.metric-label {
    font-size: 0.85rem;
    color: #7f8c8d;
}

.metric-number {
    font-weight: 600;
    color: #2c3e50;
}

.status-badge {
    padding: 4px 12px;
    border-radius: 20px;
    font-size: 0.8rem;
    font-weight: 500;
}

.status-active {
    background: #d5f5d7;
    color: #2ecc71;
}

.status-idle {
    background: #fff3cd;
    color: #856404;
}

.status-inactive {
    background: #ffeaa7;
    color: #e17055;
}

/* Modal Styles */
.modal {
    display: none;
    position: fixed;
    z-index: 1000;
    left: 0;
    top: 0;
    width: 100%;
    height: 100%;
    background-color: rgba(0, 0, 0, 0.5);
    backdrop-filter: blur(5px);
    overflow-y: auto;
    padding: 20px 0;
}

.modal-content {
    background: white;
    margin: 0 auto;
    padding: 0;
    border-radius: 15px;
    width: 90%;
    max-width: 600px;
    box-shadow: 0 10px 50px rgba(0, 0, 0, 0.3);
    position: relative;
    top: 50%;
    transform: translateY(-50%);
    max-height: 90vh;
    overflow-y: auto;
}

.modal-header {
    background: linear-gradient(135deg, #6c5ce7, #a29bfe);
    color: white;
    padding: 20px 25px;
    border-radius: 15px 15px 0 0;
    display: flex;
    justify-content: space-between;
    align-items: center;
}

.modal-header h3 {
    margin: 0;
    font-size: 1.3rem;
}

.close {
    color: white;
    font-size: 28px;
    font-weight: bold;
    cursor: pointer;
    line-height: 1;
}

.close:hover {
    opacity: 0.7;
}

form {
    padding: 25px;
}

.form-group {
    margin-bottom: 20px;
}

.form-group label {
    display: block;
    margin-bottom: 8px;
    font-weight: 500;
    color: #2c3e50;
}

.form-group input,
.form-group select {
    width: 100%;
    padding: 12px;
    border: 2px solid #ecf0f1;
    border-radius: 8px;
    font-size: 1rem;
    transition: border-color 0.3s ease;
}

.form-group input:focus,
.form-group select:focus {
    outline: none;
    border-color: #6c5ce7;
}

.form-group input[type="checkbox"] {
    width: auto;
    margin-right: 8px;
}

.form-actions {
    display: flex;
    gap: 15px;
    justify-content: flex-end;
    margin-top: 25px;
    padding-top: 20px;
    border-top: 1px solid #ecf0f1;
}

/* Toggle Switch Styles */
.toggle-container {
    position: relative;
    display: inline-block;
}

.toggle-input {
    display: none;
}

.toggle-label {
    display: block;
    width: 80px;
    height: 40px;
    background-color: #e74c3c;
    border-radius: 20px;
    position: relative;
    cursor: pointer;
    transition: background-color 0.3s ease;
    user-select: none;
}

.toggle-input:checked + .toggle-label {
    background-color: #2ecc71;
}

.toggle-slider {
    position: absolute;
    top: 3px;
    left: 3px;
    width: 34px;
    height: 34px;
    background-color: white;
    border-radius: 50%;
    transition: transform 0.3s ease;
    box-shadow: 0 2px 5px rgba(0, 0, 0, 0.2);
}

.toggle-input:checked + .toggle-label .toggle-slider {
    transform: translateX(40px);
}

.toggle-text {
    position: absolute;
    top: 50%;
    transform: translateY(-50%);
    font-size: 12px;
    font-weight: bold;
    color: white;
}

.on-text {
    left: 8px;
    opacity: 0;
    transition: opacity 0.3s ease;
}

.off-text {
    right: 8px;
    opacity: 1;
    transition: opacity 0.3s ease;
}

.toggle-input:checked + .toggle-label .on-text {
    opacity: 1;
}

.toggle-input:checked + .toggle-label .off-text {
    opacity: 0;
}

.btn-action {
    width: 70px !important;
    height: 35px !important;
    font-size: 0.85rem !important;
    padding: 8px 12px !important;
    display: inline-flex !important;
    align-items: center !important;
    justify-content: center !important;
    text-align: center !important;
    min-width: 70px !important;
    max-width: 70px !important;
    min-height: 35px !important;
    max-height: 35px !important;
    border: none !important;
    border-radius: 8px !important;
    cursor: pointer !important;
    font-weight: 500 !important;
    transition: all 0.3s ease !important;
    text-decoration: none !important;
    box-sizing: border-box !important;
    line-height: 1 !important;
    vertical-align: middle !important;
}

.btn-action.btn-secondary {
    background: linear-gradient(135deg, #74b9ff, #0984e3) !important;
    color: white !important;
}

.btn-action.btn-danger {
    background: linear-gradient(135deg, #fd79a8, #e84393) !important;
    color: white !important;
}

.btn-action:hover {
    transform: translateY(-2px) !important;
}

.btn-action.btn-secondary:hover {
    box-shadow: 0 5px 15px rgba(116, 185, 255, 0.3) !important;
}

.btn-action.btn-danger:hover {
    box-shadow: 0 5px 15px rgba(253, 121, 168, 0.3) !important;
}

.button-group {
    display: flex;
    gap: 8px;
    align-items: center;
    justify-content: center;
}

.form-group input[type="checkbox"]:not(.toggle-input) {
    width: auto;
    margin-right: 8px;
}

.form-group select:disabled {
    background-color: #f8f9fa;
    color: #6c757d;
    cursor: not-allowed;
    opacity: 0.6;
}

/* Destination Styles */
.destination-item {
    background: rgba(255, 255, 255, 0.95);
    border-radius: 8px;
    padding: 15px;
    margin-bottom: 15px;
    box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}

.destination-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 10px;
}

.destination-title {
    font-weight: 600;
    color: #2c3e50;
}

.destination-remove {
    background: none;
    border: none;
    color: #e74c3c;
    font-size: 1.2em;
    cursor: pointer;
    padding: 0 5px;
}

.destination-config {
    margin-bottom: 15px;
}

.dest-config-fields {
    margin-top: 10px;
}

.dest-config-field {
    margin-bottom: 10px;
}

.dest-config-field label {
    display: block;
    margin-bottom: 5px;
    color: #2c3e50;
    font-weight: 500;
}

.dest-config-field input {
    width: 100%;
    padding: 8px;
    border: 1px solid #ddd;
    border-radius: 4px;
    font-size: 14px;
}

.destination-actions {
    display: flex;
    align-items: center;
    gap: 15px;
    margin-top: 15px;
}

.test-button {
    padding: 6px 12px;
    font-size: 14px;
}

.test-status {
    padding: 6px 12px;
    border-radius: 4px;
    font-size: 14px;
}

.test-status.idle {
    background: #f8f9fa;
    color: #6c757d;
}

.test-status.testing {
    background: #fff3cd;
    color: #856404;
}

.test-status.success {
    background: #d4edda;
    color: #155724;
}

.test-status.failed {
    background: #f8d7da;
    color: #721c24;
}

.destination-enable {
    display: flex;
    align-items: center;
    gap: 5px;
}

.destination-enable input[type="checkbox"] {
    margin: 0;
}

.destination-enable label {
    margin: 0;
    color: #2c3e50;
}

.simulation-status {
    margin-top: 5px;
    display: flex;
    gap: 10px;
    font-size: 0.9rem;
}

.simulation-badge {
    padding: 4px 12px;
    border-radius: 20px;
    font-weight: 500;
}

.simulation-on {
    background: #d5f5d7;
    color: #2ecc71;
}

.simulation-off {
    background: #fff3cd;
    color: #856404;
}

.queue-info, .processed-info {
    color: #7f8c8d;
}

@media (max-width: 768px) {
    .container {
        padding: 10px;
    }
    
    header {
        flex-direction: column;
        gap: 15px;
        text-align: center;
    }
    
    .metrics-grid {
        grid-template-columns: repeat(2, 1fr);
    }
    
    .section-header {
        flex-direction: column;
        gap: 15px;
        align-items: stretch;
    }
    
    table {
        font-size: 0.9rem;
    }
    
    th, td {
        padding: 10px;
    }
    
    .destination-config {
        grid-template-columns: 1fr;
    }
}`

// JSContent contains the dashboard JavaScript (WITH ALL FIXES APPLIED)
const JSContent = `class SyslogDashboard {
    constructor() {
        this.ws = null;
        this.reconnectInterval = 5000;
        this.isConnected = false;
        this.destinationCounter = 0;
        this.init();
    }

    init() {
        this.connectWebSocket();
        this.setupEventListeners();
        this.loadInitialData();
    }

    connectWebSocket() {
        const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
        const wsUrl = protocol + '//' + window.location.host + '/ws';
        
        try {
            this.ws = new WebSocket(wsUrl);
            
            this.ws.onopen = () => {
                console.log('WebSocket connected');
                this.isConnected = true;
                this.updateConnectionStatus(true);
            };
            
            this.ws.onmessage = (event) => {
                const data = JSON.parse(event.data);
                this.updateDashboard(data);
            };
            
            this.ws.onclose = () => {
                console.log('WebSocket disconnected');
                this.isConnected = false;
                this.updateConnectionStatus(false);
                this.scheduleReconnect();
            };
            
            this.ws.onerror = (error) => {
                console.error('WebSocket error:', error);
            };
        } catch (error) {
            console.error('Failed to create WebSocket:', error);
            this.scheduleReconnect();
        }
    }

    scheduleReconnect() {
        setTimeout(() => {
            if (!this.isConnected) {
                console.log('Attempting to reconnect...');
                this.connectWebSocket();
            }
        }, this.reconnectInterval);
    }

    updateConnectionStatus(connected) {
        const statusDot = document.getElementById('connectionStatus');
        const statusText = document.getElementById('statusText');
        
        if (connected) {
            statusDot.classList.add('connected');
            statusText.textContent = 'Connected';
        } else {
            statusDot.classList.remove('connected');
            statusText.textContent = 'Disconnected';
        }
    }

    setupEventListeners() {
        document.getElementById('addSourceForm').addEventListener('submit', (e) => {
            e.preventDefault();
            this.addSource();
        });

        // Close modal when clicking outside
        window.addEventListener('click', (e) => {
            const modal = document.getElementById('addSourceModal');
            if (e.target === modal) {
                this.hideAddSourceModal();
            }
        });
    }

    async loadInitialData() {
        try {
            const response = await fetch('/api/metrics');
            const data = await response.json();
            this.updateDashboard(data);
        } catch (error) {
            console.error('Failed to load initial data:', error);
        }
    }

    updateDashboard(data) {
        this.updateGlobalMetrics(data.global);
        this.updateSourcesTable(data.sources);
    }

    updateGlobalMetrics(global) {
        if (!global) {
            console.log('Global metrics is null');
            return;
        }
        
        try {
            const globalEPS = document.getElementById('globalEPS');
            const globalGBps = document.getElementById('globalGBps');
            const totalHourlyAvg = document.getElementById('totalHourlyAvg');
            const totalDailyAvg = document.getElementById('totalDailyAvg');
            const activeSources = document.getElementById('activeSources');
            const totalSources = document.getElementById('totalSources');
            
            if (globalEPS) globalEPS.textContent = (global.total_realtime_eps || 0).toFixed(2);
            if (globalGBps) globalGBps.textContent = (global.total_realtime_gbps || 0).toFixed(6);
            if (totalHourlyAvg) totalHourlyAvg.textContent = ((global.total_hourly_avg && global.total_hourly_avg.hourly_avg_gb) || 0).toFixed(5);
            if (totalDailyAvg) totalDailyAvg.textContent = ((global.total_daily_avg && global.total_daily_avg.daily_avg_gb) || 0).toFixed(5);
            if (activeSources) activeSources.textContent = global.active_sources || 0;
            if (totalSources) totalSources.textContent = global.total_sources || 0;
        } catch (e) {
            console.error('Error updating global metrics:', e);
        }
    }

    // FIX 3: Fixed null sort error with comprehensive null checks
    updateSourcesTable(sources) {
        const tbody = document.getElementById('sourcesTableBody');
        if (!tbody) {
            console.error('Sources table body not found');
            return;
        }
        
        tbody.innerHTML = '';

        // FIX 3: Add null/undefined check for sources array
        if (!sources || !Array.isArray(sources)) {
            console.error('Invalid or missing sources in updateSourcesTable');
            return;
        }

        // Sort sources alphabetically by name (A to Z) with null safety
        sources.sort((a, b) => {
            const nameA = (a && a.name) ? a.name.toLowerCase() : '';
            const nameB = (b && b.name) ? b.name.toLowerCase() : '';
            return nameA.localeCompare(nameB);
        });

        sources.forEach(source => {
            if (!source) return; // Skip null/undefined sources
            
            const row = document.createElement('tr');
            
            // Determine status based on activity and message reception
            let statusClass, statusText;
            if (source.is_active && source.is_receiving) {
                statusClass = 'status-active';
                statusText = 'Active';
            } else if (source.is_active && !source.is_receiving) {
                statusClass = 'status-idle';
                statusText = 'Idle: Waiting for Logs';
            } else {
                statusClass = 'status-inactive';
                statusText = 'Inactive';
            }
            
            row.innerHTML = 
                "<td>" +
                    "<div class=\"source-info\">" +
                        "<div class=\"source-name\">" + (source.name || "Unknown") + "</div>" +
                        "<div class=\"source-address\">" + (source.source_ip || "N/A") + ":" + (source.port || "N/A") + " (" + (source.protocol || "N/A") + ")</div>" +
                        "<span class=\"status-badge " + statusClass + "\">" + statusText + "</span>" +
                        "<div class=\"simulation-status\">" +
                            "<span class=\"simulation-badge " + (source.simulation_mode ? "simulation-on" : "simulation-off") + "\">" +
                                (source.simulation_mode ? "Simulation Mode: ON" : "Simulation Mode: OFF") +
                            "</span>" +
                            "<span class=\"queue-info\">Queue: " + (source.queue_length || 0) + " logs</span>" +
                            "<span class=\"processed-info\">Processed: " + (source.processed_count || 0) + " logs</span>" +
                        "</div>" +
                    "</div>" +
                "</td>" +
                "<td>" +
                    "<div class=\"metrics-column\">" +
                        "<div class=\"metric-row\">" +
                            "<span class=\"metric-label\">EPS:</span>" +
                            "<span class=\"metric-number\">" + (source.realtime_eps || 0).toFixed(2) + "</span>" +
                        "</div>" +
                        "<div class=\"metric-row\">" +
                            "<span class=\"metric-label\">GB/s:</span>" +
                            "<span class=\"metric-number\">" + (source.realtime_gbps || 0).toFixed(6) + "</span>" +
                        "</div>" +
                    "</div>" +
                "</td>" +
                "<td>" +
                    "<div class=\"metrics-column\">" +
                        "<div class=\"metric-row\">" +
                            "<span class=\"metric-label\">Logs:</span>" +
                            "<span class=\"metric-number\">" + (source.hourly_avg_logs || 0).toLocaleString() + "</span>" +
                        "</div>" +
                        "<div class=\"metric-row\">" +
                            "<span class=\"metric-label\">GB:</span>" +
                            "<span class=\"metric-number\">" + (source.hourly_avg_gb || 0).toFixed(4) + "</span>" +
                        "</div>" +
                    "</div>" +
                "</td>" +
                "<td>" +
                    "<div class=\"metrics-column\">" +
                        "<div class=\"metric-row\">" +
                            "<span class=\"metric-label\">Logs:</span>" +
                            "<span class=\"metric-number\">" + (source.daily_avg_logs || 0).toLocaleString() + "</span>" +
                        "</div>" +
                        "<div class=\"metric-row\">" +
                            "<span class=\"metric-label\">GB:</span>" +
                            "<span class=\"metric-number\">" + (source.daily_avg_gb || 0).toFixed(4) + "</span>" +
                        "</div>" +
                    "</div>" +
                "</td>" +
                "<td>" +
                    "<div class=\"button-group\">" +
                        "<button onclick=\"dashboard.editSource('" + (source.name || "") + "')\" class=\"btn btn-secondary btn-action\">" +
                            "Edit" +
                        "</button>" +
                        "<button onclick=\"dashboard.deleteSource('" + (source.name || "") + "')\" class=\"btn btn-danger btn-action\">" +
                            "Delete" +
                        "</button>" +
                    "</div>" +
                "</td>";
            
            tbody.appendChild(row);
        });
    }

    showAddSourceModal() {
        // Reset form for new source
        document.getElementById('addSourceForm').reset();
        document.getElementById('sourceProtocol').value = 'UDP';
        document.getElementById('simulationMode').checked = true;
        
        // Clear destinations
        document.getElementById('destinationsContainer').innerHTML = '';
        this.destinationCounter = 0;
        
        // Update modal title
        document.querySelector('#addSourceModal .modal-header h3').textContent = 'Add New Syslog Source';
        
        // Reset form action
        this.editingSourceName = null;
        
        document.getElementById('addSourceModal').style.display = 'block';

        const submitBtn = document.querySelector('#addSourceForm .btn.btn-primary');
        if (submitBtn) submitBtn.textContent = 'Add Source';
    }

    showEditSourceModal(source) {
        // Populate form with existing source data
        document.getElementById('sourceName').value = source.name;
        document.getElementById('sourceIP').value = source.ip;
        document.getElementById('sourcePort').value = source.port;
        document.getElementById('sourceProtocol').value = source.protocol;
        document.getElementById('simulationMode').checked = source.simulation_mode !== false;
        
        // Clear and populate destinations
        const container = document.getElementById('destinationsContainer');
        container.innerHTML = '';
        this.destinationCounter = 0;
        
        if (source.destinations && source.destinations.length > 0) {
            source.destinations.forEach(dest => {
                this.addDestination(dest);
            });
        }
        
        // Update modal title
        document.querySelector('#addSourceModal .modal-header h3').textContent = 'Edit Syslog Source';
        
        // Set editing mode
        this.editingSourceName = source.name;
        
        document.getElementById('addSourceModal').style.display = 'block';

        const submitBtnEdit = document.querySelector('#addSourceForm .btn.btn-primary');
        if (submitBtnEdit) submitBtnEdit.textContent = 'Apply Changes';
    }

    hideAddSourceModal() {
        document.getElementById('addSourceModal').style.display = 'none';
        document.getElementById('addSourceForm').reset();
        document.getElementById('destinationsContainer').innerHTML = '';
        this.destinationCounter = 0;
    }

    // FIX 1: Removed Destination Name field completely from addDestination
    addDestination(existingDest = null) {
        const container = document.getElementById('destinationsContainer');
        const destId = existingDest ? existingDest.id : 'dest_' + (++this.destinationCounter);
        
        const destDiv = document.createElement('div');
        destDiv.className = 'destination-item';
        destDiv.setAttribute('data-dest-id', destId);
        
        const destType = existingDest ? existingDest.type : 'storage';
        const destConfig = existingDest ? existingDest.config : {};
        const destEnabled = existingDest ? existingDest.enabled : false;
        const destTested = existingDest ? existingDest.tested : false;
        const testStatus = existingDest ? existingDest.test_status : 'idle';
        const testMessage = existingDest ? existingDest.test_message : '';
        
        destDiv.innerHTML = 
            "<div class=\"destination-header\">" +
                "<div class=\"destination-title\">Destination " + this.destinationCounter + "</div>" +
                "<button type=\"button\" class=\"destination-remove\" onclick=\"dashboard.removeDestination('" + destId + "')\">&times;</button>" +
            "</div>" +
            "<div class=\"destination-config\">" +
                "<div class=\"form-group\">" +
                    "<label>Destination Type:</label>" +
                    "<select class=\"dest-type\" onchange=\"dashboard.updateDestinationConfig('" + destId + "')\">" +
                        "<option value=\"storage\"" + (destType === "storage" ? " selected" : "") + ">Storage</option>" +
                        "<option value=\"hec\"" + (destType === "hec" ? " selected" : "") + ">HEC (HTTP Event Collector)</option>" +
                    "</select>" +
                "</div>" +
            "</div>" +
            "<div class=\"dest-config-fields\">" +
                this.getDestinationConfigHTML(destType, destConfig) +
            "</div>" +
            "<div class=\"destination-actions\">" +
                "<button type=\"button\" class=\"btn btn-secondary test-button\" onclick=\"dashboard.testDestination('" + destId + "')\">" +
                    "Test Connection" +
                "</button>" +
                "<div class=\"test-status " + testStatus + "\" id=\"test-status-" + destId + "\">" +
                    (testMessage || "Not tested") +
                "</div>" +
                "<div class=\"destination-enable\">" +
                    "<input type=\"checkbox\" class=\"dest-enabled\"" + (destEnabled ? " checked" : "") + (!destTested ? " disabled" : "") + ">" +
                    "<label>Enable</label>" +
                "</div>" +
            "</div>";
        
        container.appendChild(destDiv);
    }

    getDestinationConfigHTML(type, config) {
        if (type === "storage") {
            return "<div class=\"dest-config-field\">" +
                    "<label>Storage Path:</label>" +
                    "<input type=\"text\" class=\"dest-config-path\" value=\"" + (config.path || "") + "\" placeholder=\"/path/to/storage\">" +
                "</div>";
        } else if (type === "hec") {
            return "<div class=\"dest-config-field\">" +
                    "<label>HEC URL:</label>" +
                    "<input type=\"text\" class=\"dest-config-url\" value=\"" + (config.url || "") + "\" placeholder=\"https://splunk:8088/services/collector\">" +
                "</div>" +
                "<div class=\"dest-config-field\">" +
                    "<label>API Key:</label>" +
                    "<input type=\"text\" class=\"dest-config-apikey\" value=\"" + (config.api_key || "") + "\" placeholder=\"xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx\">" +
                "</div>";
        }
        return "";
    }

    updateDestinationConfig(destId) {
        const destDiv = document.querySelector("[data-dest-id=\"" + destId + "\"]");
        const typeSelect = destDiv.querySelector(".dest-type");
        const configFields = destDiv.querySelector(".dest-config-fields");
        
        configFields.innerHTML = this.getDestinationConfigHTML(typeSelect.value, {});
        
        // Reset test status
        const testStatus = destDiv.querySelector("#test-status-" + destId);
        testStatus.className = "test-status idle";
        testStatus.textContent = "Not tested";
        
        // Disable enable checkbox
        const enableCheckbox = destDiv.querySelector(".dest-enabled");
        enableCheckbox.checked = false;
        enableCheckbox.disabled = true;
    }

    removeDestination(destId) {
        const destDiv = document.querySelector("[data-dest-id=\"" + destId + "\"]");
        if (destDiv) {
            destDiv.remove();
        }
    }

    // FIX 2: Fixed testDestination to properly send source IP
    async testDestination(destId) {
        const destDiv = document.querySelector("[data-dest-id=\"" + destId + "\"]");
        const testStatus = destDiv.querySelector("#test-status-" + destId);
        const enableCheckbox = destDiv.querySelector(".dest-enabled");
        
        // FIX 2: Properly get both source name and IP
        const sourceName = document.getElementById("sourceName")?.value?.trim();
        const sourceIP = document.getElementById("sourceIP")?.value?.trim();
        
        if (!sourceName || !sourceIP) {
            alert("Please enter both source name and IP address before testing destinations");
            return;
        }
        
        // Collect destination data
        const destination = this.collectSingleDestination(destId);
        if (!destination) {
            alert("Invalid destination configuration");
            return;
        }
        
        // Update status to testing
        testStatus.className = "test-status testing";
        testStatus.textContent = "Testing...";
        enableCheckbox.checked = false;
        enableCheckbox.disabled = true;
        
        try {
            const response = await fetch("/api/destinations/test", {
                method: "POST",
                headers: {
                    "Content-Type": "application/json",
                },
                body: JSON.stringify({
                    source_name: sourceName,
                    source_ip: sourceIP,  // FIX 2: Now properly sending source IP
                    destination: destination
                })
            });
            
            const result = await response.json();
            
            if (result.success) {
                testStatus.className = "test-status success";
                testStatus.textContent = result.message;
                enableCheckbox.disabled = false;
            } else {
                testStatus.className = "test-status failed";
                testStatus.textContent = result.message;
                enableCheckbox.disabled = true;
            }
        } catch (error) {
            console.error("Error testing destination:", error);
            testStatus.className = "test-status failed";
            testStatus.textContent = "Test failed: " + error.message;
            enableCheckbox.disabled = true;
        }
    }

    collectSingleDestination(destId) {
        const destDiv = document.querySelector("[data-dest-id=\"" + destId + "\"]");
        if (!destDiv) {
            return null;
        }
        
        const type = destDiv.querySelector(".dest-type").value;
        const enabled = destDiv.querySelector(".dest-enabled").checked;
        
        let config = {};
        if (type === "storage") {
            const pathInput = destDiv.querySelector(".dest-config-path");
            if (!pathInput) {
                return null;
            }
            const path = pathInput.value;
            config = { path };
        } else if (type === "hec") {
            const urlInput = destDiv.querySelector(".dest-config-url");
            const apiKeyInput = destDiv.querySelector(".dest-config-apikey");
            if (!urlInput || !apiKeyInput) {
                return null;
            }
            const url = urlInput.value;
            const apiKey = apiKeyInput.value;
            config = { url, api_key: apiKey };
        }
        
        return {
            id: destId,
            type: type,
            name: type + "_destination", // Auto-generate name instead of user input
            config: config,
            enabled: enabled,
            tested: false,
            test_status: "idle",
            test_message: ""
        };
    }

    collectDestinations() {
        const destinations = [];
        const destDivs = document.querySelectorAll('.destination-item');
        
        destDivs.forEach(destDiv => {
            const destId = destDiv.getAttribute('data-dest-id');
            const destination = this.collectSingleDestination(destId);
            if (destination) {
                destinations.push(destination);
            }
        });
        
        return destinations;
    }

    async addSource() {
        const formData = {
            name: document.getElementById("sourceName").value,
            ip: document.getElementById("sourceIP").value,
            port: parseInt(document.getElementById("sourcePort").value),
            protocol: document.getElementById("sourceProtocol").value,
            destinations: this.collectDestinations(),
            simulation_mode: document.getElementById("simulationMode").checked
        };

        try {
            let response;
            if (this.editingSourceName) {
                // Update existing source
                response = await fetch("/api/sources/" + encodeURIComponent(this.editingSourceName), {
                    method: "PUT",
                    headers: {
                        "Content-Type": "application/json",
                    },
                    body: JSON.stringify(formData)
                });
            } else {
                // Create new source
                response = await fetch("/api/sources", {
                    method: "POST",
                    headers: {
                        "Content-Type": "application/json",
                    },
                    body: JSON.stringify(formData)
                });
            }

            if (response.ok) {
                this.hideAddSourceModal();
                // Automatically reload the dashboard data
                setTimeout(() => {
                    this.loadInitialData();
                }, 500);
            } else {
                const result = await response.json();
                alert("Error " + (this.editingSourceName ? "updating" : "adding") + " source: " + (result.error || "Unknown error"));
                // Don't hide modal on error
                return;
            }
        } catch (error) {
            console.error("Error " + (this.editingSourceName ? "updating" : "adding") + " source:", error);
            alert("Error " + (this.editingSourceName ? "updating" : "adding") + " source: " + error.message);
            // Don't hide modal on error
            return;
        }
    }

    async editSource(name) {
        try {
            const response = await fetch("/api/sources");
            const sources = await response.json();
            const source = sources.find(s => s.name === name);
            
            if (source) {
                this.showEditSourceModal(source);
            } else {
                alert("Source not found");
            }
        } catch (error) {
            console.error("Error loading source for edit:", error);
            alert("Error loading source for edit: " + error.message);
        }
    }

    async deleteSource(name) {
        if (!confirm("Are you sure you want to delete source \"" + name + "\"?")) {
            return;
        }

        try {
            const response = await fetch("/api/sources/" + encodeURIComponent(name), {
                method: "DELETE"
            });

            if (response.ok) {
                this.loadInitialData(); // Refresh data
            } else {
                const result = await response.json();
                alert("Error deleting source: " + (result.error || "Unknown error"));
            }
        } catch (error) {
            console.error("Error deleting source:", error);
            alert("Error deleting source: " + error.message);
        }
    }

    generateReport() {
        window.open('/api/report', '_blank');
    }
}

// Global functions for onclick handlers
function showAddSourceModal() {
    dashboard.showAddSourceModal();
}

function hideAddSourceModal() {
    dashboard.hideAddSourceModal();
}

function addDestination() {
    dashboard.addDestination();
}

function editSource(name) {
    dashboard.editSource(name);
}

function generateReport() {
    dashboard.generateReport();
}

// Initialize dashboard when page loads
let dashboard;
document.addEventListener('DOMContentLoaded', () => {
    dashboard = new SyslogDashboard();
});`
