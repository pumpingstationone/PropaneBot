package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

type WebServer struct {
	Port      int
	Datastore *Datastore
	server    *http.Server
}

func (ws *WebServer) Run(ctx context.Context) func() error {
	return func() error {
		mux := http.NewServeMux()

		// Endpoint that returns the same string as Discord bot
		mux.HandleFunc("/propane", ws.handlePropaneText)

		// JSON API endpoint for structured data
		mux.HandleFunc("/api/propane", ws.handlePropaneJSON)

		// Serve static files for the web page
		mux.HandleFunc("/", ws.handleIndex)

		ws.server = &http.Server{
			Addr:    fmt.Sprintf(":%d", ws.Port),
			Handler: mux,
		}

		// Start server in a goroutine
		go func() {
			log.Printf("Web server starting on port %d", ws.Port)
			if err := ws.server.ListenAndServe(); err != http.ErrServerClosed {
				log.Printf("Web server error: %v", err)
			}
		}()

		// Wait for context to be done, then shutdown
		<-ctx.Done()
		log.Printf("Web server received shutdown signal. Shutting down...")

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		return ws.server.Shutdown(shutdownCtx)
	}
}

func (ws *WebServer) handlePropaneText(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	fmt.Fprint(w, ws.Datastore.GetString())
}

func (ws *WebServer) handlePropaneJSON(w http.ResponseWriter, r *http.Request) {
	data := ws.Datastore.Get()

	response := struct {
		Weight    float64   `json:"weight"`
		TimeStamp time.Time `json:"timestamp"`
		Remaining float64   `json:"remaining"`
		Message   string    `json:"message"`
	}{
		Weight:    data.Weight,
		TimeStamp: data.TimeStamp,
		Remaining: data.Remaining,
		Message:   ws.Datastore.GetString(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode JSON", http.StatusInternalServerError)
		return
	}
}

func (ws *WebServer) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	html := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>PropaneBot - Tank Level Monitor</title>
    <style>
        * {
            box-sizing: border-box;
        }
        body {
            font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
            margin: 0;
            padding: 0;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            min-height: 100vh;
            display: flex;
            align-items: center;
            justify-content: center;
        }
        .container {
            background-color: white;
            border-radius: 15px;
            box-shadow: 0 10px 30px rgba(0,0,0,0.2);
            width: 95vw;
            max-width: 800px;
            min-height: 80vh;
            display: flex;
            flex-direction: column;
            overflow: hidden;
        }
        .header {
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
            padding: 2rem;
            text-align: center;
        }
        .header h1 {
            margin: 0 0 0.5rem 0;
            font-size: clamp(1.5rem, 4vw, 2.5rem);
        }
        .header p {
            margin: 0;
            opacity: 0.9;
            font-size: clamp(0.9rem, 2vw, 1.1rem);
        }
        .content {
            flex: 1;
            padding: 2rem;
            display: flex;
            flex-direction: column;
        }
        .status {
            background-color: #e8f4fd;
            padding: 1rem 1.5rem;
            border-radius: 10px;
            border-left: 4px solid #2196F3;
            margin-bottom: 2rem;
            font-size: clamp(0.9rem, 2vw, 1rem);
        }
        .status.error {
            background-color: #ffebee;
            border-left-color: #f44336;
        }
        .data-display {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
            gap: 1.5rem;
            margin: 2rem 0;
        }
        .data-item {
            text-align: center;
            padding: 1.5rem;
            background: linear-gradient(145deg, #f8f9fa, #e9ecef);
            border-radius: 15px;
            box-shadow: 0 4px 15px rgba(0,0,0,0.1);
            transition: transform 0.2s ease;
        }
        .data-item:hover {
            transform: translateY(-2px);
        }
        .data-label {
            font-size: clamp(0.8rem, 1.5vw, 0.9rem);
            color: #666;
            margin-bottom: 0.5rem;
            text-transform: uppercase;
            letter-spacing: 0.5px;
        }
        .data-value {
            font-size: clamp(1.5rem, 4vw, 2.5rem);
            font-weight: bold;
            color: #333;
            margin-bottom: 0.25rem;
        }
        .data-unit {
            font-size: clamp(0.7rem, 1.2vw, 0.8rem);
            color: #666;
        }
        .progress-section {
            margin: 2rem 0;
        }
        .progress-bar {
            width: 100%;
            height: clamp(40px, 6vw, 50px);
            background-color: #e0e0e0;
            border-radius: 25px;
            overflow: hidden;
            box-shadow: inset 0 2px 4px rgba(0,0,0,0.1);
        }
        .progress-fill {
            height: 100%;
            background: linear-gradient(45deg, #4CAF50, #8BC34A);
            transition: width 0.5s ease;
            display: flex;
            align-items: center;
            justify-content: center;
            color: white;
            font-weight: bold;
            font-size: clamp(0.9rem, 2vw, 1.1rem);
        }
        .refresh-info {
            text-align: center;
            color: #666;
            font-size: clamp(0.8rem, 1.5vw, 0.9rem);
            margin-top: auto;
            padding-top: 2rem;
        }
        
        /* Mobile-specific adjustments */
        @media (max-width: 768px) {
            body {
                background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
                align-items: flex-start;
                padding: 1rem;
            }
            .container {
                width: 100%;
                min-height: calc(100vh - 2rem);
                margin: 0;
            }
            .header {
                padding: 1.5rem 1rem;
            }
            .content {
                padding: 1.5rem;
            }
            .data-display {
                grid-template-columns: 1fr;
                gap: 1rem;
            }
            .data-item {
                padding: 1rem;
            }
        }
        
        /* Tablet adjustments */
        @media (min-width: 769px) and (max-width: 1024px) {
            .container {
                width: 90vw;
                max-width: 700px;
            }
            .data-display {
                grid-template-columns: repeat(2, 1fr);
            }
        }
        
        /* Large screen adjustments */
        @media (min-width: 1200px) {
            .container {
                max-width: 900px;
            }
            .header {
                padding: 3rem;
            }
            .content {
                padding: 3rem;
            }
        }
        
        /* Landscape phone adjustments */
        @media (max-height: 500px) and (orientation: landscape) {
            body {
                align-items: flex-start;
                padding: 0.5rem;
            }
            .container {
                min-height: calc(100vh - 1rem);
            }
            .header {
                padding: 1rem;
            }
            .content {
                padding: 1rem;
            }
            .data-display {
                grid-template-columns: repeat(3, 1fr);
                margin: 1rem 0;
            }
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🔥 PropaneBot Tank Monitor</h1>
            <p>Real-time propane tank level monitoring</p>
        </div>
        
        <div class="content">
            <div id="status" class="status">
                <div id="message">Loading propane data...</div>
            </div>
            
            <div class="data-display">
                <div class="data-item">
                    <div class="data-label">Current Weight</div>
                    <div id="weight" class="data-value">--</div>
                    <div class="data-unit">lbs</div>
                </div>
                <div class="data-item">
                    <div class="data-label">Remaining</div>
                    <div id="remaining" class="data-value">--</div>
                    <div class="data-unit">%</div>
                </div>
                <div class="data-item">
                    <div class="data-label">Last Updated</div>
                    <div id="timestamp" class="data-value">--</div>
                    <div class="data-unit"></div>
                </div>
            </div>
            
            <div class="progress-section">
                <div class="progress-bar">
                    <div id="progress-fill" class="progress-fill" style="width: 0%;">
                        0%
                    </div>
                </div>
            </div>
            
            <div class="refresh-info">
                Data refreshes automatically every 5 seconds
            </div>
        </div>
    </div>

    <script>
        let updateInterval;
        
        async function fetchPropaneData() {
            try {
                const response = await fetch('/api/propane');
                if (!response.ok) {
                    throw new Error('Network response was not ok');
                }
                const data = await response.json();
                updateDisplay(data);
                updateStatus(data.message, false);
            } catch (error) {
                console.error('Error fetching propane data:', error);
                updateStatus('Error: Unable to fetch propane data', true);
            }
        }
        
        function updateDisplay(data) {
            // Update individual data points
            document.getElementById('weight').textContent = Math.round(data.weight);
            document.getElementById('remaining').textContent = Math.round(data.remaining);
            
            // Format timestamp
            const date = new Date(data.timestamp);
            const timeStr = date.toLocaleString();
            document.getElementById('timestamp').textContent = timeStr;
            
            // Update progress bar
            const progressFill = document.getElementById('progress-fill');
            const percentage = Math.round(data.remaining);
            progressFill.style.width = percentage + '%';
            progressFill.textContent = percentage + '%';
            
            // Change progress bar color based on level
            if (percentage > 50) {
                progressFill.style.background = 'linear-gradient(45deg, #4CAF50, #8BC34A)';
            } else if (percentage > 25) {
                progressFill.style.background = 'linear-gradient(45deg, #FF9800, #FFC107)';
            } else {
                progressFill.style.background = 'linear-gradient(45deg, #f44336, #FF5722)';
            }
        }
        
        function updateStatus(message, isError) {
            const statusElement = document.getElementById('status');
            const messageElement = document.getElementById('message');
            
            messageElement.textContent = message;
            
            if (isError) {
                statusElement.className = 'status error';
            } else {
                statusElement.className = 'status';
            }
        }
        
        // Start fetching data immediately and then every 5 seconds
        fetchPropaneData();
        updateInterval = setInterval(fetchPropaneData, 5000);
        
        // Clean up interval when page is unloaded
        window.addEventListener('beforeunload', function() {
            if (updateInterval) {
                clearInterval(updateInterval);
            }
        });
    </script>
</body>
</html>`

	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, html)
}
