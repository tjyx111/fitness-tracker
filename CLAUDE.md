# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

健身追踪器 (Fitness Tracker) - A web-based fitness tracking application with Go backend and vanilla JavaScript frontend. Uses CSV files for data persistence.

**Architecture:**
- **Backend**: Go HTTP server serving RESTful API on port 8080
- **Frontend**: Single-page application with vanilla JavaScript, HTML, CSS
- **Data Storage**: CSV files in `backend/data/` directory
- **No database**: All data stored as CSV with manual ID generation

## Common Commands

### Starting the Server
```bash
# From project root
./start.sh

# Or manually
cd backend
export PATH=/usr/local/go/bin:$PATH
go run .
```

### Restarting with API Tests
```bash
./restart_server.sh  # Stops existing server, starts new one, runs basic API tests
```

### Testing APIs
```bash
./test_api.sh  # Tests exercises, groups, progress endpoints

# Manual API testing
curl http://localhost:8080/api/exercises
curl http://localhost:8080/api/stats/volume?days=30
curl http://localhost:8080/api/stats/personal-records
```

### Go Environment
The project requires Go to be in `/usr/local/go/bin`. Set PATH before running:
```bash
export PATH=/usr/local/go/bin:$PATH
```

## Architecture

### Backend Structure (`backend/`)

**Core Files:**
- `main.go` - HTTP server setup, route registration, serves frontend static files
- `models.go` - Data structures (Exercise, ExerciseGroup, TrainingSession, TrainingRecord, etc.)
- `csv_handler.go` - All CSV file I/O operations with getNextID() for auto-increment
- `handlers.go` - REST API handlers for exercises, groups, sessions
- `progress.go` - ProgressAnalyzer for training progress calculations
- `stats.go` - Statistics calculation models (VolumeStats, IntensityStats, etc.)
- `stats_handlers.go` - Statistics API endpoints
- `weight_handler.go` - Weight tracking CRUD operations

**Data Models:**
- `Exercise` - Individual exercises with muscle group and unit type (kg/duration)
- `ExerciseGroup` - Training plans containing multiple exercise IDs
- `TrainingSession` - Daily training sessions linked to exercise groups
- `TrainingRecord` - Individual set data (weight, reps, duration) per exercise

**API Route Pattern:**
- `/api/exercises` - GET/POST/PUT/DELETE for exercise CRUD
- `/api/groups` - GET/POST for exercise groups
- `/api/groups/{id}/last-record` - GET previous training for reference
- `/api/session` - POST to submit training session
- `/api/progress/exercise/{id}` - GET progress for specific exercise
- `/api/stats/*` - Various statistics endpoints
- `/api/weight` - GET/POST weight records

### Frontend Structure (`frontend/`)

**Core Files:**
- `index.html` - Single-page app with navigation (Training/Exercises/Statistics)
- `app.js` - Main application logic for training logging
- `stats.js` - Statistics page rendering and charts
- `styles.css` - All styling
- Test HTML files for debugging specific features

**Frontend Architecture:**
- Single-page application with tab-based navigation
- Three main pages: 今日训练 / 动作管理 / 统计分析
- State management via global `state` object
- API calls to `http://localhost:8080/api`

### CSV Data Storage

**Data Files (`backend/data/`):**
- `exercises.csv` - Exercise library (id, name, muscleGroup, unit)
- `exercise_groups.csv` - Training plans (id, name, exerciseIds comma-separated)
- `training_sessions.csv` - Training sessions (sessionId, groupId, date, status)
- `training_records.csv` - Individual sets (recordId, sessionId, exerciseId, setNumber, weight, reps, duration, note)
- `weight_records.csv` - Weight tracking

**CSV Handler Behavior:**
- All CSV files have Chinese headers in first row
- `LoadExercises()` etc. skip header rows
- `getNextID()` scans entire CSV for max ID + 1
- No transaction support - concurrent writes can cause data loss

## Key Implementation Details

### Exercise Unit Types
Exercises have `unit` field: "kg" for weight-based exercises, "duration" for time-based exercises. Affects UI rendering (weight inputs vs duration inputs).

### Statistics System
Comprehensive stats system in `stats.go`:
- Volume load calculation (weight × reps)
- 1RM estimation using Epley formula
- Training frequency analysis
- Personal record tracking
- Muscle balance analysis
- Comprehensive scoring (quality score, fatigue index, etc.)

### Progress Tracking
`ProgressAnalyzer` in `progress.go`:
- Analyzes historical training data
- Provides progress trends by exercise or muscle group
- Used for displaying progress charts and last session references

### ID Generation
No auto-increment at database level:
- CSV handler uses `getNextID()` to scan entire file for max ID
- New records get `maxID + 1`
- Not thread-safe - concurrent writes can duplicate IDs

## Development Notes

- When modifying API endpoints, update both `main.go` routing and corresponding handler
- CSV files are manually editable - format must match exactly (comma-separated, Chinese headers)
- Frontend uses vanilla JS with no build step - edit and refresh browser
- Statistics calculations can be CPU-intensive for large datasets
- Server runs on port 8080 - ensure it's not already in use before starting
