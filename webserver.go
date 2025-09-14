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
        body {
            font-family: Arial, sans-serif;
            max-width: 800px;
            margin: 0 auto;
            padding: 20px;
            background-color: #f5f5f5;
        }
        .container {
            background-color: white;
            padding: 30px;
            border-radius: 10px;
            box-shadow: 0 2px 10px rgba(0,0,0,0.1);
        }
        .header {
            text-align: center;
            color: #333;
            margin-bottom: 30px;
        }
        .status {
            background-color: #e8f4fd;
            padding: 20px;
            border-radius: 8px;
            border-left: 4px solid #2196F3;
            margin-bottom: 20px;
        }
        .status.error {
            background-color: #ffebee;
            border-left-color: #f44336;
        }
        .data-display {
            display: grid;
            grid-template-columns: 1fr 1fr 1fr;
            gap: 20px;
            margin: 20px 0;
        }
        .data-item {
            text-align: center;
            padding: 15px;
            background-color: #f8f9fa;
            border-radius: 8px;
        }
        .data-label {
            font-size: 14px;
            color: #666;
            margin-bottom: 5px;
        }
        .data-value {
            font-size: 24px;
            font-weight: bold;
            color: #333;
        }
        .refresh-info {
            text-align: center;
            color: #666;
            font-size: 14px;
            margin-top: 20px;
        }
        .progress-bar {
            width: 100%;
            height: 30px;
            background-color: #e0e0e0;
            border-radius: 15px;
            overflow: hidden;
            margin: 10px 0;
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
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🔥 PropaneBot Tank Monitor</h1>
            <p>Real-time propane tank level monitoring</p>
        </div>
        
        <div id="status" class="status">
            <div id="message">Loading propane data...</div>
        </div>
        
        <div class="data-display">
            <div class="data-item">
                <div class="data-label">Current Weight</div>
                <div id="weight" class="data-value">--</div>
                <div style="font-size: 12px; color: #666;">lbs</div>
            </div>
            <div class="data-item">
                <div class="data-label">Remaining</div>
                <div id="remaining" class="data-value">--</div>
                <div style="font-size: 12px; color: #666;">%</div>
            </div>
            <div class="data-item">
                <div class="data-label">Last Updated</div>
                <div id="timestamp" class="data-value">--</div>
            </div>
        </div>
        
        <div class="progress-bar">
            <div id="progress-fill" class="progress-fill" style="width: 0%;">
                0%
            </div>
        </div>
        
        <div class="refresh-info">
            Data refreshes automatically every 5 seconds
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
