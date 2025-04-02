package controller

import (
	"blue-note/model"
	"blue-note/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

type FileController struct {
	fileService *service.FileService
}

func NewFileController(fileService *service.FileService) *FileController {
	return &FileController{fileService: fileService}
}

// DeleteFile 删除文件
func (c *FileController) DeleteFile(ctx *gin.Context) {
	var req model.DeleteFileRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code":    40003,
			"message": "请求参数错误",
			"error":   err.Error(),
		})
		return
	}
	
	userID := ctx.GetString("userId")
	
	err := c.fileService.DeleteFile(userID, req.FilePath)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code":    50001,
			"message": "删除文件失败",
			"error":   err.Error(),
		})
		return
	}
	
	ctx.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "删除文件成功",
	})
} 