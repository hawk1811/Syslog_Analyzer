package web

// EmbeddedContent contains all embedded web assets
// This replaces separate static files with embedded content for single binary deployment

// HTMLContent contains the complete dashboard HTML
const HTMLContent = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Professional Syslog Analyzer Dashboard</title>
    <style>
        ` + CSSContent + `
    </style>
</head>
<body>
    <div class="container">
        <header>
            <h1>🚀 Professional Syslog Analyzer</h1>
            <div class="status-indicator">
                <div class="status-dot" id="connectionStatus"></div>
                <span id="statusText">Connecting...</span>
            </div>
        </header>

        <div class="dashboard">
            <div class="global-metrics">
                <h2>📊 Global Summary</h2>
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
                        <h3>Total Logs Ingested</h3>
                        <div class="metric-value" id="totalLogsIngested">0</div>
                    </div>
                    <div class="metric-card">
                        <h3>Total Hourly Avg Logs</h3>
                        <div class="metric-value" id="totalHourlyAvgLogs">0</div>
                    </div>
                    <div class="metric-card">
                        <h3>Total Daily Avg Logs</h3>
                        <div class="metric-value" id="totalDailyAvgLogs">0</div>
                    </div>
                    <div class="metric-card">
                        <h3>Active / Total Sources</h3>
                        <div class="metric-value"><span id="activeSources">0</span> / <span id="totalSources">0</span></div>
                    </div>
                </div>
            </div>

            <div class="sources-section">
                <div class="section-header">
                    <h2>📡 Syslog Sources</h2>
                    <div class="actions">
                        <button onclick="generateReport()" class="btn btn-secondary">📊 Export Report</button>
                        <button onclick="showAddSourceModal()" class="btn btn-primary">➕ Add Source</button>
                    </div>
                </div>
                
                <div class="sources-table">
                    <table id="sourcesTable">
                        <thead>
                            <tr>
                                <th>Source Details</th>
                                <th>Real-time Metrics</th>
                                <th>Hourly Averages</th>
                                <th>Daily Averages</th>
                                <th>Queue & Processing</th>
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
                    <input type="text" id="sourceIP" required placeholder="192.168.1.100 or 0.0.0.0 for any">
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
                    <button type="button" onclick="addDestination()" class="btn btn-secondary btn-small">➕ Add Destination</button>
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
                    <small class="help-text">When ON: Processes logs for metrics only. When OFF: Full processing with destinations.</small>
                </div>
                
                <div class="form-actions">
                    <button type="button" onclick="hideAddSourceModal()" class="btn btn-secondary">Cancel</button>
                    <button type="submit" class="btn btn-primary">Add Source</button>
                </div>
            </form>
        </div>
    </div>

    <script>
        ` + JSContent + `
    </script>
</body>
</html>`

// CSSContent contains the enhanced dashboard CSS
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
    max-width: 1600px;
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
    grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
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

.simulation-mode {
    padding: 4px 8px;
    border-radius: 12px;
    font-size: 0.8rem;
    font-weight: 500;
    display: inline-block;
}

.simulation-mode.on {
    background: #2ecc71;
    color: white;
}

.simulation-mode.off {
    background: #3498db;
    color: white;
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
    max-width: 700px;
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

.help-text {
    display: block;
    margin-top: 5px;
    font-size: 0.85rem;
    color: #7f8c8d;
    font-style: italic;
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

/* Destination Styles */
.destination-item {
    border: 2px solid #ecf0f1;
    border-radius: 12px;
    padding: 20px;
    margin-bottom: 15px;
    background: #f8f9fa;
    position: relative;
}

.destination-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 15px;
}

.destination-title {
    font-weight: 600;
    color: #2c3e50;
    font-size: 1.1rem;
}

.destination-remove {
    background: #e74c3c;
    color: white;
    border: none;
    border-radius: 50%;
    width: 30px;
    height: 30px;
    cursor: pointer;
    font-size: 18px;
    display: flex;
    align-items: center;
    justify-content: center;
}

.destination-remove:hover {
    background: #c0392b;
}

.destination-config {
    display: grid;
    grid-template-columns: 1fr;
    gap: 15px;
    margin-bottom: 15px;
}

.destination-config .form-group {
    margin-bottom: 0;
}

.destination-actions {
    display: flex;
    gap: 15px;
    align-items: center;
    margin-top: 15px;
    padding-top: 15px;
    border-top: 1px solid #dee2e6;
}

.test-button {
    padding: 8px 16px;
    font-size: 0.9rem;
}

.test-status {
    font-size: 0.9rem;
    font-weight: 500;
    padding: 4px 8px;
    border-radius: 4px;
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
    gap: 8px;
}

.destination-enable input[type="checkbox"] {
    width: auto;
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
}`

// JSContent contains the enhanced dashboard JavaScript - Fixed for Go embedding
const JSContent = `class SyslogDashboard {
    constructor() {
        this.ws = null;
        this.reconnectInterval = 5000;
        this.isConnected = false;
        this.destinationCounter = 0;
        this.editingSourceName = null;
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
        if (!global) return;
        
        try {
            document.getElementById('globalEPS').textContent = (global.total_realtime_eps || 0).toFixed(2);
            document.getElementById('globalGBps').textContent = (global.total_realtime_gbps || 0).toFixed(6);
            document.getElementById('totalLogsIngested').textContent = (global.total_logs_ingested || 0).toLocaleString();
            document.getElementById('totalHourlyAvgLogs').textContent = (global.total_hourly_avg_logs || 0).toLocaleString();
            document.getElementById('totalDailyAvgLogs').textContent = (global.total_daily_avg_logs || 0).toLocaleString();
            document.getElementById('activeSources').textContent = global.active_sources || 0;
            document.getElementById('totalSources').textContent = global.total_sources || 0;
        } catch (e) {
            console.error('Error updating global metrics:', e);
        }
    }

    updateSourcesTable(sources) {
        const tbody = document.getElementById('sourcesTableBody');
        if (!tbody) return;
        
        tbody.innerHTML = '';

        if (!sources || !Array.isArray(sources)) return;

        sources.sort((a, b) => {
            const nameA = (a && a.name) ? a.name.toLowerCase() : '';
            const nameB = (b && b.name) ? b.name.toLowerCase() : '';
            return nameA.localeCompare(nameB);
        });

        sources.forEach(source => {
            if (!source) return;
            
            const row = document.createElement('tr');
            
            let statusClass, statusText;
            if (source.is_active && source.is_receiving) {
                statusClass = 'status-active';
                statusText = 'Active & Receiving';
            } else if (source.is_active && !source.is_receiving) {
                statusClass = 'status-idle';
                statusText = 'Idle: Waiting for Logs';
            } else {
                statusClass = 'status-inactive';
                statusText = 'Inactive';
            }
            
            const simulationClass = source.simulation_mode ? 'on' : 'off';
            const simulationText = source.simulation_mode ? 'ON' : 'OFF';
            
            row.innerHTML = '<td><div class="source-info"><div class="source-name">' + (source.name || 'Unknown') + '</div><div class="source-address">' + (source.source_ip || 'N/A') + ':' + (source.port || 'N/A') + ' (' + (source.protocol || 'N/A') + ')</div><span class="status-badge ' + statusClass + '">' + statusText + '</span><span class="simulation-mode ' + simulationClass + '">Simulation: ' + simulationText + '</span></div></td><td><div class="metrics-column"><div class="metric-row"><span class="metric-label">EPS:</span><span class="metric-number">' + (source.realtime_eps || 0).toFixed(2) + '</span></div><div class="metric-row"><span class="metric-label">GB/s:</span><span class="metric-number">' + (source.realtime_gbps || 0).toFixed(6) + '</span></div><div class="metric-row"><span class="metric-label">Total:</span><span class="metric-number">' + (source.total_logs_ingested || 0).toLocaleString() + '</span></div></div></td><td><div class="metrics-column"><div class="metric-row"><span class="metric-label">Logs:</span><span class="metric-number">' + (source.hourly_avg_logs || 0).toLocaleString() + '</span></div><div class="metric-row"><span class="metric-label">GB:</span><span class="metric-number">' + (source.hourly_avg_gb || 0).toFixed(4) + '</span></div></div></td><td><div class="metrics-column"><div class="metric-row"><span class="metric-label">Logs:</span><span class="metric-number">' + (source.daily_avg_logs || 0).toLocaleString() + '</span></div><div class="metric-row"><span class="metric-label">GB:</span><span class="metric-number">' + (source.daily_avg_gb || 0).toFixed(4) + '</span></div></div></td><td><div class="metrics-column"><div class="metric-row"><span class="metric-label">Queue:</span><span class="metric-number">' + (source.queue_depth || 0).toLocaleString() + '</span></div><div class="metric-row"><span class="metric-label">Processed:</span><span class="metric-number">' + (source.processed_count || 0).toLocaleString() + '</span></div><div class="metric-row"><span class="metric-label">Sent:</span><span class="metric-number">' + (source.sent_count || 0).toLocaleString() + '</span></div></div></td><td><div class="button-group"><button onclick="dashboard.editSource(\'' + (source.name || '') + '\')" class="btn btn-secondary btn-action">Edit</button><button onclick="dashboard.deleteSource(\'' + (source.name || '') + '\')" class="btn btn-danger btn-action">Delete</button></div></td>';
            
            tbody.appendChild(row);
        });
    }

    showAddSourceModal() {
        document.getElementById('addSourceForm').reset();
        document.getElementById('sourceProtocol').value = 'UDP';
        document.getElementById('simulationMode').checked = true;
        
        document.getElementById('destinationsContainer').innerHTML = '';
        this.destinationCounter = 0;
        
        document.querySelector('#addSourceModal .modal-header h3').textContent = 'Add New Syslog Source';
        this.editingSourceName = null;
        
        document.getElementById('addSourceModal').style.display = 'block';
    }

    hideAddSourceModal() {
        document.getElementById('addSourceModal').style.display = 'none';
        document.getElementById('addSourceForm').reset();
        document.getElementById('destinationsContainer').innerHTML = '';
        this.destinationCounter = 0;
    }

    addDestination() {
        const container = document.getElementById('destinationsContainer');
        const destId = 'dest_' + (++this.destinationCounter);
        
        const destDiv = document.createElement('div');
        destDiv.className = 'destination-item';
        destDiv.setAttribute('data-dest-id', destId);
        
        destDiv.innerHTML = '<div class="destination-header"><div class="destination-title">Destination ' + this.destinationCounter + '</div><button type="button" class="destination-remove" onclick="dashboard.removeDestination(\'' + destId + '\')">&times;</button></div><div class="destination-config"><div class="form-group"><label>Destination Type:</label><select class="dest-type" onchange="dashboard.updateDestinationConfig(\'' + destId + '\')"><option value="storage" selected>Storage</option><option value="hec">HEC (HTTP Event Collector)</option></select></div></div><div class="dest-config-fields"><div class="form-group"><label>Storage Path:</label><input type="text" class="dest-config-path" placeholder="C:\\\\logs\\\\test or //share/logs/test"></div></div><div class="destination-actions"><button type="button" class="btn btn-secondary test-button" onclick="dashboard.testDestination(\'' + destId + '\')">Test Connection</button><div class="test-status idle" id="test-status-' + destId + '">Not tested</div><div class="destination-enable"><input type="checkbox" class="dest-enabled" disabled><label>Enable</label></div></div>';
        
        container.appendChild(destDiv);
    }

    updateDestinationConfig(destId) {
        const destDiv = document.querySelector('[data-dest-id="' + destId + '"]');
        const typeSelect = destDiv.querySelector('.dest-type');
        const configFields = destDiv.querySelector('.dest-config-fields');
        
        if (typeSelect.value === 'storage') {
            configFields.innerHTML = '<div class="form-group"><label>Storage Path:</label><input type="text" class="dest-config-path" placeholder="C:\\\\logs\\\\test or //share/logs/test"></div>';
        } else if (typeSelect.value === 'hec') {
            configFields.innerHTML = '<div class="form-group"><label>HEC URL:</label><input type="text" class="dest-config-url" placeholder="https://splunk.example.com:8088/services/collector"></div><div class="form-group"><label>API Key:</label><input type="text" class="dest-config-apikey" placeholder="xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"></div>';
        }
        
        const testStatus = destDiv.querySelector('#test-status-' + destId);
        testStatus.className = 'test-status idle';
        testStatus.textContent = 'Not tested';
        
        const enableCheckbox = destDiv.querySelector('.dest-enabled');
        enableCheckbox.checked = false;
        enableCheckbox.disabled = true;
    }

    removeDestination(destId) {
        const destDiv = document.querySelector('[data-dest-id="' + destId + '"]');
        if (destDiv) {
            destDiv.remove();
        }
    }

    async testDestination(destId) {
        console.log('Testing destination:', destId);
    }

    async addSource() {
        console.log('Adding source...');
    }

    async editSource(name) {
        console.log('Editing source:', name);
    }

    async deleteSource(name) {
        if (!confirm('Are you sure you want to delete source "' + name + '"?')) {
            return;
        }
        console.log('Deleting source:', name);
    }

    generateReport() {
        window.open('/api/report', '_blank');
    }
}

function showAddSourceModal() { dashboard.showAddSourceModal(); }
function hideAddSourceModal() { dashboard.hideAddSourceModal(); }
function addDestination() { dashboard.addDestination(); }
function editSource(name) { dashboard.editSource(name); }
function generateReport() { dashboard.generateReport(); }

let dashboard;
document.addEventListener('DOMContentLoaded', () => {
    dashboard = new SyslogDashboard();
});`