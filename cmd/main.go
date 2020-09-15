package main

import (
	"fmt"

	fileRepo "github.com/freemen-app/file_storage/adapter/repository/file"
	"github.com/freemen-app/file_storage/domain/dto"
	awsSession "github.com/freemen-app/file_storage/infrastructure/store/aws"
	fileUseCase "github.com/freemen-app/file_storage/usecase/file"
)

func main() {
	sess := awsSession.New()
	repo := fileRepo.New(sess, "o4eredko.photos")
	useCase := fileUseCase.New(repo)
	err := useCase.BatchDelete(&dto.BatchDeleteInput{
		"https://s3.eu-central-1.amazonaws.com/o4eredko.photos/duump.tsv",
		"https://s3.eu-central-1.amazonaws.com/o4eredko.photos/test2.yml",
	})
	fmt.Println(err)
	// s3Store := aws.New()
	// repo := fileRepo.New(s3Store, "o4eredko.photos")
	// useCase := fileUseCase.New(repo)
	// fmt.Println(useCase.Delete(&dto.DeleteInput{Url: "https://aws.eu-central-1.amazonaws.com/o4eredko.photos/test1.yml"}))

	// c := config.NewConfig("config.yml")
	//
	// app := app.New(c)
	//
	// api := api.New(c, app)
	// // Start server
	// go api.Start()
	// // Wait for interrupt signal to gracefully shutdown the server with
	// // api timeout of 10 seconds.
	// quit := make(chan os.Signal)
	// signal.Notify(quit, os.Interrupt)
	// <-quit
	// ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	// defer cancel()
	// api.Shutdown(ctx)
}
