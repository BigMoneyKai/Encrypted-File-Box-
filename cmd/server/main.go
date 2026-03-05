package main

import (
	"fmt"
	"log"
	"path/filepath"
	"runtime"
	"strconv"

	"github.com/Kaikai20040827/graduation/internal/config"
	"github.com/Kaikai20040827/graduation/internal/handler"
	"github.com/Kaikai20040827/graduation/internal/logo"
	"github.com/Kaikai20040827/graduation/internal/pkg"
	"github.com/Kaikai20040827/graduation/internal/routes"
	"github.com/Kaikai20040827/graduation/internal/service"
)

const (
	Debug = true
)

func main() {
	logo.DrawLogo()
	fmt.Println("-----Secure File Box-----")
	fmt.Println("")
	//counters := config.NewFuncCounters()

	// 1. 加载配置
	fmt.Println("-----Starting loading configuration-----")
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
	// fmt.Printf("(%d/5) done", )
	// fmt.Println("")
	fmt.Println("-----Loaded successfully-----")
	fmt.Println("")

	// 2. Logger
	fmt.Println("-----Starting initializing logger-----")
	pkg.InitLogger(Debug)
	// fmt.Printf("(%d/1) done", )
	// fmt.Println("")
	fmt.Println("-----Initialized logger successfully-----")
	fmt.Println("")

	// 3. DB(mysql)
	fmt.Println("-----Starting initializing database-----")
	db, err := pkg.NewDatabase(&cfg.Database)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	fmt.Println("-----Initialized database successfully-----")
	fmt.Println("")

	// 4. Services
	fmt.Println("-----Starting initializing service(UserService, FileService)-----")
	userSrv := service.NewUserService(db)
	_, currentFile, _, _ := runtime.Caller(0)
	projectRoot := filepath.Dir(filepath.Dir(filepath.Dir(currentFile)))
	storagePath := filepath.Join(projectRoot, "storage")
	fileSrv := service.NewFileService(db, storagePath, cfg.FileCrypto.Key)
	// fmt.Printf("(%d/2) done", )
	// fmt.Println("")
	fmt.Println("-----Initialized UserService and FileService successfully-----")
	fmt.Println("")

	// 5. Handlers
	fmt.Println("-----Starting initializing handlers(UserService, FileService)-----")
	authH := handler.NewAuthHandler(userSrv, &cfg.JWT)
	userH := handler.NewUserHandler(userSrv, fileSrv)
	fileH := handler.NewFileHandler(fileSrv)
	// fmt.Printf("(%d/3) done", )
	// fmt.Println("")
	fmt.Println("-----Initialized UserService and FileService successfully-----")
	fmt.Println("")

	// 6. Gin
	fmt.Println("-----Starting initializing Gin framework-----")
	r := routes.SetupRouter(&cfg.JWT)

	fmt.Println("-----Initialized Gin framework successfully-----")
	fmt.Println("")

	// 7. 注册 API 路由（最关键）
	fmt.Println("-----Starting initializing API-----")
	routes.RegisterAPIRoutes(r, authH, userH, fileH, &cfg.JWT)
	fmt.Println("-----Initialized API successfully-----")
	fmt.Println("")

	// 8. 启动
	port := strconv.Itoa(cfg.Server.Port)
	host := cfg.Server.Host
	addr := host + ":" + port
	log.Println("server running at", addr)

	if err := r.Run(addr); err != nil {
		log.Fatalf("Error: %v", err)
	}
}
