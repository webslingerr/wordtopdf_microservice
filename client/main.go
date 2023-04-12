package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "app/docpdf"
)

func main() {
	conn, err := grpc.Dial("localhost:9001",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		panic(err)
	}

	cc := pb.NewConvertServiceClient(conn)

	r := gin.Default()

	r.POST("/convert", func(ctx *gin.Context) {
		file, err := ctx.FormFile("file")
		if err != nil {
			log.Fatal(err)
		}

		ext := filepath.Ext(file.Filename)
		if ext != ".docx" {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid file extension. Only .docx files are allowed."})
			return
		}

		tempDir, err := os.MkdirTemp("", "docx-to-pdf-*")
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer os.RemoveAll(tempDir)

		filePath := filepath.Join(tempDir, file.Filename)
		if err := ctx.SaveUploadedFile(file, filePath); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		pdfPath, err := cc.ConvertToPdf(
			context.Background(),
			&pb.ConvertRequest{
				Path: filePath,
			},
		)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		ctx.Header("Content-Type", "application/pdf")
		ctx.File(pdfPath.Path)
	})

	if err := r.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}
