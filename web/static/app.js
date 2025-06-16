// WebSocket connection
let ws = null;
let reconnectAttempts = 0;
const maxReconnectAttempts = 5;
const reconnectDelay = 3000;

// Initialize WebSocket connection
function initWebSocket() {
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const wsUrl = `${protocol}//${window.location.host}/ws`;
    
    ws = new WebSocket(wsUrl);
    
    ws.onopen = function() {
        console.log('WebSocket connected');
        document.getElementById('connectionStatus').classList.add('connected');
        document.getElementById('statusText').textContent = 'Connected';
        reconnectAttempts = 0;
    };
    
    ws.onclose = function() {
        console.log('WebSocket disconnected');
        document.getElementById('connectionStatus').classList.remove('connected');
        document.getElementById('statusText').textContent = 'Disconnected';
        
        if (reconnectAttempts < maxReconnectAttempts) {
            setTimeout(initWebSocket, reconnectDelay);
            reconnectAttempts++;
        }
    };
    
    ws.onerror = function(error) {
        console.error('WebSocket error:', error);
    };
    
    ws.onmessage = function(event) {
        const data = JSON.parse(event.data);
        updateDashboard(data);
    };
}

// Format numbers with commas and optional decimal places
function formatNumber(num, decimals = 0) {
    if (num === undefined || num === null) return '0';
    return num.toLocaleString('en-US', {
        minimumFractionDigits: decimals,
        maximumFractionDigits: decimals
    });
}

// Update dashboard with new data
function updateDashboard(data) {
    if (data.sources) {
        updateSourcesTable(data.sources);
    }
    if (data.global) {
        updateGlobalMetrics(data.global);
    }
}

// Update sources table
function updateSourcesTable(sources) {
    const tbody = document.getElementById('sourcesTableBody');
    tbody.innerHTML = '';
    
    sources.forEach(source => {
        const row = updateSourceRow(source);
        tbody.appendChild(row);
    });
}

function updateSourceRow(source) {
    const row = document.createElement('tr');
    row.innerHTML = `
        <td>
            <div class="source-info">
                <span class="source-name">${source.name}</span>
                <span class="source-ip">${source.source_ip}:${source.port}</span>
                <span class="source-protocol">${source.protocol}</span>
            </div>
        </td>
        <td>
            <div class="metrics">
                <div class="metric">
                    <span class="label">EPS:</span>
                    <span class="value">${formatNumber(source.realtime_eps)}</span>
                </div>
                <div class="metric">
                    <span class="label">GB/s:</span>
                    <span class="value">${formatNumber(source.realtime_gbps, 6)}</span>
                </div>
            </div>
        </td>
        <td>
            <div class="metrics">
                <div class="metric">
                    <span class="label">Logs:</span>
                    <span class="value">${formatNumber(source.hourly_avg_logs)}</span>
                </div>
                <div class="metric">
                    <span class="label">GB:</span>
                    <span class="value">${formatNumber(source.hourly_avg_gb, 2)}</span>
                </div>
            </div>
        </td>
        <td>
            <div class="metrics">
                <div class="metric">
                    <span class="label">Logs:</span>
                    <span class="value">${formatNumber(source.daily_avg_logs)}</span>
                </div>
                <div class="metric">
                    <span class="label">GB:</span>
                    <span class="value">${formatNumber(source.daily_avg_gb, 2)}</span>
                </div>
            </div>
        </td>
        <td>
            <div class="destinations">
                ${source.destinations.map(dest => `
                    <div class="destination">
                        <span class="dest-name">${dest.name}</span>
                        <span class="dest-type">${dest.type}</span>
                        <div class="dest-metrics">
                            <span class="label">Processed:</span>
                            <span class="value">${formatNumber(dest.processed_logs)}</span>
                            <span class="label">Queue:</span>
                            <span class="value">${formatNumber(dest.queue_size)}</span>
                        </div>
                    </div>
                `).join('')}
            </div>
        </td>
        <td>
            <div class="simulation-mode ${source.simulation_mode ? 'active' : 'inactive'}">
                ${source.simulation_mode ? 'ON' : 'OFF'}
            </div>
        </td>
        <td>
            <div class="actions">
                <button onclick="testSource('${source.name}')" class="btn btn-secondary btn-small">Test</button>
                <button onclick="removeSource('${source.name}')" class="btn btn-danger btn-small">Remove</button>
            </div>
        </td>
    `;
    return row;
}

function updateGlobalMetrics(global) {
    document.getElementById('globalEPS').textContent = formatNumber(global.total_realtime_eps);
    document.getElementById('globalGBps').textContent = formatNumber(global.total_realtime_gbps, 6);
    document.getElementById('totalHourlyAvg').textContent = formatNumber(global.total_hourly_avg_gb, 2);
    document.getElementById('totalDailyAvg').textContent = formatNumber(global.total_daily_avg_gb, 2);
    document.getElementById('activeSources').textContent = global.active_sources;
    document.getElementById('totalSources').textContent = global.total_sources;
}

// Modal functions
function showAddSourceModal() {
    document.getElementById('addSourceModal').style.display = 'block';
}

function hideAddSourceModal() {
    document.getElementById('addSourceModal').style.display = 'none';
}

// Destination management
function addDestination() {
    const container = document.getElementById('destinationsContainer');
    const destDiv = document.createElement('div');
    destDiv.className = 'destination-entry';
    destDiv.innerHTML = `
        <div class="form-group">
            <label>Destination Type:</label>
            <select class="dest-type" required>
                <option value="storage">Storage</option>
                <option value="hec">HEC</option>
            </select>
        </div>
        <div class="form-group">
            <label>Name:</label>
            <input type="text" class="dest-name" required>
        </div>
        <div class="form-group storage-config">
            <label>Path:</label>
            <input type="text" class="dest-path" placeholder="/path/to/storage">
        </div>
        <div class="form-group hec-config" style="display: none;">
            <label>URL:</label>
            <input type="text" class="dest-url" placeholder="https://hec-endpoint:8088/services/collector">
            <label>API Key:</label>
            <input type="text" class="dest-api-key" placeholder="your-api-key">
        </div>
        <button type="button" onclick="removeDestination(this)" class="btn btn-danger btn-small">Remove</button>
    `;
    container.appendChild(destDiv);
    
    // Add event listener for destination type change
    const typeSelect = destDiv.querySelector('.dest-type');
    typeSelect.addEventListener('change', function() {
        const storageConfig = destDiv.querySelector('.storage-config');
        const hecConfig = destDiv.querySelector('.hec-config');
        if (this.value === 'storage') {
            storageConfig.style.display = 'block';
            hecConfig.style.display = 'none';
        } else {
            storageConfig.style.display = 'none';
            hecConfig.style.display = 'block';
        }
    });
}

function removeDestination(button) {
    button.parentElement.remove();
}

function getDestinations() {
    const destinations = [];
    const entries = document.querySelectorAll('.destination-entry');
    
    entries.forEach(entry => {
        const type = entry.querySelector('.dest-type').value;
        const name = entry.querySelector('.dest-name').value;
        let config = {};
        
        if (type === 'storage') {
            config = {
                path: entry.querySelector('.dest-path').value
            };
        } else {
            config = {
                url: entry.querySelector('.dest-url').value,
                api_key: entry.querySelector('.dest-api-key').value
            };
        }
        
        destinations.push({
            type: type,
            name: name,
            config: config,
            enabled: true
        });
    });
    
    return destinations;
}

// Source management
async function testSource(sourceName) {
    try {
        const response = await fetch(`/api/sources/${encodeURIComponent(sourceName)}/test`, {
            method: 'POST'
        });
        
        if (!response.ok) {
            const error = await response.json();
            throw new Error(error.message || 'Failed to test source');
        }
        
        const result = await response.json();
        alert(result.message);
    } catch (error) {
        alert(error.message);
    }
}

async function removeSource(sourceName) {
    if (!confirm(`Are you sure you want to remove source "${sourceName}"?`)) {
        return;
    }
    
    try {
        const response = await fetch(`/api/sources/${encodeURIComponent(sourceName)}`, {
            method: 'DELETE'
        });
        
        if (!response.ok) {
            const error = await response.json();
            throw new Error(error.message || 'Failed to remove source');
        }
    } catch (error) {
        alert(error.message);
    }
}

// Form submission
document.getElementById('addSourceForm').addEventListener('submit', async function(e) {
    e.preventDefault();
    
    const sourceData = {
        name: document.getElementById('sourceName').value,
        ip: document.getElementById('sourceIP').value,
        port: parseInt(document.getElementById('sourcePort').value),
        protocol: document.getElementById('sourceProtocol').value,
        destinations: getDestinations(),
        simulation_mode: document.getElementById('simulationMode').checked
    };
    
    try {
        const response = await fetch('/api/sources', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(sourceData)
        });
        
        const result = await response.json();
        
        if (!response.ok) {
            throw new Error(result.message || 'Failed to add source');
        }
        
        hideAddSourceModal();
        this.reset();
        document.getElementById('destinationsContainer').innerHTML = '';
        
        // Clear any previous error messages
        const errorElements = document.querySelectorAll('.error-message');
        errorElements.forEach(el => el.remove());
    } catch (error) {
        // Show error message in the form
        const errorDiv = document.createElement('div');
        errorDiv.className = 'error-message';
        errorDiv.style.color = '#e74c3c';
        errorDiv.style.marginTop = '10px';
        errorDiv.style.padding = '10px';
        errorDiv.style.backgroundColor = '#fde8e8';
        errorDiv.style.borderRadius = '6px';
        errorDiv.textContent = error.message;
        
        const form = document.getElementById('addSourceForm');
        form.insertBefore(errorDiv, form.firstChild);
        
        // Enable the Add Source button
        const submitButton = form.querySelector('button[type="submit"]');
        submitButton.disabled = false;
    }
});

// Add event listeners to form inputs to clear error messages
const formInputs = document.querySelectorAll('#addSourceForm input, #addSourceForm select');
formInputs.forEach(input => {
    input.addEventListener('input', function() {
        const errorMessage = document.querySelector('.error-message');
        if (errorMessage) {
            errorMessage.remove();
        }
    });
});

// Initialize WebSocket connection when page loads
document.addEventListener('DOMContentLoaded', initWebSocket); 