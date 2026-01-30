from flask import Flask, jsonify, render_template_string, request
import time
import psutil
import os
from datetime import datetime
from threading import Lock

app = Flask(__name__)

# Metrics storage
class MetricsCollector:
    """Collects and stores application metrics."""
    
    def __init__(self):
        self.lock = Lock()
        self.request_count = 0
        self.total_response_time = 0.0
        self.endpoint_stats = {}  # Changed from defaultdict to regular dict
        self.start_time = datetime.now()
        self.error_count = 0
        self.cpu_percent = 0.0  # Cache CPU percentage
        self._last_cpu_check = time.time()
    
    def record_request(self, endpoint, response_time, status_code):
        """Record a request's metrics."""
        with self.lock:
            self.request_count += 1
            self.total_response_time += response_time
            
            # Initialize endpoint stats if not exists (thread-safe)
            if endpoint not in self.endpoint_stats:
                self.endpoint_stats[endpoint] = {
                    'count': 0,
                    'total_time': 0.0,
                    'errors': 0,
                    'status_codes': {}
                }
            
            stats = self.endpoint_stats[endpoint]
            stats['count'] += 1
            stats['total_time'] += response_time
            
            # Track status codes
            if status_code not in stats['status_codes']:
                stats['status_codes'][status_code] = 0
            stats['status_codes'][status_code] += 1
            
            if status_code >= 400:
                stats['errors'] += 1
                self.error_count += 1
    
    def get_metrics(self):
        """Get current metrics snapshot."""
        with self.lock:
            uptime = (datetime.now() - self.start_time).total_seconds()
            avg_response_time = (self.total_response_time / self.request_count 
                               if self.request_count > 0 else 0)
            
            # Make a copy of endpoint stats to avoid holding lock too long
            endpoint_stats_copy = {}
            for endpoint, stats in self.endpoint_stats.items():
                endpoint_stats_copy[endpoint] = {
                    'count': stats['count'],
                    'total_time': stats['total_time'],
                    'errors': stats['errors'],
                    'status_codes': dict(stats['status_codes'])
                }
            
            request_count = self.request_count
            error_count = self.error_count
        
        # Release lock before doing expensive operations
        # Update CPU percentage periodically (every 5 seconds) to avoid blocking
        current_time = time.time()
        if current_time - self._last_cpu_check > 5:
            self.cpu_percent = psutil.cpu_percent(interval=0)
            self._last_cpu_check = current_time
        
        # System metrics
        process = psutil.Process(os.getpid())
        memory_info = process.memory_info()
        
        endpoint_data = []
        for endpoint, stats in endpoint_stats_copy.items():
            avg_time = stats['total_time'] / stats['count'] if stats['count'] > 0 else 0
            endpoint_data.append({
                'endpoint': endpoint,
                'count': stats['count'],
                'avg_response_time': round(avg_time * 1000, 2),  # Convert to ms
                'errors': stats['errors'],
                'error_rate': round((stats['errors'] / stats['count'] * 100) 
                                  if stats['count'] > 0 else 0, 2),
                'status_codes': stats['status_codes']
            })
        
        return {
            'overview': {
                'total_requests': request_count,
                'total_errors': error_count,
                'error_rate': round((error_count / request_count * 100) 
                                  if request_count > 0 else 0, 2),
                'avg_response_time_ms': round(avg_response_time * 1000, 2),
                'uptime_seconds': round(uptime, 2),
                'uptime_formatted': self._format_uptime(uptime)
            },
            'system': {
                'cpu_percent': self.cpu_percent,
                'memory_mb': round(memory_info.rss / 1024 / 1024, 2),
                'memory_percent': process.memory_percent(),
                'threads': process.num_threads()
            },
            'endpoints': sorted(endpoint_data, key=lambda x: x['count'], reverse=True)
        }
    
    def _format_uptime(self, seconds):
        """Format uptime in a human-readable way."""
        days = int(seconds // 86400)
        hours = int((seconds % 86400) // 3600)
        minutes = int((seconds % 3600) // 60)
        secs = int(seconds % 60)
        
        parts = []
        if days > 0:
            parts.append(f"{days}d")
        if hours > 0:
            parts.append(f"{hours}h")
        if minutes > 0:
            parts.append(f"{minutes}m")
        parts.append(f"{secs}s")
        
        return " ".join(parts)


# Initialize metrics collector
metrics = MetricsCollector()


# Middleware for metrics collection
@app.before_request
def before_request():
    """Record request start time."""
    request.start_time = time.time()


@app.after_request
def after_request(response):
    """Record request completion and metrics."""
    # Skip tracking for metrics endpoints to avoid recursive tracking
    if hasattr(request, 'start_time') and request.endpoint not in ['metrics_dashboard', 'metrics_json']:
        response_time = time.time() - request.start_time
        endpoint = request.endpoint or 'unknown'
        metrics.record_request(endpoint, response_time, response.status_code)
    return response


# Application routes
@app.route("/")
def home():
    """Main application endpoint."""
    return "Hello, this is a test container!"


@app.route("/health")
def health():
    """Health check endpoint."""
    return jsonify({
        'status': 'healthy',
        'timestamp': datetime.now().isoformat()
    })


@app.route("/metrics/json")
def metrics_json():
    """
    Return metrics in JSON format.
    
    Note: In production, this endpoint should be protected with authentication
    or restricted by IP whitelist to prevent exposing system information.
    """
    return jsonify(metrics.get_metrics())


@app.route("/metrics")
def metrics_dashboard():
    """
    Display metrics dashboard.
    
    Note: In production, this endpoint should be protected with authentication
    or restricted by IP whitelist to prevent exposing system information.
    """
    metrics_data = metrics.get_metrics()
    
    # HTML template for the dashboard
    template = """
    <!DOCTYPE html>
    <html>
    <head>
        <title>Metrics Dashboard</title>
        <meta charset="utf-8">
        <meta name="viewport" content="width=device-width, initial-scale=1">
        <style>
            * { margin: 0; padding: 0; box-sizing: border-box; }
            body {
                font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, sans-serif;
                background: #f5f7fa;
                padding: 20px;
                color: #2c3e50;
            }
            .container { max-width: 1200px; margin: 0 auto; }
            h1 {
                color: #2c3e50;
                margin-bottom: 30px;
                font-size: 2em;
            }
            .metrics-grid {
                display: grid;
                grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
                gap: 20px;
                margin-bottom: 30px;
            }
            .metric-card {
                background: white;
                padding: 20px;
                border-radius: 8px;
                box-shadow: 0 2px 4px rgba(0,0,0,0.1);
            }
            .metric-card h3 {
                color: #7f8c8d;
                font-size: 0.9em;
                text-transform: uppercase;
                margin-bottom: 10px;
                font-weight: 600;
            }
            .metric-value {
                font-size: 2em;
                font-weight: bold;
                color: #3498db;
            }
            .metric-unit {
                font-size: 0.6em;
                color: #7f8c8d;
                margin-left: 5px;
            }
            .status-good { color: #27ae60; }
            .status-warning { color: #f39c12; }
            .status-bad { color: #e74c3c; }
            .endpoint-table {
                background: white;
                border-radius: 8px;
                box-shadow: 0 2px 4px rgba(0,0,0,0.1);
                overflow: hidden;
            }
            .endpoint-table h2 {
                padding: 20px;
                background: #ecf0f1;
                margin: 0;
            }
            table {
                width: 100%;
                border-collapse: collapse;
            }
            th, td {
                padding: 12px 20px;
                text-align: left;
                border-bottom: 1px solid #ecf0f1;
            }
            th {
                background: #f8f9fa;
                font-weight: 600;
                color: #2c3e50;
            }
            tr:hover { background: #f8f9fa; }
            .badge {
                display: inline-block;
                padding: 4px 8px;
                border-radius: 4px;
                font-size: 0.85em;
                font-weight: 600;
            }
            .badge-success { background: #d4edda; color: #155724; }
            .badge-warning { background: #fff3cd; color: #856404; }
            .badge-danger { background: #f8d7da; color: #721c24; }
            .refresh-info {
                text-align: center;
                margin-top: 20px;
                color: #7f8c8d;
                font-size: 0.9em;
            }
        </style>
        <script>
            // Auto-refresh every 5 seconds
            setTimeout(function(){ location.reload(); }, 5000);
        </script>
    </head>
    <body>
        <div class="container">
            <h1>📊 Application Metrics Dashboard</h1>
            
            <div class="metrics-grid">
                <div class="metric-card">
                    <h3>Total Requests</h3>
                    <div class="metric-value">{{ overview.total_requests }}</div>
                </div>
                <div class="metric-card">
                    <h3>Error Rate</h3>
                    <div class="metric-value {% if overview.error_rate > 10 %}status-bad{% elif overview.error_rate > 5 %}status-warning{% else %}status-good{% endif %}">
                        {{ overview.error_rate }}<span class="metric-unit">%</span>
                    </div>
                </div>
                <div class="metric-card">
                    <h3>Avg Response Time</h3>
                    <div class="metric-value {% if overview.avg_response_time_ms > 1000 %}status-bad{% elif overview.avg_response_time_ms > 500 %}status-warning{% else %}status-good{% endif %}">
                        {{ overview.avg_response_time_ms }}<span class="metric-unit">ms</span>
                    </div>
                </div>
                <div class="metric-card">
                    <h3>Uptime</h3>
                    <div class="metric-value" style="font-size: 1.5em;">{{ overview.uptime_formatted }}</div>
                </div>
                <div class="metric-card">
                    <h3>CPU Usage</h3>
                    <div class="metric-value {% if system.cpu_percent > 80 %}status-bad{% elif system.cpu_percent > 60 %}status-warning{% else %}status-good{% endif %}">
                        {{ system.cpu_percent }}<span class="metric-unit">%</span>
                    </div>
                </div>
                <div class="metric-card">
                    <h3>Memory Usage</h3>
                    <div class="metric-value {% if system.memory_percent > 80 %}status-bad{% elif system.memory_percent > 60 %}status-warning{% else %}status-good{% endif %}">
                        {{ system.memory_mb }}<span class="metric-unit">MB</span>
                    </div>
                </div>
            </div>
            
            <div class="endpoint-table">
                <h2>Endpoint Statistics</h2>
                <table>
                    <thead>
                        <tr>
                            <th>Endpoint</th>
                            <th>Requests</th>
                            <th>Avg Time (ms)</th>
                            <th>Errors</th>
                            <th>Error Rate</th>
                            <th>Status Codes</th>
                        </tr>
                    </thead>
                    <tbody>
                        {% for endpoint in endpoints %}
                        <tr>
                            <td><strong>{{ endpoint.endpoint }}</strong></td>
                            <td>{{ endpoint.count }}</td>
                            <td>{{ endpoint.avg_response_time }}</td>
                            <td>{{ endpoint.errors }}</td>
                            <td>
                                <span class="badge {% if endpoint.error_rate > 10 %}badge-danger{% elif endpoint.error_rate > 5 %}badge-warning{% else %}badge-success{% endif %}">
                                    {{ endpoint.error_rate }}%
                                </span>
                            </td>
                            <td>
                                {% for code, count in endpoint.status_codes.items() %}
                                    <span class="badge {% if code >= 500 %}badge-danger{% elif code >= 400 %}badge-warning{% else %}badge-success{% endif %}">
                                        {{ code }}: {{ count }}
                                    </span>
                                {% endfor %}
                            </td>
                        </tr>
                        {% endfor %}
                    </tbody>
                </table>
            </div>
            
            <div class="refresh-info">
                🔄 Auto-refreshing every 5 seconds | Last updated: {{ datetime.now().strftime('%Y-%m-%d %H:%M:%S') }}
            </div>
        </div>
    </body>
    </html>
    """
    
    return render_template_string(
        template,
        overview=metrics_data['overview'],
        system=metrics_data['system'],
        endpoints=metrics_data['endpoints'],
        datetime=datetime
    )


if __name__ == "__main__":
    app.run(host="0.0.0.0", port=8080, debug=False)
