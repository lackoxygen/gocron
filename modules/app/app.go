package app

import (
    "os"
    "runtime"

    "github.com/ouqiang/cron-scheduler/models"
    "github.com/ouqiang/cron-scheduler/modules/ansible"
    "github.com/ouqiang/cron-scheduler/modules/crontask"
    "github.com/ouqiang/cron-scheduler/modules/utils"
    "github.com/ouqiang/cron-scheduler/service"
    "github.com/ouqiang/cron-scheduler/modules/setting"
    "github.com/ouqiang/cron-scheduler/modules/logger"
)

var (
    AppDir       string // 应用根目录
    ConfDir      string // 配置目录
    LogDir       string // 日志目录
    DataDir      string // 数据目录，存放session文件等
    AppConfig    string // 应用配置文件
    AnsibleHosts string // ansible hosts文件
    Installed    bool   // 应用是否安装过
)

func InitEnv() {
    logger.InitLogger()
    CheckEnv()
    wd, err := os.Getwd()
    if err != nil {
        panic(err)
    }
    AppDir = wd
    ConfDir = AppDir + "/conf"
    LogDir = AppDir + "/log"
    DataDir = AppDir + "/data"
    AppConfig = ConfDir + "/app.ini"
    AnsibleHosts = ConfDir + "/ansible_hosts.ini"
    checkDirExists(ConfDir, LogDir, DataDir)
    // ansible配置文件目录
    os.Setenv("ANSIBLE_CONFIG", ConfDir)
    Installed = IsInstalled()
    if Installed {
        InitDb()
        InitResource()
    }
}

// 判断应用是否安装过
func IsInstalled() bool {
    _, err := os.Stat(ConfDir + "/install.lock")
    if os.IsNotExist(err) {
        return false
    }

    return true
}

// 检测环境
func CheckEnv() {
    // ansible不支持安装在windows上, windows只能作为被控机
    if runtime.GOOS == "windows" {
        logger.Fatal("不支持在windows上运行")
    }
    _, err := utils.ExecShell("ansible", "--version")
    if err != nil {
        logger.Fatal(err)
    }
}

// 创建安装锁文件
func CreateInstallLock() error {
    _, err := os.Create(ConfDir + "/install.lock")
    if err != nil {
        logger.Error("创建安装锁文件失败")
    }

    return err
}

// 初始化资源
func InitResource() {
    // 初始化ansible Hosts
    ansible.DefaultHosts = ansible.NewHosts(AnsibleHosts)
    // 初始化定时任务
    crontask.DefaultCronTask = crontask.NewCronTask()
    serviceTask := new(service.Task)
    serviceTask.Initialize()
}

// 初始化DB
func InitDb()  {
    dbConfig := getDbConfig(AppConfig)
    models.Db = models.CreateDb(dbConfig)
}

// 检测目录是否存在
func checkDirExists(path ...string) {
    for _, value := range path {
        _, err := os.Stat(value)
        if os.IsNotExist(err) {
            logger.Fatal(value + "目录不存在")
        }
        if os.IsPermission(err) {
            logger.Fatal(value + "目录无权限操作")
        }
    }
}

// 获取数据库配置
func getDbConfig(configFile string) map[string]string {
    config, err := setting.Read(configFile)
    if err != nil {
        logger.Fatal(err)
    }
    section := config.Section("db")
    if err != nil {
        logger.Fatal(err)
    }
    var db map[string]string = make(map[string]string)
    db["user"] = section.Key("user").String()
    db["password"] = section.Key("password").String()
    db["host"] = section.Key("host").String()
    db["port"] = section.Key("port").String()
    db["database"] = section.Key("database").String()
    db["charset"] = section.Key("charset").String()
    db["prefix"] = section.Key("prefix").String()
    db["engine"] = section.Key("engine").String()

    return db
}