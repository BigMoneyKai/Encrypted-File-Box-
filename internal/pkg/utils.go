package pkg

import (
	"net/http"
	"strconv"
	
	"golang.org/x/crypto/bcrypt"
	"github.com/gin-gonic/gin"
)

func HashPassword(pwd string) (string, error) {
	hashedpwd, err := bcrypt.GenerateFromPassword([]byte(pwd), 11) //也可以设置成12
	//cost:
	// 2024年默认值：10-12
	// Web应用：11（~400ms）
	// 生产环境：测试后确定，通常在 11-13 之间
	// 定期评估：每2-3年重新评估一次 cost 值

	if err != nil {
		panic(err)
	}

	return string(hashedpwd), nil

}

// CheckPassword
func CheckPassword(hash, plain string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain))
}

// JSON Success
func JSONOK(context *gin.Context, data interface{}) {
	context.JSON(http.StatusOK, gin.H{
			"code":    0,
			"message": "成功",
			"data":    data,
		})
}

// JSON error
func JSONError(context *gin.Context, code int, message string) {
	status := http.StatusBadRequest
	if code >= 100 && code <= 599 {
		status = code
	}
	context.JSON(status, gin.H{
		"code":    code,
		"message": message,
	})
}

func GetPageParams(context *gin.Context) (int, int) {
	pageStr := context.DefaultQuery("page", "1") //获取页数
	sizeStr := context.DefaultQuery("size", "20") //获取大小

	//将string转换成int
	page, _ := strconv.Atoi(pageStr)
	size, _ := strconv.Atoi(sizeStr)

	// 逻辑
	if page < 1 {
		page = 1
	}

	if size < 1 || size > 100 {
		size = 20 //防止过大查询
	}

	return page, size
}
