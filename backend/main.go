package main

import (
	"embed"
	"fmt"
	"io/fs"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
)

//go:embed frontend/*
var embeddedFrontend embed.FS

func main() {
	// 获取当前目录
	dataDir := os.Getenv("DATA_DIR")
	if dataDir == "" {
		dataDir = "./data"
	}
	absPath, err := filepath.Abs(dataDir)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Data directory: %s\n", absPath)

	// 创建服务器
	server, err := NewServer(dataDir)
	if err != nil {
		log.Fatalf("initialize SQLite storage: %v", err)
	}
	defer server.csv.Close()

	// 设置路由
	http.HandleFunc("/api/exercises", server.handleExercises)
	http.HandleFunc("/api/exercises/", server.handleExercises) // 支持DELETE /api/exercises/{id}
	http.HandleFunc("/api/groups", server.handleGroups)
	http.HandleFunc("/api/groups/", func(w http.ResponseWriter, r *http.Request) {
		// 检查是否是 /api/groups/{id}/last-record
		if len(r.URL.Path) > len("/api/groups/") && filepath.Base(r.URL.Path) == "last-record" {
			server.handleGroupLastRecord(w, r)
			return
		}
		// 其他group相关的请求
		server.handleGroups(w, r)
	})
	http.HandleFunc("/api/session/day-records", server.handleSessionDayRecords)
	http.HandleFunc("/api/session/records", server.handleSessionRecords)
	http.HandleFunc("/api/session", server.handleSessionSubmit)
	http.HandleFunc("/api/progress/exercise/", server.handleExerciseProgress)
	http.HandleFunc("/api/progress/muscle/", server.handleMuscleProgress)

	// 统计API路由
	http.HandleFunc("/api/stats/volume", server.handleStatsVolume)
	http.HandleFunc("/api/stats/intensity", server.handleStatsIntensity)
	http.HandleFunc("/api/stats/progress-rate/", server.handleStatsProgressRate)
	http.HandleFunc("/api/stats/personal-records", server.handleStatsPersonalRecords)
	http.HandleFunc("/api/stats/frequency", server.handleStatsFrequency)
	http.HandleFunc("/api/stats/comprehensive", server.handleStatsComprehensive)
	http.HandleFunc("/api/stats/report", server.handleStatsReport)
	http.HandleFunc("/api/stats/calendar", server.handleStatsCalendar)
	http.HandleFunc("/api/stats/exercise/", server.handleStatsExerciseDetail)
	http.HandleFunc("/api/stats/overview-history", server.handleStatsOverviewHistory)
	http.HandleFunc("/api/stats/filtered", server.handleStatsFiltered)
	http.HandleFunc("/api/stats/day-records", server.handleStatsDayRecords)

	// 体重记录API路由
	http.HandleFunc("/api/weight", server.handleWeightRecords)
	http.HandleFunc("/api/weight/latest", server.handleWeightLatest)

	// 笔记 API 路由
	http.HandleFunc("/api/note-tags", server.handleNoteTags)
	http.HandleFunc("/api/note-tags/", server.handleNoteTagActions)
	http.HandleFunc("/api/notes", server.handleNotes)
	http.HandleFunc("/api/notes/new", server.handleNewNote)
	http.HandleFunc("/api/notes/history", server.handleNoteHistory)
	http.HandleFunc("/api/todos", server.handleTodos)
	http.HandleFunc("/api/todos/", server.handleTodoItem)
	http.HandleFunc("/api/challenges", server.handleChallenges)
	http.HandleFunc("/api/challenges/", server.handleChallenge)
	http.HandleFunc("/api/challenge-daily-items/", server.handleChallengeDailyItem)
	http.HandleFunc("/api/stats/challenges", server.handleChallengeStats)

	// 数据库同步
	http.HandleFunc("/api/sync/database", server.handleDatabaseSync)

	// 静态文件服务（前端）
	http.Handle("/", http.FileServer(frontendFileSystem()))

	// 启动服务器
	listenAddr := resolveListenAddr()
	fmt.Printf("Server starting on %s\n", listenAddr)
	fmt.Println("Available endpoints:")
	fmt.Println("  GET    /api/exercises          - 获取所有动作")
	fmt.Println("  POST   /api/exercises          - 添加新动作")
	fmt.Println("  GET    /api/groups             - 获取所有动作组")
	fmt.Println("  POST   /api/groups             - 添加新动作组")
	fmt.Println("  GET    /api/groups/{id}/last-record - 获取某动作组上次训练记录")
	fmt.Println("  POST   /api/session            - 提交训练会话")
	fmt.Println("  GET    /api/progress/exercise/{id} - 获取某动作进度")
	fmt.Println("  GET    /api/progress/muscle/{name} - 获取某肌肉群进度")
	fmt.Println("")
	fmt.Println("  Statistics API:")
	fmt.Println("  GET    /api/stats/volume?days=30 - 容量统计")
	fmt.Println("  GET    /api/stats/intensity?days=30 - 强度统计")
	fmt.Println("  GET    /api/stats/progress-rate/{id}?target=100 - 进度速度")
	fmt.Println("  GET    /api/stats/personal-records - 个人记录")
	fmt.Println("  GET    /api/stats/frequency - 训练频率")
	fmt.Println("  GET    /api/stats/comprehensive?days=30 - 综合统计")
	fmt.Println("  GET    /api/stats/report?days=30 - 完整报告")
	fmt.Println("")
	fmt.Println("  Weight Records API:")
	fmt.Println("  GET    /api/weight - 获取体重记录")
	fmt.Println("  POST   /api/weight - 添加体重记录")
	fmt.Println("  GET    /api/weight/latest - 获取最新体重")
	fmt.Println("  GET    /api/note-tags - 获取笔记标签")
	fmt.Println("  POST   /api/note-tags - 添加笔记标签")
	fmt.Println("  GET    /api/notes?tagId=1 - 获取标签笔记")
	fmt.Println("  PUT    /api/notes - 保存标签笔记")
	fmt.Println("  GET    /api/notes/history?tagId=1 - 获取标签历史笔记")
	fmt.Println("  GET    /api/todos - 获取待办事项")
	fmt.Println("  POST   /api/todos - 添加待办事项")
	fmt.Println("  GET    /api/sync/database - 下载 SQLite 快照")

	if err := http.ListenAndServe(listenAddr, nil); err != nil {
		log.Fatal(err)
	}
}

func frontendFileSystem() http.FileSystem {
	frontendFS, err := fs.Sub(embeddedFrontend, "frontend")
	if err != nil {
		log.Fatal(err)
	}
	return http.FS(frontendFS)
}

func resolveListenAddr() string {
	if addr := os.Getenv("LISTEN_ADDR"); addr != "" {
		return addr
	}

	host := os.Getenv("BIND_HOST")
	if host == "" {
		host = os.Getenv("HOST")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	if host == "" {
		return ":" + port
	}
	return net.JoinHostPort(host, port)
}
